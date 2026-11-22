package stream

import (
	"context"
	"io"
	"math/big"
	"math/rand"
	"strings"
	"unicode/utf8"
)

// fromSlice creates a stream, bound to a context with the passed done channel,
// that will return the items in the passed slice.
func fromSlice[T comparable](done <-chan struct{}, slice []T) Stream[T] {
	i := 0

	return Stream[T]{done, func(done <-chan struct{}) (T, bool) {
		if i >= len(slice) || cancelled(done) {
			var zero T
			return zero, false
		}
		val := slice[i]
		i++
		return val, true
	}}
}

// FromSlice creates a stream that will return the items in the passed slice.
// The stream ends when the slice values are exhausted.
func FromSlice[T comparable](ctx context.Context, slice []T) Stream[T] {
	return fromSlice(ctx.Done(), slice)
}

// Cycle creates a stream that will repeat the items in the passed slice
// infinitely. This stream will not end by itself and should be limited using
// other methods.
func Cycle[T comparable](ctx context.Context, slice []T) Stream[T] {
	i := 0

	return newStream(ctx, func(done <-chan struct{}) (T, bool) {
		if len(slice) == 0 || cancelled(done) {
			var zero T
			return zero, false
		}
		val := slice[i]
		i++
		if i == len(slice) {
			i = 0
		}
		return val, true
	})
}

// Generate creates a stream that will return values returned from the passed
// function. This stream will not end by itself and should be limited using
// other methods.
func Generate[T comparable](ctx context.Context, f func() T) Stream[T] {
	return newStream(ctx, func(done <-chan struct{}) (T, bool) {
		if cancelled(done) {
			var zero T
			return zero, false
		}
		return f(), true
	})
}

// Repeat creates a stream that will repeat the passed value infinitely. This
// stream will not end by itself and should be limited using other methods.
func Repeat[T comparable](ctx context.Context, val T) Stream[T] {
	return newStream(ctx, func(done <-chan struct{}) (T, bool) {
		if cancelled(done) {
			var zero T
			return zero, false
		}
		return val, true
	})
}

// FromChannel creates a stream that will return the values of the passed
// channel. The stream ends when the passed channel is closed.
func FromChannel[T comparable](ctx context.Context, c <-chan T) Stream[T] {
	return newStream(ctx, func(done <-chan struct{}) (T, bool) {
		// The context is checked first because select chooses randomly when
		// both a value and the cancellation are ready.
		if !cancelled(done) {
			select {
			case val, ok := <-c:
				return val, ok
			case <-done:
			}
		}
		var zero T
		return zero, false
	})
}

// FromString creates a stream that will return strings delimited by a
// separator. The stream ends when the strings are exhausted.
func FromString(ctx context.Context, str string, sep string) Stream[string] {
	// An empty separator splits the string into its UTF-8 sequences, of which
	// an empty string has none.
	finished := sep == "" && str == ""

	return newStream(ctx, func(done <-chan struct{}) (string, bool) {
		if finished || cancelled(done) {
			return "", false
		}

		var val string
		if sep == "" {
			_, size := utf8.DecodeRuneInString(str)
			val, str = str[:size], str[size:]
			finished = str == ""
		} else if i := strings.Index(str, sep); i >= 0 {
			val, str = str[:i], str[i+len(sep):]
		} else {
			val, finished = str, true
		}
		return val, true
	})
}

// FromRunes creates a stream that will return the runes in the string. The
// stream ends when the runes are exhausted.
func FromRunes(ctx context.Context, str string) Stream[rune] {
	return newStream(ctx, func(done <-chan struct{}) (rune, bool) {
		if str == "" || cancelled(done) {
			return 0, false
		}
		r, size := utf8.DecodeRuneInString(str)
		str = str[size:]
		return r, true
	})
}

// FromReader creates a byte stream returning bytes read from the passed reader.
// The stream ends when the reader returns an error. This error could be a EOF
// indicating the data has been exhausted or any other error. Bytes are read
// from the reader a page at a time, so more may be read than are consumed.
func FromReader(ctx context.Context, r io.Reader) Stream[byte] {
	var buffer []byte
	var pending []byte // The bytes that have been read but not returned.
	finished := false

	return newStream(ctx, func(done <-chan struct{}) (byte, bool) {
		if cancelled(done) {
			return 0, false
		}
		for len(pending) == 0 {
			if finished {
				return 0, false
			}
			if buffer == nil {
				buffer = make([]byte, 4096) // Default page size.
			}
			n, err := r.Read(buffer)
			pending = buffer[:n]
			finished = err != nil
		}
		b := pending[0]
		pending = pending[1:]
		return b, true
	})
}

// Iota creates a stream that will return integers from start (inclusive) to
// end (exclusive), incremented by step. The stream ends when end is reached or
// when the sequence exceeds the type limits.
func Iota(ctx context.Context, start, end, step int) Stream[int] {
	i := start
	finished := step <= 0 || start >= end

	return newStream(ctx, func(done <-chan struct{}) (int, bool) {
		if finished || cancelled(done) {
			return 0, false
		}
		val := i
		// Stop if the next value would reach the end, before adding the step
		// can overflow. The distance is compared as unsigned because it can
		// exceed the int range.
		if uint(end)-uint(i) <= uint(step) {
			finished = true
		} else {
			i += step
		}
		return val, true
	})
}

// Fibonacci creates an integer stream returning the fibonacci sequence. This
// stream will not end by itself and should be limited using other methods.
func Fibonacci(ctx context.Context) Stream[*big.Int] {
	a := big.NewInt(0)
	b := big.NewInt(1)

	return newStream(ctx, func(done <-chan struct{}) (*big.Int, bool) {
		if cancelled(done) {
			return nil, false
		}
		val := big.NewInt(0).Set(a.Add(a, b))
		a, b = b, a
		return val, true
	})
}

// Primes creates an integer stream returning prime numbers. The stream ends
// when the sequence exceeds the type limits which may take a long time.
func Primes(ctx context.Context) Stream[int] {
	// The odd primes found so far and the next candidate. A candidate of one
	// means two hasn't been returned yet.
	primes := make([]int, 0, 64)
	n := 1

	return newStream(ctx, func(done <-chan struct{}) (int, bool) {
		if cancelled(done) {
			return 0, false
		}

		if n == 1 {
			n = 3
			return 2, true
		}

		for ; n > 0; n += 2 {
			isPrime := true

			for _, prime := range primes {
				// A composite number has a prime factor no larger than its
				// square root. The square is unsigned so it cannot overflow.
				if uint(prime)*uint(prime) > uint(n) {
					break
				}
				if n%prime == 0 {
					isPrime = false
					break
				}
			}

			if isPrime {
				prime := n
				primes = append(primes, prime)
				n += 2
				return prime, true
			}
		}

		return 0, false
	})
}

// RandInt creates an integer stream returning random integers. This stream
// will not end by itself and should be limited using other methods.
func RandInt(ctx context.Context) Stream[int] {
	return Generate(ctx, rand.Int)
}

// RandFloat32 creates a (32bit) float stream returning random floats. This
// stream will not end by itself and should be limited using other methods.
func RandFloat32(ctx context.Context) Stream[float32] {
	return Generate(ctx, rand.Float32)
}

// RandFloat64 creates a (64bit) float stream returning random floats. This
// stream will not end by itself and should be limited using other methods.
func RandFloat64(ctx context.Context) Stream[float64] {
	return Generate(ctx, rand.Float64)
}
