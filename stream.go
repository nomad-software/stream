// Package stream provides generic, lazy, single pass stream processors.
//
// A stream hands out each of its values once, like a channel. Taking a value
// from a stream moves it on, so whatever consumes the stream next carries on
// from there, and a stream that has been exhausted stays empty.
//
// Nothing happens until a value is asked for, by ranging over All or by calling
// a sink such as Slice or Pop, and then each stage runs on the caller's
// goroutine, one value at a time. MapParallel, Merge and Buffer are the
// exceptions: they run the stages before them in other goroutines, and are the
// way to opt in to concurrency. Those goroutines run until their streams are
// exhausted or the context is cancelled.
//
// A stream must not be used by more than one goroutine at a time.
package stream

import (
	"context"
	"iter"
	"sync"
	"time"
)

// Stream is a stream of values bound to a context. Cancelling the context stops
// the stream. Streams should be created using a generator. The zero value is an
// empty stream.
type Stream[T comparable] struct {
	// done is the done channel of the stream's context.
	done <-chan struct{}

	// next returns the stream's next value. It returns false if the values are
	// exhausted or the passed done channel is closed. The done channel is
	// passed in, rather than taken from the stream, so that a stream can be
	// consumed on behalf of another stream's context. Any state is held by the
	// function, so copies of a stream share their position.
	next func(done <-chan struct{}) (T, bool)
}

// Enum adds an enumeration index to a value.
type Enum[T comparable] struct {
	Index int `json:"index"`
	Val   T   `json:"val"`
}

// newStream creates a stream bound to the passed context.
func newStream[T comparable](ctx context.Context, next func(done <-chan struct{}) (T, bool)) Stream[T] {
	return Stream[T]{ctx.Done(), next}
}

// source returns the function that returns the stream's next value, which is
// never nil.
func (c Stream[T]) source() func(done <-chan struct{}) (T, bool) {
	if c.next == nil {
		return func(<-chan struct{}) (T, bool) {
			var zero T
			return zero, false
		}
	}
	return c.next
}

// cancelled returns true if the passed done channel is closed. A nil channel
// is never cancelled.
func cancelled(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

// concurrently returns a function that returns the values emitted by a set of
// producers, in the order they arrive, through a channel with a buffer of the
// passed size. The producers are created, and each started in its own
// goroutine, the first time a value is asked for. A producer must stop when
// emit returns false, which it does when the context is cancelled.
func concurrently[T any](size int, producers func(done <-chan struct{}) []func(emit func(T) bool)) func(done <-chan struct{}) (T, bool) {
	var values chan T

	start := func(done <-chan struct{}) {
		values = make(chan T, size)

		emit := func(val T) bool {
			select {
			case values <- val:
				return true
			case <-done:
				return false
			}
		}

		wg := new(sync.WaitGroup)
		for _, produce := range producers(done) {
			wg.Go(func() { produce(emit) })
		}
		go func() {
			wg.Wait()
			close(values)
		}()
	}

	return func(done <-chan struct{}) (T, bool) {
		if values == nil {
			start(done)
		}

		// Values that are already buffered can still be received after the
		// context is cancelled.
		if size > 0 {
			select {
			case val, ok := <-values:
				return val, ok
			default:
			}
		}

		// The context is checked first because select chooses randomly when
		// both a value and the cancellation are ready.
		if !cancelled(done) {
			select {
			case val, ok := <-values:
				return val, ok
			case <-done:
			}
		}

		var zero T
		return zero, false
	}
}

// All returns an iterator over the stream's remaining values. Iteration stops
// when the values are exhausted or the stream's context is cancelled. If the
// iteration is stopped early, the stream carries on from the next value when
// it's next used.
func (c Stream[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		next := c.source()
		for {
			val, ok := next(c.done)
			if !ok || !yield(val) {
				return
			}
		}
	}
}

// Take returns the next n values of the stream.
func (c Stream[T]) Take(n int) Stream[T] {
	next := c.source()
	i := 0

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		if i >= n {
			var zero T
			return zero, false
		}
		val, ok := next(done)
		if ok {
			i++
		}
		return val, ok
	}}
}

// Until ends the stream when the passed function returns true, otherwise it
// will keep returning values. The passed function is called once for each value.
func (c Stream[T]) Until(f func(T) bool) Stream[T] {
	next := c.source()
	stopped := false

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		if !stopped {
			val, ok := next(done)
			if ok && !f(val) {
				return val, true
			}
			stopped = ok
		}
		var zero T
		return zero, false
	}}
}

// Map mutates stream values based on the passed function. The passed function
// is called once for each value.
func (c Stream[T]) Map[R comparable](f func(T) R) Stream[R] {
	next := c.source()

	return Stream[R]{c.done, func(done <-chan struct{}) (R, bool) {
		val, ok := next(done)
		if !ok {
			var zero R
			return zero, false
		}
		return f(val), true
	}}
}

// MapParallel mutates stream values based on the passed function, which is
// called concurrently by the passed number of workers. The passed function is
// called once for each value. The streamed values are not ordered. The stages
// before this one run in their own goroutine.
func (c Stream[T]) MapParallel[R comparable](workers int, f func(T) R) Stream[R] {
	if workers <= 0 {
		return Stream[R]{done: c.done}
	}

	next := c.source()

	return Stream[R]{c.done, concurrently(0, func(done <-chan struct{}) []func(func(R) bool) {
		work := make(chan T)

		go func() {
			defer close(work)
			for {
				val, ok := next(done)
				if !ok {
					return
				}
				select {
				case work <- val:
				case <-done:
					return
				}
			}
		}()

		worker := func(emit func(R) bool) {
			for val := range work {
				if !emit(f(val)) {
					return
				}
			}
		}

		producers := make([]func(func(R) bool), workers)
		for i := range producers {
			producers[i] = worker
		}
		return producers
	})}
}

// Filter filters stream values based on the passed function returning true.
// The passed function is called once for each value.
func (c Stream[T]) Filter(f func(T) bool) Stream[T] {
	next := c.source()

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		for {
			val, ok := next(done)
			if !ok || f(val) {
				return val, ok
			}
		}
	}}
}

// Reduce reduces the remaining stream values to one value based on the passed
// function. The passed function is called once for each value after the first.
// Nothing is returned if the stream is empty or its context is cancelled.
func (c Stream[T]) Reduce(f func(T, T) T) Stream[T] {
	next := c.source()
	finished := false

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		var zero T
		if finished {
			return zero, false
		}
		finished = true

		a, ok := next(done)
		if !ok {
			return zero, false
		}
		for {
			val, ok := next(done)
			if !ok {
				break
			}
			a = f(a, val)
		}
		if cancelled(done) {
			return zero, false
		}
		return a, true
	}}
}

// Last will return the final value of the stream. Nothing is returned if the
// stream is empty or its context is cancelled.
func (c Stream[T]) Last() Stream[T] {
	next := c.source()
	finished := false

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		var zero T
		if finished {
			return zero, false
		}
		finished = true

		last, ok := next(done)
		if !ok {
			return zero, false
		}
		for {
			val, ok := next(done)
			if !ok {
				break
			}
			last = val
		}
		if cancelled(done) {
			return zero, false
		}
		return last, true
	}}
}

// sources returns the functions that return the next values of the main stream
// and the passed streams, in order.
func (c Stream[T]) sources(b Stream[T], args []Stream[T]) []func(<-chan struct{}) (T, bool) {
	sources := make([]func(<-chan struct{}) (T, bool), 0, 2+len(args))
	sources = append(sources, c.source(), b.source())
	for _, arg := range args {
		sources = append(sources, arg.source())
	}
	return sources
}

// Chain will append values from the passed streams to the end of the main
// stream. The main stream's context is used for all streams.
func (c Stream[T]) Chain(b Stream[T], args ...Stream[T]) Stream[T] {
	inputs := c.sources(b, args)

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		for len(inputs) > 0 {
			val, ok := inputs[0](done)
			if ok {
				return val, true
			}
			if cancelled(done) {
				break
			}
			inputs = inputs[1:]
		}
		var zero T
		return zero, false
	}}
}

// Merge will return values from the main stream and the passed streams as they
// become available. Each stream runs in its own goroutine and the streamed
// values are not ordered. The main stream's context is used for all streams.
func (c Stream[T]) Merge(b Stream[T], args ...Stream[T]) Stream[T] {
	inputs := c.sources(b, args)

	return Stream[T]{c.done, concurrently(0, func(done <-chan struct{}) []func(func(T) bool) {
		producers := make([]func(func(T) bool), len(inputs))
		for i, next := range inputs {
			producers[i] = func(emit func(T) bool) {
				for {
					val, ok := next(done)
					if !ok || !emit(val) {
						return
					}
				}
			}
		}
		return producers
	})}
}

// RoundRobin will return alternate values from the main stream and the passed
// streams, in order, until they are all exhausted. The main stream's context
// is used for all streams.
func (c Stream[T]) RoundRobin(b Stream[T], args ...Stream[T]) Stream[T] {
	inputs := c.sources(b, args)
	i := 0

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		// Each input is tried once, starting with the one after the last one
		// that returned a value. A nil input is exhausted.
		for range inputs {
			next := inputs[i]
			current := i
			i++
			if i == len(inputs) {
				i = 0
			}
			if next == nil {
				continue
			}
			val, ok := next(done)
			if ok {
				return val, true
			}
			if cancelled(done) {
				break
			}
			inputs[current] = nil
		}
		var zero T
		return zero, false
	}}
}

// Chunk returns an iterator of streams of the passed length, filled with the
// values of the main stream. The last stream is shorter if the values run out.
func (c Stream[T]) Chunk(n int) iter.Seq[Stream[T]] {
	next := c.source()

	return func(yield func(Stream[T]) bool) {
		if n <= 0 {
			return
		}

		for {
			chunk := make([]T, 0, n)
			for len(chunk) < n {
				val, ok := next(c.done)
				if !ok {
					if len(chunk) > 0 && !cancelled(c.done) {
						yield(fromSlice(c.done, chunk))
					}
					return
				}
				chunk = append(chunk, val)
			}
			if !yield(fromSlice(c.done, chunk)) {
				return
			}
		}
	}
}

// Drop removes the next n values from the stream before continuing.
func (c Stream[T]) Drop(n int) Stream[T] {
	next := c.source()
	dropped := 0

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		for dropped < n {
			if val, ok := next(done); !ok {
				return val, false
			}
			dropped++
		}
		return next(done)
	}}
}

// Stride iterates over stream values returning every n value of the stream.
func (c Stream[T]) Stride(n int) Stream[T] {
	if n <= 0 {
		return Stream[T]{done: c.done}
	}

	next := c.source()
	started := false

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		if started {
			for range n - 1 {
				if val, ok := next(done); !ok {
					return val, false
				}
			}
		}
		started = true
		return next(done)
	}}
}

// Tail returns the last n values of the stream once it's exhausted.
func (c Stream[T]) Tail(n int) Stream[T] {
	if n <= 0 {
		return Stream[T]{done: c.done}
	}

	next := c.source()
	collected := false
	var tail []T

	// collect returns the last n values in order.
	collect := func(done <-chan struct{}) []T {
		// A ring buffer of the last n values, where i is the oldest value
		// once the buffer is full.
		ring := make([]T, 0, min(n, 1024))
		i := 0
		for {
			val, ok := next(done)
			if !ok {
				break
			}
			if len(ring) < n {
				ring = append(ring, val)
				continue
			}
			ring[i] = val
			i++
			if i == n {
				i = 0
			}
		}
		if cancelled(done) {
			return nil
		}
		return append(ring[i:len(ring):len(ring)], ring[:i]...)
	}

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		if !collected {
			collected = true
			tail = collect(done)
		}
		if len(tail) == 0 || cancelled(done) {
			var zero T
			return zero, false
		}
		val := tail[0]
		tail = tail[1:]
		return val, true
	}}
}

// Zip returns an iterator of streams containing the next values of the main
// stream and all other passed streams, in order. It stops when any of the
// streams is exhausted. The main stream's context is used for all streams.
func (c Stream[T]) Zip(b Stream[T], args ...Stream[T]) iter.Seq[Stream[T]] {
	inputs := c.sources(b, args)

	return func(yield func(Stream[T]) bool) {
		for {
			zip := make([]T, 0, len(inputs))
			for _, next := range inputs {
				val, ok := next(c.done)
				if !ok {
					return
				}
				zip = append(zip, val)
			}
			if !yield(fromSlice(c.done, zip)) {
				return
			}
		}
	}
}

// PadRight adds values to the end of the stream if the stream's values are
// fewer than the passed padding amount once the stream is exhausted.
func (c Stream[T]) PadRight(val T, n int) Stream[T] {
	next := c.source()
	i := 0
	exhausted := false

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		if !exhausted {
			if e, ok := next(done); ok {
				i++
				return e, true
			}
			exhausted = true
		}
		if i >= n || cancelled(done) {
			var zero T
			return zero, false
		}
		i++
		return val, true
	}}
}

// PadLeft adds values to the beginning of the stream if the stream's values
// are fewer than the passed padding amount.
func (c Stream[T]) PadLeft(val T, n int) Stream[T] {
	next := c.source()
	started := false
	padding := 0
	var head []T

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		var zero T

		if !started {
			started = true

			// Hold back the first n values to find out whether padding is
			// needed.
			for len(head) < n {
				e, ok := next(done)
				if !ok {
					break
				}
				head = append(head, e)
			}
			padding = max(0, n-len(head))
		}

		if cancelled(done) {
			return zero, false
		}

		if padding > 0 {
			padding--
			return val, true
		}

		if len(head) > 0 {
			e := head[0]
			head = head[1:]
			return e, true
		}

		return next(done)
	}}
}

// Tee passes each stream value to the passed function. The passed function is
// called once for each value.
func (c Stream[T]) Tee(f func(T)) Stream[T] {
	next := c.source()

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		val, ok := next(done)
		if ok {
			f(val)
		}
		return val, ok
	}}
}

// Enumerate decorates stream values with an enumerated index starting at the
// passed n.
func (c Stream[T]) Enumerate(n int) iter.Seq[Enum[T]] {
	next := c.source()

	return func(yield func(Enum[T]) bool) {
		for {
			val, ok := next(c.done)
			if !ok {
				return
			}
			enum := Enum[T]{Index: n, Val: val}
			n++
			if !yield(enum) {
				return
			}
		}
	}
}

// Find drops stream values until the passed needle value is found then normal
// iteration continues.
func (c Stream[T]) Find(needle T) Stream[T] {
	next := c.source()
	found := false

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		for {
			val, ok := next(done)
			if !ok || found {
				return val, ok
			}
			if val == needle {
				found = true
				return val, true
			}
		}
	}}
}

// Substitute iterates over stream values replacing the passed old value with
// the new value.
func (c Stream[T]) Substitute(old, new T) Stream[T] {
	next := c.source()

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		val, ok := next(done)
		if ok && val == old {
			val = new
		}
		return val, ok
	}}
}

// Skip iterates over stream values skipping those equal to the passed value.
func (c Stream[T]) Skip(needle T) Stream[T] {
	next := c.source()

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		for {
			val, ok := next(done)
			if !ok || val != needle {
				return val, ok
			}
		}
	}}
}

// Throttle iterates over stream values processing no more than n within any
// period of the passed duration.
func (c Stream[T]) Throttle(n int, d time.Duration) Stream[T] {
	if n <= 0 || d <= 0 {
		return c
	}

	next := c.source()

	// The times of the last n values, used as a ring buffer. Each slot holds
	// the time of the value n values ago, or zero if there wasn't one.
	sent := make([]time.Time, n)
	i := 0

	// The time of the last value. The current time is never earlier, so if the
	// window has passed by this time there's no need to read the clock.
	var last time.Time

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		val, ok := next(done)
		if !ok {
			return val, false
		}

		if until := sent[i].Add(d); !sent[i].IsZero() && until.After(last) {
			if wait := time.Until(until); wait > 0 {
				timer := time.NewTimer(wait)
				select {
				case <-timer.C:
				case <-done:
					timer.Stop()
					var zero T
					return zero, false
				}
			}
		}

		last = time.Now()
		sent[i] = last
		i++
		if i == n {
			i = 0
		}

		return val, true
	}}
}

// Distinct iterates over stream values returning only distinct values. This
// method will continually allocate memory tracking distinct values when
// streaming.
func (c Stream[T]) Distinct() Stream[T] {
	next := c.source()
	values := make(map[T]struct{})

	return Stream[T]{c.done, func(done <-chan struct{}) (T, bool) {
		for {
			val, ok := next(done)
			if !ok {
				return val, false
			}
			if _, seen := values[val]; !seen {
				values[val] = struct{}{}
				return val, true
			}
		}
	}}
}

// Buffer runs the stages before it in their own goroutine, letting them get up
// to n values ahead of the stages after it. Values that are buffered when the
// context is cancelled can still be received.
func (c Stream[T]) Buffer(n int) Stream[T] {
	next := c.source()

	return Stream[T]{c.done, concurrently(max(n, 0), func(done <-chan struct{}) []func(func(T) bool) {
		return []func(func(T) bool){func(emit func(T) bool) {
			for {
				val, ok := next(done)
				if !ok || !emit(val) {
					return
				}
			}
		}}
	})}
}
