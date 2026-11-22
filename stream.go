package stream

import (
	"context"
	"iter"
	"sync"
	"time"
)

// Stream is a stream of values bound to a context. Cancelling the context stops
// every stage of the stream. A context is stored here, against the usual
// guidance, because a stream represents the lifetime of a single operation.
// Streams should be created using a generator. The zero value behaves like a nil
// channel and blocks forever.
type Stream[T comparable] struct {
	ctx  context.Context
	done <-chan struct{} // ctx.Done(), cached because it's read for every value.
	c    chan T
}

// Enum adds an enumeration index to a value.
type Enum[T comparable] struct {
	Index int `json:"index"`
	Val   T   `json:"val"`
}

// Passing a value between two goroutines is by far the most expensive thing a
// stream does, so the stages avoid selecting on the context for every value.
// Instead:
//
//   - A stage sends with a plain channel send. If the context is cancelled
//     while it's blocked, a function registered with the context drains the
//     channel, which unblocks the send. The stage then sees the cancellation
//     and stops.
//   - A stage receives from its input with a plain channel receive. Its input
//     shares its context, so the input is closed when the context is cancelled.
//   - Anything that hands values to the caller (All, Pop, Chunk, Zip, Enumerate)
//     or reads a stream bound to another context still selects on the context,
//     so the caller is never left waiting on a stage that's stuck in a callback.

// newStream creates a stream bound to the passed context with a buffer of the
// passed size.
func newStream[T comparable](ctx context.Context, size int) Stream[T] {
	return Stream[T]{
		ctx:  ctx,
		done: ctx.Done(),
		c:    make(chan T, size),
	}
}

// context returns the stream's context, or a background context if the stream
// is the zero value.
func (c Stream[T]) context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
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

// run calls the passed function, which is expected to send values to the
// passed channel, and closes the channel when the function returns. While the
// function is running, cancelling the context drains the channel so sends to
// it cannot block.
func run[T any](ctx context.Context, done <-chan struct{}, c chan T, f func()) {
	if done != nil {
		stop := context.AfterFunc(ctx, func() {
			for range c {
			}
		})
		defer stop()
	}
	defer close(c)
	f()
}

// run calls the passed function, which is expected to send values to the
// stream, and closes the stream when the function returns.
func (c Stream[T]) run(f func()) {
	run(c.ctx, c.done, c.c, f)
}

// send sends a value to the stream. It returns false if the stream's context
// has been cancelled, in which case the stage must stop. It must only be called
// from within the function passed to run.
func (c Stream[T]) send(val T) bool {
	c.c <- val
	return !cancelled(c.done)
}

// recv receives a value from the stream. It returns false if the stream was
// closed or the stream's context was cancelled.
func (c Stream[T]) recv() (T, bool) {
	return recv(c.done, c.c)
}

// recv receives a value from a channel. It returns false if the channel was
// closed or the passed done channel was closed.
func recv[T any](done <-chan struct{}, c <-chan T) (T, bool) {
	if done == nil {
		val, ok := <-c
		return val, ok
	}
	select {
	case val, ok := <-c:
		return val, ok
	case <-done:
		var zero T
		return zero, false
	}
}

// recvIn receives a value from the stream on behalf of a stage whose context
// has the passed done channel. It returns false if the stream was closed or
// that context was cancelled.
func (c Stream[T]) recvIn(done <-chan struct{}) (T, bool) {
	if done == c.done {
		// The stream is closed when the context is cancelled.
		val, ok := <-c.c
		return val, ok
	}
	return recv(done, c.c)
}

// in returns an iterator over the stream's values on behalf of a stage whose
// context has the passed done channel.
func (c Stream[T]) in(done <-chan struct{}) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			val, ok := c.recvIn(done)
			if !ok || !yield(val) {
				return
			}
		}
	}
}

// close closes the stream.
func (c Stream[T]) close() {
	close(c.c)
}

// All returns an iterator over the main channel values. Iteration stops when
// the main channel is closed or its context is cancelled.
func (c Stream[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		if c.done == nil {
			for val := range c.c {
				if !yield(val) {
					return
				}
			}
			return
		}
		for {
			val, ok := c.recv()
			if !ok || !yield(val) {
				return
			}
		}
	}
}

// Take returns n items from the main channel before closing it.
func (c Stream[T]) Take(n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for range n {
			val, ok := <-c.c
			if !ok {
				return
			}
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Until closes a channel when the passed function returns true, otherwise it
// wll keep returning values. The passed function is called once for each value.
func (c Stream[T]) Until(f func(T) bool) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			if f(val) {
				return
			}
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Map mutates main channel values based on the passed function. The passed
// function is called once for each value.
func (c Stream[T]) Map[R comparable](f func(T) R) Stream[R] {
	output := newStream[R](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			if !output.send(f(val)) {
				return
			}
		}
	})

	return output
}

// Map mutates main channel values based on the passed function. The passes
// workers value is the amount of parallel workers spawned. The passed function
// is called once for each value. The streamed values are not ordered.
func (c Stream[T]) MapParallel[R comparable](workers int, f func(T) R) Stream[R] {
	output := newStream[R](c.context(), 0)

	if workers <= 0 {
		output.close()
		return output
	}

	wg := new(sync.WaitGroup)
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()
			for val := range c.c {
				if !output.send(f(val)) {
					return
				}
			}
		}()
	}

	go output.run(wg.Wait)

	return output
}

// Filter filters main channel values based on the passed function returning
// true. The passed function is called once for each value.
func (c Stream[T]) Filter(f func(T) bool) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			if !f(val) {
				continue
			}
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Reduce reduces main channel values to one value based on the passed function.
// The passed function is called once for each value.
func (c Stream[T]) Reduce(f func(T, T) T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {

		a, ok := <-c.c
		if !ok {
			return
		}
		for val := range c.c {
			a = f(a, val)
		}
		if c.context().Err() != nil {
			return
		}
		output.send(a)
	})

	return output
}

// Last will return the final value from the channel once it is closed.
func (c Stream[T]) Last() Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		last, ok := <-c.c
		if !ok {
			return
		}
		for val := range c.c {
			last = val
		}
		if c.context().Err() != nil {
			return
		}
		output.send(last)
	})

	return output
}

// Chain will append values from the passed channels to the end of the main
// channel. The main channel's context is used for all channels.
func (c Stream[T]) Chain(b Stream[T], args ...Stream[T]) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			if !output.send(val) {
				return
			}
		}
		for val := range b.in(c.done) {
			if !output.send(val) {
				return
			}
		}
		for _, arg := range args {
			for val := range arg.in(c.done) {
				if !output.send(val) {
					return
				}
			}
		}
	})

	return output
}

// Merge will return alternate values from the main channel and the passed
// channels, when they are available.
func (c Stream[T]) Merge(b Stream[T], args ...Stream[T]) Stream[T] {
	output := newStream[T](c.context(), 0)

	wg := new(sync.WaitGroup)
	wg.Add(2 + len(args))

	go func() {
		defer wg.Done()
		for val := range c.c {
			if !output.send(val) {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for val := range b.in(c.done) {
			if !output.send(val) {
				return
			}
		}
	}()

	for _, arg := range args {
		go func() {
			defer wg.Done()
			for val := range arg.in(c.done) {
				if !output.send(val) {
					return
				}
			}
		}()
	}

	go output.run(wg.Wait)

	return output
}

// RoundRobin will return alternate values from the main channel and the passed
// channels, in order. The main channel's context is used for all channels.
func (c Stream[T]) RoundRobin(b Stream[T], args ...Stream[T]) Stream[T] {
	output := newStream[T](c.context(), 0)
	inputs := append([]Stream[T]{c, b}, args...)

	go output.run(func() {
		for {
			available := false
			for _, input := range inputs {
				val, ok := input.recvIn(c.done)
				if !ok {
					continue
				}
				available = true
				if !output.send(val) {
					return
				}
			}
			if !available {
				return
			}
		}
	})

	return output
}

// Chunk returns a channel full of channels of the passed length, filled with
// values of the main channel.
func (c Stream[T]) Chunk(n int) chan Stream[T] {
	output := make(chan Stream[T])

	if n <= 0 {
		close(output)
		return output
	}

	go run(c.context(), c.done, output, func() {
		for {
			// The chunk is never sent to more than it can hold and is closed
			// before it's handed over, so it needs no draining.
			chunk := Stream[T]{ctx: c.ctx, done: c.done, c: make(chan T, n)}
			for i := range n {
				val, ok := c.recv()
				if !ok {
					if i > 0 && c.context().Err() == nil {
						chunk.close()
						output <- chunk
					}
					return
				}
				chunk.c <- val
			}
			chunk.close()
			output <- chunk
			if cancelled(c.done) {
				return
			}
		}
	})

	return output
}

// Drop removes n values from the main channel before continuing.
func (c Stream[T]) Drop(n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for range n {
			_, ok := <-c.c
			if !ok {
				return
			}
		}
		for val := range c.c {
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Stride iterates over channel values returning every n value of the main
// channel.
func (c Stream[T]) Stride(n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	if n <= 0 {
		output.close()
		return output
	}

	go output.run(func() {
		i := 0
		for val := range c.c {
			if i%n == 0 {
				if !output.send(val) {
					return
				}
				i = 0
			}
			i++
		}
	})

	return output
}

// Tail returns a channel containing the last n values of the main channel once
// it's closed.
func (c Stream[T]) Tail(n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	if n <= 0 {
		output.close()
		return output
	}

	go output.run(func() {
		// A ring buffer of the last n values, where i is the oldest value
		// once the buffer is full.
		tail := make([]T, 0, min(n, 1024))
		i := 0
		for val := range c.c {
			if len(tail) < n {
				tail = append(tail, val)
				continue
			}
			tail[i] = val
			i++
			if i == n {
				i = 0
			}
		}
		if c.context().Err() != nil {
			return
		}
		for _, val := range tail[i:] {
			if !output.send(val) {
				return
			}
		}
		for _, val := range tail[:i] {
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Zip returns a channel of channels containing the next values of the main
// channel and all other passed channels, in order. The main channel's context
// is used for all channels.
func (c Stream[T]) Zip(b Stream[T], args ...Stream[T]) chan Stream[T] {
	output := make(chan Stream[T])
	inputs := append([]Stream[T]{c, b}, args...)

	go run(c.context(), c.done, output, func() {
		for {
			zip := Stream[T]{ctx: c.ctx, done: c.done, c: make(chan T, len(inputs))}
			for _, input := range inputs {
				val, ok := recv(c.done, input.c)
				if !ok {
					return
				}
				zip.c <- val
			}
			zip.close()
			output <- zip
			if cancelled(c.done) {
				return
			}
		}
	})

	return output
}

// PadRight adds values to the end of the main channel if that channel's values
// are fewer than the passed padding amount once the channel is closed.
func (c Stream[T]) PadRight(val T, n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {

		i := 0
		for val := range c.c {
			if !output.send(val) {
				return
			}
			i++
		}

		if c.context().Err() != nil {
			return
		}

		for ; i < n; i++ {
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// PadLeft adds values to the beginning of the main channel if that channel's values
// are fewer than the passed padding amount.
func (c Stream[T]) PadLeft(val T, n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {

		head := make([]T, 0)
		for range n {
			e, ok := <-c.c
			if !ok {
				break
			}
			head = append(head, e)
		}

		if c.context().Err() != nil {
			return
		}

		for range n - len(head) {
			if !output.send(val) {
				return
			}
		}

		for _, e := range head {
			if !output.send(e) {
				return
			}
		}

		for e := range c.c {
			if !output.send(e) {
				return
			}
		}
	})

	return output
}

// Tee passes each main channel value to the passed function. The passed
// function is called once for each value.
func (c Stream[T]) Tee(f func(T)) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			f(val)
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Enumerate decorates main channel values with an enumerated index starting at
// the passed n.
func (c Stream[T]) Enumerate(n int) chan Enum[T] {
	output := make(chan Enum[T])

	go run(c.context(), c.done, output, func() {
		for val := range c.All() {
			output <- Enum[T]{
				Index: n,
				Val:   val,
			}
			if cancelled(c.done) {
				return
			}
			n++
		}
	})

	return output
}

// Find drains the main channel until the passed needle value is found then
// normal iteration continues.
func (c Stream[T]) Find(needle T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			if val == needle {
				if !output.send(val) {
					return
				}
				break
			}
		}
		for val := range c.c {
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Substitute iterates over main channel values replacing the passed old value
// with the new value.
func (c Stream[T]) Substitute(old, new T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			if val == old {
				val = new
			}
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Skip iterates over main channel values skipping those equal to the passed
// value.
func (c Stream[T]) Skip(needle T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		for val := range c.c {
			if val == needle {
				continue
			}
			if !output.send(val) {
				return
			}
		}
	})

	return output
}

// Throttle iterates over main channel values processing no more than n within
// any period of the passed duration.
func (c Stream[T]) Throttle(n int, d time.Duration) Stream[T] {
	if n <= 0 || d <= 0 {
		return c
	}

	output := newStream[T](c.context(), 0)

	go output.run(func() {

		// The times of the last n sends, used as a ring buffer. Each slot
		// holds the time of the send n sends ago, or zero if there wasn't one.
		sent := make([]time.Time, n)
		i := 0

		// The time of the last send. The current time is never earlier, so if
		// the window has passed by this time there's no need to read the clock.
		var last time.Time

		for val := range c.c {
			if next := sent[i].Add(d); !sent[i].IsZero() && next.After(last) {
				if wait := time.Until(next); wait > 0 {
					select {
					case <-time.After(wait):
					case <-c.done:
						return
					}
				}
			}

			if !output.send(val) {
				return
			}
			last = time.Now()
			sent[i] = last
			i++
			if i == n {
				i = 0
			}
		}
	})

	return output
}

// Distinct iterates over main channel values returning only distinct values.
// This method will continually allocate memory tracking distinct values when
// streaming.
func (c Stream[T]) Distinct() Stream[T] {
	output := newStream[T](c.context(), 0)

	go output.run(func() {
		values := make(map[T]struct{})
		for val := range c.c {
			if _, ok := values[val]; ok {
				continue
			}
			if !output.send(val) {
				return
			}
			values[val] = struct{}{}
		}
	})

	return output
}

// Buffer iterates over main channel values returning a buffered channel for n
// values.
func (c Stream[T]) Buffer(n int) Stream[T] {
	output := newStream[T](c.context(), max(n, 0))

	// Values that are buffered when the context is cancelled can still be
	// received, so the output isn't drained on cancellation like other stages.
	// The send selects on the context instead.
	go func() {
		defer output.close()
		for val := range c.c {
			select {
			case output.c <- val:
				continue
			default:
			}
			select {
			case output.c <- val:
			case <-output.done:
				return
			}
		}
	}()

	return output
}
