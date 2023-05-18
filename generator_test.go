package stream

import (
	"fmt"
	"math/big"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestFromSlice(t *testing.T) {
	slice := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}

	expected := []string{"Lorem", "ipsum"}
	result := FromSlice(slice).Take(2).Slice()

	assert.Equal(t, expected, result)
}

func ExampleFromSlice() {
	slice := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}

	result := FromSlice(slice).Take(2).Slice()

	fmt.Println(result)
	// Output: [Lorem ipsum]
}

func TestCycle(t *testing.T) {
	slice := []string{"Lorem", "ipsum"}

	expected := []string{"Lorem", "ipsum", "Lorem", "ipsum", "Lorem", "ipsum", "Lorem", "ipsum", "Lorem"}
	result := Cycle(slice).Take(9).Slice()

	assert.Equal(t, expected, result)
}

func ExampleCycle() {
	slice := []string{"Lorem", "ipsum"}

	result := Cycle(slice).Take(9).Slice()

	fmt.Println(result)
	// Output: [Lorem ipsum Lorem ipsum Lorem ipsum Lorem ipsum Lorem]
}

func TestGenerate(t *testing.T) {
	expected := []int{9, 18, 27, 36, 45, 54, 63, 72, 81, 90}
	i := 0
	result := Generate(func() int {
		i++
		return i * 9
	}).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func ExampleGenerate() {
	i := 0
	result := Generate(func() int {
		i++
		return i * 9
	}).Take(10).Slice()

	fmt.Println(result)
	// Output: [9 18 27 36 45 54 63 72 81 90]
}

func TestRepeat(t *testing.T) {
	expected := []string{"Lorem", "Lorem", "Lorem", "Lorem", "Lorem"}
	result := Repeat("Lorem").Take(5).Slice()

	assert.Equal(t, expected, result)
}

func ExampleRepeat() {
	result := Repeat("Lorem").Take(5).Slice()

	fmt.Println(result)
	// Output: [Lorem Lorem Lorem Lorem Lorem]
}

func TestFromChannel(t *testing.T) {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := 0; i < 10; i++ {
			c <- i
		}
	}()

	expected := []int{0, 1, 2, 3, 4}
	result := FromChannel(c).Take(5).Slice()

	assert.Equal(t, expected, result)
}

func ExampleFromChannel() {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := 0; i < 10; i++ {
			c <- i
		}
	}()

	result := FromChannel(c).Take(5).Slice()

	fmt.Println(result)
	// Output: [0 1 2 3 4]
}

func TestFromString(t *testing.T) {
	expected := []string{"Lorem", "ipsum"}
	result := FromString("Lorem ipsum dolor sit amet", " ").Take(2).Slice()

	assert.Equal(t, expected, result)
}

func ExampleFromString() {
	result := FromString("Lorem ipsum dolor sit amet", " ").Take(2).Slice()

	fmt.Println(result)
	// Output: [Lorem ipsum]
}

func TestFromRunes(t *testing.T) {
	expected := []rune{'L', 'o', 'r', 'e', 'm', ' ', 'i', 'p', 's', 'u'}
	result := FromRunes("Lorem ipsum dolor sit amet").Take(10).Slice()

	assert.Equal(t, expected, result)
}

func ExampleFromRunes() {
	result := FromRunes("Hello, 世界 😊").Filter(unicode.IsSymbol).Slice()

	fmt.Println(string(result))
	// Output: 😊
}

func TestIota(t *testing.T) {
	expected := []int{-10, -8, -6, -4, -2, 0, 2, 4, 6, 8}
	result := Iota(-10, 10, 2).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func ExampleIota() {
	result := Iota(-10, 10, 2).Take(10).Slice()

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
	result := Fibonacci().Take(10).Slice()

	assert.Equal(t, expected, result)
}

func ExampleFibonacci() {
	result := Fibonacci().Take(20).Slice()

	fmt.Println(result)
	// Output: [1 2 3 5 8 13 21 34 55 89 144 233 377 610 987 1597 2584 4181 6765 10946]
}

func TestPrimes(t *testing.T) {
	expected := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	result := Primes().Take(10).Slice()

	assert.Equal(t, expected, result)
}

func ExamplePrimes() {
	result := Primes().Take(20).Slice()

	fmt.Println(result)
	// Output: [2 3 5 7 11 13 17 19 23 29 31 37 41 43 47 53 59 61 67 71]
}

func TestRandInt(t *testing.T) {
	expected := 10
	result := len(RandInt().Take(10).Slice())

	assert.Equal(t, expected, result)
}

func ExampleRandInt() {
	result := RandInt().Take(10).Slice()

	fmt.Println(len(result))
	// Output: 10
}

func TestRandFloat32(t *testing.T) {
	expected := 10
	result := len(RandFloat32().Take(10).Slice())

	assert.Equal(t, expected, result)
}

func ExampleRandFloat32() {
	result := RandFloat32().Take(10).Slice()

	fmt.Println(len(result))
	// Output: 10
}

func TestRandFloat64(t *testing.T) {
	expected := 10
	result := len(RandFloat64().Take(10).Slice())

	assert.Equal(t, expected, result)
}

func ExampleRandFloat64() {
	result := RandFloat64().Take(10).Slice()

	fmt.Println(len(result))
	// Output: 10
}
