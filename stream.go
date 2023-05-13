package stream

// Generic channel types.
type Chan[T comparable] chan T
type ChanChan[T comparable] chan Chan[T]

// Take filters a channel to only return n items and then close.
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

// Until closes a channel when the predicate returns true, otherwise it wll keep
// returning values. The predicate is called for each value.
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

// Map maps channel values based on the function argument.
// The function is called for each value.
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

// Filter filters channel values based on the predicate returning true.
// The predicate is called for each value.
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

// Reduce reduces the channel values to one value based on the function
// argument. The function is called for each value.
func (c Chan[T]) Reduce(f func(a, b T) T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		var a T
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
		var last T
		for val := range c {
			last = val
		}
		output <- last
	}()

	return output
}

// Chain will append values from the passed channels when the main channel is
// exhausted.
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

// Zip will alternate taking values from the passed channels and the main
// channel. If any one channel is exhausted the returned channel will close.
func (c Chan[T]) Zip(b Chan[T], args ...Chan[T]) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for {
			val, ok := <-c
			if !ok {
				return
			}
			output <- val
			val, ok = <-b
			if !ok {
				return
			}
			output <- val
			for _, arg := range args {
				val, ok = <-arg
				if !ok {
					return
				}
				output <- val
				break
			}
		}
	}()

	return output
}

// Chunk returns a channel full of slices of the passed length.
// func (c Chan[T]) Chunk(n int) SliceChan[T] {
// 	output := make(SliceChan[T])

// 	go func() {
// 		defer close(output)
// 		for {
// 			chunk := c.Take(n).Slice()
// 			if len(chunk) == 0 {
// 				return
// 			}
// 			output <- chunk
// 			if len(chunk) < n {
// 				return
// 			}
// 		}
// 	}()

// 	return output
// }

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

// Drop removes n values from the channel before continuing.
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
