package stream

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEuler1(t *testing.T) {
	expected := 233168
	result := Iota(0, 1000, 1).
		Filter(func(n int) bool { return n%3 == 0 || n%5 == 0 }).
		Reduce(func(a, b int) int { return a + b }).
		Pop()

	assert.Equal(t, expected, result)
}

func TestEuler2(t *testing.T) {
	expected := big.NewInt(4613732)
	result := Fibonacci().
		Until(func(n *big.Int) bool { return n.Cmp(big.NewInt(4000000)) > 0 }).
		Filter(func(n *big.Int) bool { return big.NewInt(0).Mod(n, big.NewInt(2)).Cmp(big.NewInt(0)) == 0 }).
		Reduce(func(a, b *big.Int) *big.Int { return big.NewInt(0).Set(a.Add(a, b)) }).
		Pop()

	assert.Equal(t, expected, result)
}
