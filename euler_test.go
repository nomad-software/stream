package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEuler1(t *testing.T) {
	expected := 233168
	actual := Iota(0, 1000, 1).
		Filter(func(n int) bool { return n%3 == 0 || n%5 == 0 }).
		Reduce(func(a, b int) int { return a + b }).
		Pop()

	assert.Equal(t, expected, actual)
}

func TestEuler2(t *testing.T) {
	expected := uint(4613732)
	actual := Fibonacci().
		Until(func(n uint) bool { return n > 4000000 }).
		Filter(func(n uint) bool { return n%2 == 0 }).
		Reduce(func(a, b uint) uint { return a + b }).
		Pop()

	assert.Equal(t, expected, actual)
}
