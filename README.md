# stream

**Abstract stream processors written in Go**

---

## Example

```go
package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEuler1 solves https://projecteuler.net/problem=1
func TestEuler1(t *testing.T) {
	expected := 233168

	value := Iota(0, 1000, 1).
		Filter(func(n int) bool { return n%3 == 0 || n%5 == 0 }).
		Reduce(func(a, b int) int { return a + b }).
		Pop()

	assert.Equal(t, expected, value)
}

// TestEuler2 solves https://projecteuler.net/problem=2
func TestEuler2(t *testing.T) {
	expected := uint(4613732)

	value := Fibonacci().
		Until(func(n uint) bool { return n > 4000000 }).
		Filter(func(n uint) bool { return n%2 == 0 }).
		Reduce(func(a, b uint) uint { return a + b }).
		Pop()

	assert.Equal(t, expected, value)
}

```

## Documentation

https://pkg.go.dev/github.com/nomad-software/stream
