package stream

import (
	"math"
	"math/rand"
	"strings"
)

// FromSlice creates a channel that will return the items in the passed slice.
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
// infinitely.
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
// function.
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

// Repeat creates a channel that will repeat the passed value infinitely.
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

// FromString creates a channel that will return strings delimited by a separator.
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

// FromRunes creates a channel that will return the runes in the string.
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

// Iota creates a channel that will return integers based on the supplied arguments.
// This channel will not close by itself and should be limited using other methods.
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

// Fibonacci creates an integer channel returning the fibonacci sequence.
// The channel will close when the sequence exceeds the returned channel's
// type limits.
func Fibonacci() Chan[uint] {
	output := make(Chan[uint])

	go func() {
		defer close(output)
		var a uint = 0
		var b uint = 1
		for {
			if a > math.MaxInt64-b {
				return
			}
			a = (a + b)
			output <- a
			a, b = b, a
		}
	}()

	return output
}

// RandInt creates an integer channel returning random integers.
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

// RandFloat32 creates a (32bit) float channel returning random floats.
func RandFloat32() Chan[float64] {
	output := make(Chan[float64])

	go func() {
		defer close(output)
		for {
			output <- rand.Float64()
		}
	}()

	return output
}

// RandFloat64 creates a (64bit) float channel returning random floats.
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
