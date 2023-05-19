package stream

// Generic channel types.
type Chan[T comparable] chan T
type ChanChan[T comparable] chan Chan[T]
type ChanEnum[T comparable] chan Enum[T]

// Enum adds an enumeration index to a value.
type Enum[T comparable] struct {
	Index int `json:"index"`
	Val   T   `json:"val"`
}

// Take returns n items from the main channel before closing it.
func (c Chan[T]) Take(n int) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for i := 0; i < n; i++ {
			val, ok := <-c
			if !ok {
				return
			}
			output <- val
		}
	}()

	return output
}

// Until closes a channel when the passed function returns true, otherwise it
// wll keep returning values. The passed function is called once for each value.
func (c Chan[T]) Until(f func(val T) bool) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for {
			val, ok := <-c
			if !ok {
				return
			}
			if f(val) {
				return
			}
			output <- val
		}
	}()

	return output
}

// Map mutates main channel values based on the passed function. The passed
// function is called once for each value.
func (c Chan[T]) Map(f func(val T) T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for {
			val, ok := <-c
			if !ok {
				return
			}
			output <- f(val)
		}
	}()

	return output
}

// Filter filters main channel values based on the passed function returning
// true. The passed function is called once for each value.
func (c Chan[T]) Filter(f func(val T) bool) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for {
			val, ok := <-c
			if !ok {
				return
			}
			if f(val) {
				output <- val
			}
		}
	}()

	return output
}

// Reduce reduces main channel values to one value based on the passed function.
// The passed function is called once for each value.
func (c Chan[T]) Reduce(f func(a, b T) T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)

		var a T = <-c
		for val := range c {
			a = f(a, val)
		}
		output <- a
	}()

	return output
}

// Last will return the final value from the channel once it is closed.
func (c Chan[T]) Last() Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		var last T = <-c
		for val := range c {
			last = val
		}
		output <- last
	}()

	return output
}

// Chain will append values from the passed channels to the end of the main
// channel.
func (c Chan[T]) Chain(b Chan[T], args ...Chan[T]) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for val := range c {
			output <- val
		}
		for val := range b {
			output <- val
		}
		for _, arg := range args {
			for val := range arg {
				output <- val
			}
		}
	}()

	return output
}

// RoundRobin will return alternate values from the main channel and the passed
// channels, in order.
func (c Chan[T]) RoundRobin(b Chan[T], args ...Chan[T]) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for {
			available := false
			val, ok := <-c
			if ok {
				available = true
				output <- val
			}
			val, ok = <-b
			if ok {
				available = true
				output <- val
			}
			for _, arg := range args {
				val, ok = <-arg
				if ok {
					available = true
					output <- val
				}
				break
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
func (c Chan[T]) Chunk(n int) ChanChan[T] {
	output := make(ChanChan[T])

	go func() {
		defer close(output)
		for {
			chunk := make(Chan[T], n)
			for i := 0; i < n; i++ {
				val, ok := <-c
				if !ok {
					if i > 0 {
						close(chunk)
						output <- chunk
					}
					return
				}
				chunk <- val
			}
			close(chunk)
			output <- chunk
		}
	}()

	return output
}

// Drop removes n values from the main channel before continuing.
func (c Chan[T]) Drop(n int) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for i := 0; i < n; i++ {
			_, ok := <-c
			if !ok {
				return
			}
		}
		for val := range c {
			output <- val
		}
	}()

	return output
}

// Stride iterates over channel values returning every n value of the main
// channel.
func (c Chan[T]) Stride(n int) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		i := 0
		for val := range c {
			if i%n == 0 {
				output <- val
				i = 0
			}
			i++
		}
	}()

	return output
}

// Tail returns a channel containing the last n values of the main channel once
// it's closed.
func (c Chan[T]) Tail(n int) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		tail := make(Chan[T], n)
		i := 0
		for val := range c {
			if i >= n {
				<-tail
			} else {
				i++
			}
			tail <- val
		}
		close(tail)
		for val := range tail {
			output <- val
		}
	}()

	return output
}

// Zip returns a channel of channels containing the next values of the main
// channel and all other passed channels, in order.
func (c Chan[T]) Zip(b Chan[T], args ...Chan[T]) ChanChan[T] {
	output := make(ChanChan[T])

	go func() {
		defer close(output)
		for {
			zip := make(Chan[T], len(args)+2)
			val, ok := <-c
			if !ok {
				return
			}
			zip <- val
			val, ok = <-b
			if !ok {
				return
			}
			zip <- val
			for _, arg := range args {
				val, ok = <-arg
				if !ok {
					return
				}
				zip <- val
				break
			}
			close(zip)
			output <- zip
		}
	}()

	return output
}

// PadRight adds values to the end of the main channel if that channel's values
// are fewer than the passed padding amount once the channel is closed.
func (c Chan[T]) PadRight(val T, n int) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		i := 0
		for val := range c {
			output <- val
			i++
		}
		if i < n {
			for val := range Repeat(val).Take(n - i) {
				output <- val
			}
		}
	}()

	return output
}

// PadLeft adds values to the beginning of the main channel if that channel's values
// are fewer than the passed padding amount.
func (c Chan[T]) PadLeft(val T, n int) Chan[T] {
	output := make(Chan[T], n)

	for val := range Repeat(val).Take(n) {
		output <- val
	}

	for i := 0; i < n; i++ {
		val, ok := <-c
		if !ok {
			break
		}
		<-output
		output <- val
	}

	go func() {
		defer close(output)
		for val := range c {
			output <- val
		}
	}()

	return output
}

// Tee passes each main channel value to the passed function. The passed
// function is called once for each value.
func (c Chan[T]) Tee(f func(val T)) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for val := range c {
			f(val)
			output <- val
		}
	}()

	return output
}

// Enumerate decorates main channel values with an enumerated index starting at
// the passed n.
func (c Chan[T]) Enumerate(n int) chan Enum[T] {
	output := make(chan Enum[T])

	go func() {
		defer close(output)
		for val := range c {
			output <- Enum[T]{
				Index: n,
				Val:   val,
			}
			n++
		}
	}()

	return output
}
