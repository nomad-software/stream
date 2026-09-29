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
	ctx context.Context
	c   chan T
}

// Enum adds an enumeration index to a value.
type Enum[T comparable] struct {
	Index int `json:"index"`
	Val   T   `json:"val"`
}

// newStream creates a stream bound to the passed context with a buffer of the
// passed size.
func newStream[T comparable](ctx context.Context, size int) Stream[T] {
	return Stream[T]{
		ctx: ctx,
		c:   make(chan T, size),
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

// withContext returns the same stream bound to the passed context.
func (c Stream[T]) withContext(ctx context.Context) Stream[T] {
	c.ctx = ctx
	return c
}

// send sends a value to a channel. It returns false if the context was
// cancelled before the value could be sent.
func send[T any](ctx context.Context, c chan<- T, val T) bool {
	select {
	case c <- val:
		return true
	case <-ctx.Done():
		return false
	}
}

// recv receives a value from a channel. It returns false if the channel was
// closed or the context was cancelled.
func recv[T any](ctx context.Context, c <-chan T) (T, bool) {
	select {
	case val, ok := <-c:
		return val, ok
	case <-ctx.Done():
		var zero T
		return zero, false
	}
}

// send sends a value to the stream. It returns false if the stream's context
// was cancelled before the value could be sent.
func (c Stream[T]) send(val T) bool {
	return send(c.context(), c.c, val)
}

// recv receives a value from the stream. It returns false if the stream was
// closed or the stream's context was cancelled.
func (c Stream[T]) recv() (T, bool) {
	return recv(c.context(), c.c)
}

// close closes the stream.
func (c Stream[T]) close() {
	close(c.c)
}

// All returns an iterator over the main channel values. Iteration stops when
// the main channel is closed or its context is cancelled.
func (c Stream[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			val, ok := c.recv()
			if !ok {
				return
			}
			if !yield(val) {
				return
			}
		}
	}
}

// Take returns n items from the main channel before closing it.
func (c Stream[T]) Take(n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for range n {
			val, ok := c.recv()
			if !ok {
				return
			}
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Until closes a channel when the passed function returns true, otherwise it
// wll keep returning values. The passed function is called once for each value.
func (c Stream[T]) Until(f func(T) bool) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			if f(val) {
				return
			}
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Map mutates main channel values based on the passed function. The passed
// function is called once for each value.
func (c Stream[T]) Map[R comparable](f func(T) R) Stream[R] {
	output := newStream[R](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			if !output.send(f(val)) {
				return
			}
		}
	}()

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
			for val := range c.All() {
				if !output.send(f(val)) {
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		output.close()
	}()

	return output
}

// Filter filters main channel values based on the passed function returning
// true. The passed function is called once for each value.
func (c Stream[T]) Filter(f func(T) bool) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			if !f(val) {
				continue
			}
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Reduce reduces main channel values to one value based on the passed function.
// The passed function is called once for each value.
func (c Stream[T]) Reduce(f func(T, T) T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()

		a, ok := c.recv()
		if !ok {
			return
		}
		for val := range c.All() {
			a = f(a, val)
		}
		if c.context().Err() != nil {
			return
		}
		output.send(a)
	}()

	return output
}

// Last will return the final value from the channel once it is closed.
func (c Stream[T]) Last() Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		last, ok := c.recv()
		if !ok {
			return
		}
		for val := range c.All() {
			last = val
		}
		if c.context().Err() != nil {
			return
		}
		output.send(last)
	}()

	return output
}

// Chain will append values from the passed channels to the end of the main
// channel. The main channel's context is used for all channels.
func (c Stream[T]) Chain(b Stream[T], args ...Stream[T]) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			if !output.send(val) {
				return
			}
		}
		for val := range b.withContext(c.context()).All() {
			if !output.send(val) {
				return
			}
		}
		for _, arg := range args {
			for val := range arg.withContext(c.context()).All() {
				if !output.send(val) {
					return
				}
			}
		}
	}()

	return output
}

// Merge will return alternate values from the main channel and the passed
// channels, when they are available.
func (c Stream[T]) Merge(b Stream[T], args ...Stream[T]) Stream[T] {
	output := newStream[T](c.context(), 0)

	wg := new(sync.WaitGroup)
	wg.Add(2+len(args))

	go func() {
		defer wg.Done()
		for val := range c.All() {
			if !output.send(val) {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for val := range b.withContext(c.context()).All() {
			if !output.send(val) {
				return
			}
		}
	}()

	for _, arg := range args {
		go func() {
			defer wg.Done()
			for val := range arg.withContext(c.context()).All() {
				if !output.send(val) {
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		output.close()
	}()

	return output
}

// RoundRobin will return alternate values from the main channel and the passed
// channels, in order. The main channel's context is used for all channels.
func (c Stream[T]) RoundRobin(b Stream[T], args ...Stream[T]) Stream[T] {
	output := newStream[T](c.context(), 0)
	inputs := []Stream[T]{c, b.withContext(c.context())}
	for _, arg := range args {
		inputs = append(inputs, arg.withContext(c.context()))
	}

	go func() {
		defer output.close()
		for {
			available := false
			for _, input := range inputs {
				val, ok := input.recv()
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
	}()

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

	go func() {
		defer close(output)
		for {
			chunk := newStream[T](c.context(), n)
			for i := range n {
				val, ok := c.recv()
				if !ok {
					if i > 0 && c.context().Err() == nil {
						chunk.close()
						send(c.context(), output, chunk)
					}
					return
				}
				chunk.c <- val
			}
			chunk.close()
			if !send(c.context(), output, chunk) {
				return
			}
		}
	}()

	return output
}

// Drop removes n values from the main channel before continuing.
func (c Stream[T]) Drop(n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for range n {
			_, ok := c.recv()
			if !ok {
				return
			}
		}
		for val := range c.All() {
			if !output.send(val) {
				return
			}
		}
	}()

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

	go func() {
		defer output.close()
		i := 0
		for val := range c.All() {
			if i%n == 0 {
				if !output.send(val) {
					return
				}
				i = 0
			}
			i++
		}
	}()

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

	go func() {
		defer output.close()
		tail := make(chan T, n)
		i := 0
		for val := range c.All() {
			if i >= n {
				<-tail
			} else {
				i++
			}
			tail <- val
		}
		if c.context().Err() != nil {
			return
		}
		close(tail)
		for val := range tail {
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Zip returns a channel of channels containing the next values of the main
// channel and all other passed channels, in order. The main channel's context
// is used for all channels.
func (c Stream[T]) Zip(b Stream[T], args ...Stream[T]) chan Stream[T] {
	output := make(chan Stream[T])
	inputs := []Stream[T]{c, b.withContext(c.context())}
	for _, arg := range args {
		inputs = append(inputs, arg.withContext(c.context()))
	}

	go func() {
		defer close(output)
		for {
			zip := newStream[T](c.context(), len(inputs))
			for _, input := range inputs {
				val, ok := input.recv()
				if !ok {
					return
				}
				zip.c <- val
			}
			zip.close()
			if !send(c.context(), output, zip) {
				return
			}
		}
	}()

	return output
}

// PadRight adds values to the end of the main channel if that channel's values
// are fewer than the passed padding amount once the channel is closed.
func (c Stream[T]) PadRight(val T, n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()

		i := 0
		for val := range c.All() {
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
	}()

	return output
}

// PadLeft adds values to the beginning of the main channel if that channel's values
// are fewer than the passed padding amount.
func (c Stream[T]) PadLeft(val T, n int) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()

		head := make([]T, 0)
		for range n {
			e, ok := c.recv()
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

		for e := range c.All() {
			if !output.send(e) {
				return
			}
		}
	}()

	return output
}

// Tee passes each main channel value to the passed function. The passed
// function is called once for each value.
func (c Stream[T]) Tee(f func(T)) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			f(val)
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Enumerate decorates main channel values with an enumerated index starting at
// the passed n.
func (c Stream[T]) Enumerate(n int) chan Enum[T] {
	output := make(chan Enum[T])

	go func() {
		defer close(output)
		for val := range c.All() {
			enum := Enum[T]{
				Index: n,
				Val:   val,
			}
			if !send(c.context(), output, enum) {
				return
			}
			n++
		}
	}()

	return output
}

// Find drains the main channel until the passed needle value is found then
// normal iteration continues.
func (c Stream[T]) Find(needle T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			if val == needle {
				if !output.send(val) {
					return
				}
				break
			}
		}
		for val := range c.All() {
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Substitute iterates over main channel values replacing the passed old value
// with the new value.
func (c Stream[T]) Substitute(old, new T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			if val == old {
				val = new
			}
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Skip iterates over main channel values skipping those equal to the passed
// value.
func (c Stream[T]) Skip(needle T) Stream[T] {
	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()
		for val := range c.All() {
			if val == needle {
				continue
			}
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// Throttle iterates over main channel values processing no more than n within
// any period of the passed duration.
func (c Stream[T]) Throttle(n int, d time.Duration) Stream[T] {
	if n <= 0 || d <= 0 {
		return c
	}

	output := newStream[T](c.context(), 0)

	go func() {
		defer output.close()

		// The times of the last n sends, used as a ring buffer. Each slot
		// holds the time of the send n sends ago, or zero if there wasn't one.
		sent := make([]time.Time, n)
		i := 0

		for val := range c.All() {
			if wait := time.Until(sent[i].Add(d)); wait > 0 {
				select {
				case <-time.After(wait):
				case <-c.context().Done():
					return
				}
			}

			if !output.send(val) {
				return
			}
			sent[i] = time.Now()
			i = (i + 1) % n
		}
	}()

	return output
}

// Distinct iterates over main channel values returning only distinct values.
// This method will continually allocate memory tracking distinct values when
// streaming.
func (c Stream[T]) Distinct() Stream[T] {
	output := newStream[T](c.context(), 0)
	values := make(map[T]*T)

	go func() {
		defer output.close()
		for val := range c.All() {
			if _, ok := values[val]; ok {
				continue
			}
			if !output.send(val) {
				return
			}
			values[val] = nil
		}
	}()

	return output
}

// Buffer iterates over main channel values returning a buffered channel for n
// values.
func (c Stream[T]) Buffer(n int) Stream[T] {
	output := newStream[T](c.context(), max(n, 0))

	go func() {
		defer output.close()
		for val := range c.All() {
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}
