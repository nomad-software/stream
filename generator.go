package stream

import (
	"io"
	"math/big"
	"math/rand"
	"strings"
)

// FromSlice creates a channel that will return the items in the passed slice.
// The channel will close when the slice values are exhausted.
func FromSlice[T comparable](slice []T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for _, e := range slice {
			output <- e
		}
	}()

	return output
}

// Cycle creates a channel that will repeat the items in the passed slice
// infinitely. This channel will not close by itself and should be limited using
// other methods.
func Cycle[T comparable](slice []T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for i := 0; i < len(slice); i++ {
			output <- slice[i]
			if i == len(slice)-1 {
				i = -1
			}
		}
	}()

	return output
}

// Generate creates a channel that will return values returned from the passed
// function. This channel will not close by itself and should be limited using
// other methods.
func Generate[T comparable](f func() T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for {
			output <- f()
		}
	}()

	return output
}

// Repeat creates a channel that will repeat the passed value infinitely. This
// channel will not close by itself and should be limited using other methods.
func Repeat[T comparable](val T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for {
			output <- val
		}
	}()

	return output
}

// FromChannel creates a channel that will return the values of the passed
// channel. The channel will close when passed channel is closed.
func FromChannel[T comparable](c <-chan T) Chan[T] {
	output := make(Chan[T])

	go func() {
		defer close(output)
		for val := range c {
			output <- val
		}
	}()

	return output
}

// FromString creates a channel that will return strings delimited by a
// separator. The channel will close when the strings are exhausted.
func FromString(str string, sep string) Chan[string] {
	output := make(Chan[string])

	go func() {
		defer close(output)
		for val := range FromSlice(strings.Split(str, sep)) {
			output <- val
		}
	}()

	return output
}

// FromRunes creates a channel that will return the runes in the string. The
// channel will close when the runes are exhausted.
func FromRunes(str string) Chan[rune] {
	output := make(Chan[rune])

	go func() {
		defer close(output)
		for _, r := range str {
			output <- r
		}
	}()

	return output
}

// Iota creates a channel that will return integers based on the supplied
// arguments. This channel will not close by itself and should be limited using
// other methods.
func Iota(start, end, step int) Chan[int] {
	output := make(Chan[int])

	go func() {
		defer close(output)
		for i := start; i < end; i += step {
			output <- i
		}
	}()

	return output
}

// Fibonacci creates an integer channel returning the fibonacci sequence. This
// channel will not close by itself and should be limited using other methods.
func Fibonacci() Chan[*big.Int] {
	output := make(Chan[*big.Int])

	go func() {
		defer close(output)
		a := big.NewInt(0)
		b := big.NewInt(1)
		for {
			output <- big.NewInt(0).Set(a.Add(a, b))
			a, b = b, a
		}
	}()

	return output
}

// Primes creates an integer channel returning prime numbers. The channel will
// close when the sequence exceeds the returned channel's type limits which may
// take a long time.
func Primes() Chan[int] {
	output := make(Chan[int])

	go func() {
		output <- 2
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
				if n < 0 {
					return
				}
				output <- n
				primes = append(primes, n)
			}
		}
	}()

	return output
}

// RandInt creates an integer channel returning random integers. This channel
// will not close by itself and should be limited using other methods.
func RandInt() Chan[int] {
	output := make(Chan[int])

	go func() {
		defer close(output)
		for {
			output <- rand.Int()
		}
	}()

	return output
}

// RandFloat32 creates a (32bit) float channel returning random floats. This
// channel will not close by itself and should be limited using other methods.
func RandFloat32() Chan[float32] {
	output := make(Chan[float32])

	go func() {
		defer close(output)
		for {
			output <- rand.Float32()
		}
	}()

	return output
}

// RandFloat64 creates a (64bit) float channel returning random floats. This
// channel will not close by itself and should be limited using other methods.
func RandFloat64() Chan[float64] {
	output := make(Chan[float64])

	go func() {
		defer close(output)
		for {
			output <- rand.Float64()
		}
	}()

	return output
}

// ReadFrom creates a byte channel returning bytes read from the passed reader.
// The channel will close when the reader returns an error. This error could be
// a EOF indicating the data has been exhausted or any other error.
func ReadFrom(r io.Reader) Chan[byte] {
	output := make(Chan[byte])

	go func() {
		defer close(output)
		buf := make([]byte, 8)
		for {
			n, err := r.Read(buf)
			for i := 0; i < n; i++ {
				output <- buf[i]
			}
			if err != nil {
				return
			}
		}
	}()

	return output
}
