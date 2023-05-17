package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromSlice(t *testing.T) {
	slice := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}

	expected := []string{"Lorem", "ipsum"}
	actual := FromSlice(slice).Take(2).Slice()

	assert.Equal(t, expected, actual)
}

func TestCycle(t *testing.T) {
	slice := []string{"Lorem", "ipsum"}

	expected := []string{"Lorem", "ipsum", "Lorem", "ipsum", "Lorem", "ipsum", "Lorem", "ipsum", "Lorem"}
	actual := Cycle(slice).Take(9).Slice()

	assert.Equal(t, expected, actual)
}

func TestGenerate(t *testing.T) {
	expected := []int{9, 18, 27, 36, 45, 54, 63, 72, 81, 90}
	i := 0
	actual := Generate(func() int {
		i++
		return i * 9
	}).Take(10).Slice()

	assert.Equal(t, expected, actual)
}

func TestRepeat(t *testing.T) {
	expected := []string{"Lorem", "Lorem", "Lorem", "Lorem", "Lorem"}
	actual := Repeat("Lorem").Take(5).Slice()

	assert.Equal(t, expected, actual)
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
	actual := FromChannel(c).Take(5).Slice()

	assert.Equal(t, expected, actual)
}

func TestFromString(t *testing.T) {
	expected := []string{"Lorem", "ipsum"}
	actual := FromString("Lorem ipsum dolor sit amet", " ").Take(2).Slice()

	assert.Equal(t, expected, actual)
}

func TestFromRunes(t *testing.T) {
	expected := []rune{'L', 'o', 'r', 'e', 'm', ' ', 'i', 'p', 's', 'u'}
	actual := FromRunes("Lorem ipsum dolor sit amet").Take(10).Slice()

	assert.Equal(t, expected, actual)
}

func TestIota(t *testing.T) {
	expected := []int{-10, -8, -6, -4, -2, 0, 2, 4, 6, 8}
	actual := Iota(-10, 10, 2).Take(10).Slice()

	assert.Equal(t, expected, actual)
}

func TestFibonacci(t *testing.T) {
	expected := []uint{1, 2, 3, 5, 8, 13, 21, 34, 55, 89}
	actual := Fibonacci().Take(10).Slice()

	assert.Equal(t, expected, actual)
}

func TestFibonacciMaximum(t *testing.T) {
	expected := uint(12200160415121876738)

	actual := Fibonacci().Last().Pop()
	assert.Equal(t, expected, actual)
}

func TestPrimes(t *testing.T) {
	expected := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	actual := Primes().Take(10).Slice()

	assert.Equal(t, expected, actual)
}

func TestRandInt(t *testing.T) {
	expected := 10
	actual := len(RandInt().Take(10).Slice())

	assert.Equal(t, expected, actual)
}

func TestRandFloat32(t *testing.T) {
	expected := 10
	actual := len(RandFloat32().Take(10).Slice())

	assert.Equal(t, expected, actual)
}

func TestRandFloat64(t *testing.T) {
	expected := 10
	actual := len(RandFloat64().Take(10).Slice())

	assert.Equal(t, expected, actual)
}
