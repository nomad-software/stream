package stream

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"
	"testing/synctest"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestFromSlice(t *testing.T) {
	slice := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}

	expected := []string{"Lorem", "ipsum"}
	result := FromSlice(context.Background(), slice).Take(2).Slice()

	assert.Equal(t, expected, result)
}

func TestFromSliceWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		slice := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}
		stream := FromSlice(ctx, slice)

		assert.Equal(t, "Lorem", stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleFromSlice() {
	slice := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}

	result := FromSlice(context.Background(), slice).Take(2).Slice()

	fmt.Println(result)
	// Output: [Lorem ipsum]
}

func TestCycle(t *testing.T) {
	slice := []string{"Lorem", "ipsum"}

	expected := []string{"Lorem", "ipsum", "Lorem", "ipsum", "Lorem", "ipsum", "Lorem", "ipsum", "Lorem"}
	result := Cycle(context.Background(), slice).Take(9).Slice()

	assert.Equal(t, expected, result)
}

func TestCycleWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		slice := []string{"Lorem", "ipsum", "dolor"}
		stream := Cycle(ctx, slice)

		assert.Equal(t, "Lorem", stream.Pop())
		assert.Equal(t, "ipsum", stream.Pop())
		assert.Equal(t, "dolor", stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleCycle() {
	slice := []string{"Lorem", "ipsum"}

	result := Cycle(context.Background(), slice).Take(9).Slice()

	fmt.Println(result)
	// Output: [Lorem ipsum Lorem ipsum Lorem ipsum Lorem ipsum Lorem]
}

func TestGenerate(t *testing.T) {
	expected := []int{9, 18, 27, 36, 45, 54, 63, 72, 81, 90}
	i := 0
	result := Generate(context.Background(), func() int {
		i++
		return i * 9
	}).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func TestGenerateWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		i := 0
		stream := Generate(ctx, func() int {
			i++
			return i * 9
		})

		assert.Equal(t, []int{9, 18, 27, 36, 45}, stream.Take(5).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleGenerate() {
	i := 0
	result := Generate(context.Background(), func() int {
		i++
		return i * 9
	}).Take(10).Slice()

	fmt.Println(result)
	// Output: [9 18 27 36 45 54 63 72 81 90]
}

func TestRepeat(t *testing.T) {
	expected := []string{"Lorem", "Lorem", "Lorem", "Lorem", "Lorem"}
	result := Repeat(context.Background(), "Lorem").Take(5).Slice()

	assert.Equal(t, expected, result)
}

func TestRepeatWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, "Lorem")

		assert.Equal(t, []string{"Lorem", "Lorem", "Lorem", "Lorem", "Lorem"}, stream.Take(5).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleRepeat() {
	result := Repeat(context.Background(), "Lorem").Take(5).Slice()

	fmt.Println(result)
	// Output: [Lorem Lorem Lorem Lorem Lorem]
}

func TestFromChannel(t *testing.T) {
	c := make(chan int)

	go func() {
		defer close(c)
		for i := range 10 {
			c <- i
		}
	}()

	expected := []int{0, 1, 2, 3, 4}
	result := FromChannel(context.Background(), c).Take(5).Slice()

	assert.Equal(t, expected, result)
}

func TestFromChannelWithContext(t *testing.T) {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := range 10 {
			c <- i
		}
	}()

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := FromChannel(ctx, c)

		assert.Equal(t, []int{0, 1, 2, 3, 4}, stream.Take(5).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleFromChannel() {
	c := make(chan int)

	go func() {
		defer close(c)
		for i := range 10 {
			c <- i
		}
	}()

	result := FromChannel(context.Background(), c).Take(5).Slice()

	fmt.Println(result)
	// Output: [0 1 2 3 4]
}

func TestFromString(t *testing.T) {
	expected := []string{"Lorem", "ipsum"}
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Take(2).Slice()

	assert.Equal(t, expected, result)
}

func TestFromStringWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := FromString(ctx, "Lorem ipsum dolor sit amet", " ")

		assert.Equal(t, []string{"Lorem", "ipsum"}, stream.Take(2).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleFromString() {
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Take(2).Slice()

	fmt.Println(result)
	// Output: [Lorem ipsum]
}

func TestFromRunes(t *testing.T) {
	expected := []rune{'L', 'o', 'r', 'e', 'm', ' ', 'i', 'p', 's', 'u', 'm'}
	result := FromRunes(context.Background(), "Lorem ipsum dolor sit amet").Take(11).Slice()

	assert.Equal(t, expected, result)
}

func TestFromRunesWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := FromRunes(ctx, "Lorem ipsum dolor sit amet")

		assert.Equal(t, []rune{'L', 'o', 'r', 'e', 'm', ' ', 'i', 'p', 's', 'u', 'm'}, stream.Take(11).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleFromRunes() {
	result := FromRunes(context.Background(), "Hello, 世界 😊").Filter(unicode.IsSymbol).Slice()

	fmt.Println(string(result))
	// Output: 😊
}

func TestFromReader(t *testing.T) {
	expected := []byte("Lorem ipsum")
	r := bytes.NewBufferString("Lorem ipsum dolor sit amet")
	result := FromReader(context.Background(), r).Take(11).Slice()

	assert.Equal(t, expected, result)
}

func TestFromReaderWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		r := bytes.NewBufferString("Lorem ipsum dolor sit amet")
		stream := FromReader(ctx, r)

		assert.Equal(t, []byte("Lorem ipsum"), stream.Take(11).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleFromReader() {
	r := bytes.NewBufferString("Lorem ipsum dolor sit amet")
	result := FromReader(context.Background(), r).Take(11).Slice()

	fmt.Println(string(result))
	// Output: Lorem ipsum
}

func TestIota(t *testing.T) {
	expected := []int{-10, -8, -6, -4, -2, 0, 2, 4, 6, 8}
	result := Iota(context.Background(), -10, 10, 2).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func TestIotaWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Iota(ctx, -10, 10, 2)

		assert.Equal(t, []int{-10, -8, -6, -4, -2, 0, 2, 4, 6, 8}, stream.Take(10).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleIota() {
	result := Iota(context.Background(), -10, 10, 2).Take(10).Slice()

	fmt.Println(result)
	// Output: [-10 -8 -6 -4 -2 0 2 4 6 8]
}

func TestFibonacci(t *testing.T) {
	expected := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
		big.NewInt(3),
		big.NewInt(5),
		big.NewInt(8),
		big.NewInt(13),
		big.NewInt(21),
		big.NewInt(34),
		big.NewInt(55),
		big.NewInt(89),
	}
	result := Fibonacci(context.Background()).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func TestFibonacciWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Fibonacci(ctx)
		expected := []*big.Int{
			big.NewInt(1),
			big.NewInt(2),
			big.NewInt(3),
			big.NewInt(5),
			big.NewInt(8),
			big.NewInt(13),
			big.NewInt(21),
			big.NewInt(34),
			big.NewInt(55),
			big.NewInt(89),
		}

		assert.Equal(t, expected, stream.Take(10).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleFibonacci() {
	result := Fibonacci(context.Background()).Take(20).Slice()

	fmt.Println(result)
	// Output: [1 2 3 5 8 13 21 34 55 89 144 233 377 610 987 1597 2584 4181 6765 10946]
}

func TestPrimes(t *testing.T) {
	expected := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	result := Primes(context.Background()).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func TestPrimesWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Primes(ctx)

		assert.Equal(t, []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}, stream.Take(10).Slice())

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExamplePrimes() {
	result := Primes(context.Background()).Take(20).Slice()

	fmt.Println(result)
	// Output: [2 3 5 7 11 13 17 19 23 29 31 37 41 43 47 53 59 61 67 71]
}

func TestRandInt(t *testing.T) {
	expected := 10
	result := len(RandInt(context.Background()).Take(10).Slice())

	assert.Equal(t, expected, result)
}

func TestRandIntWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := RandInt(ctx)

		assert.Len(t, stream.Take(10).Slice(), 10)

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleRandInt() {
	result := RandInt(context.Background()).Take(10).Slice()

	fmt.Println(len(result))
	// Output: 10
}

func TestRandFloat32(t *testing.T) {
	expected := 10
	result := len(RandFloat32(context.Background()).Take(10).Slice())

	assert.Equal(t, expected, result)
}

func TestRandFloat32WithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := RandFloat32(ctx)

		assert.Len(t, stream.Take(10).Slice(), 10)

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleRandFloat32() {
	result := RandFloat32(context.Background()).Take(10).Slice()

	fmt.Println(len(result))
	// Output: 10
}

func TestRandFloat64(t *testing.T) {
	expected := 10
	result := len(RandFloat64(context.Background()).Take(10).Slice())

	assert.Equal(t, expected, result)
}

func TestRandFloat64WithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := RandFloat64(ctx)

		assert.Len(t, stream.Take(10).Slice(), 10)

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleRandFloat64() {
	result := RandFloat64(context.Background()).Take(10).Slice()

	fmt.Println(len(result))
	// Output: 10
}
