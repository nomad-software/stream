package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTake(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	result := Iota(1, 10, 1).Take(5).Slice()

	assert.Equal(t, expected, result)

	empty := []rune{}
	c := FromRunes("")

	assert.Equal(t, empty, c.Slice())
	assert.Equal(t, empty, c.Slice())
	assert.Equal(t, empty, c.Slice())
}

func TestTakeNotEnough(t *testing.T) {
	expected := []int{1, 2}
	result := Iota(1, 3, 1).Take(5).Slice()

	assert.Equal(t, expected, result)
}

func TestUntil(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	result := Iota(1, 10, 1).Until(func(val int) bool { return val > 5 }).Slice()

	assert.Equal(t, expected, result)
}

func TestUntilNotEnough(t *testing.T) {
	expected := []int{1, 2}
	result := Iota(1, 3, 1).Until(func(val int) bool { return val > 5 }).Slice()

	assert.Equal(t, expected, result)
}

func TestMap(t *testing.T) {
	expected := "Yberz vcfhz qbybe fvg nzrg"
	result := FromRunes("Lorem ipsum dolor sit amet").Map(func(val rune) rune {
		if (val >= 'A' && val <= 'M') || (val >= 'a' && val <= 'm') {
			return val + 13
		} else if (val >= 'N' && val <= 'Z') || (val >= 'n' && val <= 'z') {
			return val - 13
		} else {
			return val
		}
	}).String()

	assert.Equal(t, expected, result)
}

func TestFilter(t *testing.T) {
	expected := []int{2, 4, 6, 8, 10}
	result := Iota(1, 12, 1).Filter(func(val int) bool { return val%2 == 0 }).Slice()

	assert.Equal(t, expected, result)
}

func TestReduce(t *testing.T) {
	expected := 45
	result := Iota(1, 10, 1).Reduce(func(a, b int) int { return a + b }).Pop()

	assert.Equal(t, expected, result)
}

func TestLast(t *testing.T) {
	expected := 9
	result := Iota(1, 10, 1).Last().Pop()

	assert.Equal(t, expected, result)
}

func TestChainVariadic(t *testing.T) {
	expected := "Lorem ipsum dolor sit amet"

	a := FromRunes("Lorem ipsum")
	b := FromRunes(" dolor")
	c := FromRunes(" sit amet")
	result := a.Chain(b, c).String()

	assert.Equal(t, expected, result)
}

func TestChain(t *testing.T) {
	expected := "Lorem ipsum dolor sit amet"

	a := FromRunes("Lorem ipsum")
	b := FromRunes(" dolor")
	c := FromRunes(" sit amet")
	result := a.Chain(b).Chain(c).String()

	assert.Equal(t, expected, result)
}

func TestRoundRobinVariadic1(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5fF6gG7hH8I9JKLMNOPQRSTUVWXYZ"

	a := FromRunes("0123456789")
	b := FromRunes("abcdefgh")
	c := FromRunes("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := a.RoundRobin(b, c).String()

	assert.Equal(t, expected, result)
}

func TestRoundRobinVariadic2(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5f6g7h89"

	a := FromRunes("0123456789")
	b := FromRunes("abcdefgh")
	c := FromRunes("ABCDE")
	result := a.RoundRobin(b, c).String()

	assert.Equal(t, expected, result)
}

func TestRoundRobin(t *testing.T) {
	expected := "0AaB1CbD2EcF3GdH4IeJ5KfL6MgN7OhP8QiR9SjTkUlVmWnXoYpZqrstuvwxyz"

	a := FromRunes("0123456789")
	b := FromRunes("abcdefghijklmnopqrstuvwxyz")
	c := FromRunes("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := a.RoundRobin(b).RoundRobin(c).String()

	assert.Equal(t, expected, result)
}

func TestChunk(t *testing.T) {
	expected := [][]int{{2, 4}, {6, 8}, {10}}
	i := 0
	for c := range Iota(2, 12, 2).Chunk(2) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestDrop(t *testing.T) {
	expected := []int{6, 7, 8, 9, 10}
	result := Iota(1, 20, 1).Drop(5).Take(5).Slice()

	assert.Equal(t, expected, result)

	expected = []int{6, 7, 8}
	empty := []int{}
	c := Iota(1, 10, 1).Take(8)

	assert.Equal(t, expected, c.Drop(5).Slice())
	assert.Equal(t, empty, c.Drop(5).Slice())
	assert.Equal(t, empty, c.Drop(5).Slice())
}

func TestStride(t *testing.T) {
	expected := []int{1, 4, 7, 10, 13, 16, 19, 22, 25, 28}
	result := Iota(1, 100, 1).Stride(3).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func TestTail(t *testing.T) {
	expected := []int{17, 18, 19}
	result := Iota(1, 20, 1).Tail(3).Slice()

	assert.Equal(t, expected, result)
}

func TestZipVariadic1(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A'},
		{'1', 'b', 'B'},
		{'2', 'c', 'C'},
		{'3', 'd', 'D'},
	}
	a := FromRunes("0123") // Stops the zip
	b := FromRunes("abcdefg")
	c := FromRunes("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	i := 0
	for c := range a.Zip(b, c) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestZipVariadic2(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A'},
		{'1', 'b', 'B'},
		{'2', 'c', 'C'},
		{'3', 'd', 'D'},
	}
	a := FromRunes("0123456789")
	b := FromRunes("abcd") // Stops the zip
	c := FromRunes("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	i := 0
	for c := range a.Zip(b, c) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestZipVariadic3(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A'},
		{'1', 'b', 'B'},
		{'2', 'c', 'C'},
		{'3', 'd', 'D'},
	}
	a := FromRunes("0123456789")
	b := FromRunes("abcdefghi")
	c := FromRunes("ABCD") // Stops the zip
	i := 0
	for c := range a.Zip(b, c) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestPadRight(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 0, 0, 0}
	result := Iota(1, 6, 1).PadRight(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestPadRightExceeded(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := Iota(1, 10, 1).PadRight(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestPadLeft(t *testing.T) {
	expected := []int{0, 0, 0, 1, 2, 3, 4, 5}
	result := Iota(1, 6, 1).PadLeft(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestPadLeftExceeded(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := Iota(1, 10, 1).PadLeft(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestTee(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	tee := 0
	result := Iota(1, 6, 1).Tee(func(val int) { tee = val }).Slice()

	assert.Equal(t, expected, result)
	assert.Equal(t, 5, tee)
}
