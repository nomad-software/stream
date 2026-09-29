package stream

import (
	"context"
	"io"
	"math/big"
	"math/rand"
	"strings"
)

// FromSlice creates a channel that will return the items in the passed slice.
// The channel will close when the slice values are exhausted.
func FromSlice[T comparable](ctx context.Context, slice []T) Stream[T] {
	output := newStream[T](ctx, 0)

	go func() {
		defer output.close()

		for _, e := range slice {
			if !output.send(e) {
				return
			}
		}
	}()

	return output
}

// Cycle creates a channel that will repeat the items in the passed slice
// infinitely. This channel will not close by itself and should be limited using
// other methods.
func Cycle[T comparable](ctx context.Context, slice []T) Stream[T] {
	output := newStream[T](ctx, 0)

	go func() {
		defer output.close()

		for i := 0; len(slice) > 0; i++ {
			if i == len(slice) {
				i = 0
			}

			if !output.send(slice[i]) {
				return
			}
		}
	}()

	return output
}

// Generate creates a channel that will return values returned from the passed
// function. This channel will not close by itself and should be limited using
// other methods.
func Generate[T comparable](ctx context.Context, f func() T) Stream[T] {
	output := newStream[T](ctx, 0)

	go func() {
		defer output.close()

		for {
			if !output.send(f()) {
				return
			}
		}
	}()

	return output
}

// Repeat creates a channel that will repeat the passed value infinitely. This
// channel will not close by itself and should be limited using other methods.
func Repeat[T comparable](ctx context.Context, val T) Stream[T] {
	output := newStream[T](ctx, 0)

	go func() {
		defer output.close()

		for {
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// FromChannel creates a channel that will return the values of the passed
// channel. The channel will close when passed channel is closed.
func FromChannel[T comparable](ctx context.Context, c <-chan T) Stream[T] {
	output := newStream[T](ctx, 0)

	go func() {
		defer output.close()

		for {
			select {
			case val, ok := <-c:
				if !ok {
					return
				}

				if !output.send(val) {
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}

// FromString creates a channel that will return strings delimited by a
// separator. The channel will close when the strings are exhausted.
func FromString(ctx context.Context, str string, sep string) Stream[string] {
	output := newStream[string](ctx, 0)

	go func() {
		defer output.close()

		for _, val := range strings.Split(str, sep) {
			if !output.send(val) {
				return
			}
		}
	}()

	return output
}

// FromRunes creates a channel that will return the runes in the string. The
// channel will close when the runes are exhausted.
func FromRunes(ctx context.Context, str string) Stream[rune] {
	output := newStream[rune](ctx, 0)

	go func() {
		defer output.close()

		for _, r := range str {
			if !output.send(r) {
				return
			}
		}
	}()

	return output
}

// FromReader creates a byte channel returning bytes read from the passed reader.
// The channel will close when the reader returns an error. This error could be
// a EOF indicating the data has been exhausted or any other error.
func FromReader(ctx context.Context, r io.Reader) Stream[byte] {
	output := newStream[byte](ctx, 0)
	buffer := make([]byte, 4096) // Default page size.

	go func() {
		defer output.close()

		for {
			n, err := r.Read(buffer)

			for i := range n {
				if !output.send(buffer[i]) {
					return
				}
			}

			if err != nil {
				return
			}
		}
	}()

	return output
}

// Iota creates a channel that will return integers from start (inclusive) to
// end (exclusive), incremented by step. The channel will close when end is
// reached or when the sequence exceeds the channel's type limits.
func Iota(ctx context.Context, start, end, step int) Stream[int] {
	output := newStream[int](ctx, 0)

	if step <= 0 {
		output.close()
		return output
	}

	go func() {
		defer output.close()

		for i := start; i < end; i += step {
			if !output.send(i) {
				return
			}
			// Stop if the next value would reach the end, before adding the
			// step can overflow. The distance is compared as unsigned because
			// it can exceed the int range.
			if uint(end)-uint(i) <= uint(step) {
				return
			}
		}
	}()

	return output
}

// Fibonacci creates an integer channel returning the fibonacci sequence. This
// channel will not close by itself and should be limited using other methods.
func Fibonacci(ctx context.Context) Stream[*big.Int] {
	output := newStream[*big.Int](ctx, 0)

	go func() {
		defer output.close()

		a := big.NewInt(0)
		b := big.NewInt(1)

		for {
			if !output.send(big.NewInt(0).Set(a.Add(a, b))) {
				return
			}
			a, b = b, a
		}
	}()

	return output
}

// Primes creates an integer channel returning prime numbers. The channel will
// close when the sequence exceeds the returned channel's type limits which may
// take a long time.
func Primes(ctx context.Context) Stream[int] {
	output := newStream[int](ctx, 0)

	go func() {
		defer output.close()

		if !output.send(2) {
			return
		}

		primes := make([]int, 0)

		for n := 3; n > 0; n += 2 {
			isPrime := true

			for _, prime := range primes {
				if n%prime == 0 {
					isPrime = false
					break
				}
			}

			if isPrime {
				if !output.send(n) {
					return
				}
				primes = append(primes, n)
			}
		}
	}()

	return output
}

// RandInt creates an integer channel returning random integers. This channel
// will not close by itself and should be limited using other methods.
func RandInt(ctx context.Context) Stream[int] {
	output := newStream[int](ctx, 0)

	go func() {
		defer output.close()

		for {
			if !output.send(rand.Int()) {
				return
			}
		}
	}()

	return output
}

// RandFloat32 creates a (32bit) float channel returning random floats. This
// channel will not close by itself and should be limited using other methods.
func RandFloat32(ctx context.Context) Stream[float32] {
	output := newStream[float32](ctx, 0)

	go func() {
		defer output.close()

		for {
			if !output.send(rand.Float32()) {
				return
			}
		}
	}()

	return output
}

// RandFloat64 creates a (64bit) float channel returning random floats. This
// channel will not close by itself and should be limited using other methods.
func RandFloat64(ctx context.Context) Stream[float64] {
	output := newStream[float64](ctx, 0)

	go func() {
		defer output.close()

		for {
			if !output.send(rand.Float64()) {
				return
			}
		}
	}()

	return output
}
