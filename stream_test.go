package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTake(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	actual := Iota(1, 10, 1).Take(5).Slice()

	assert.Equal(t, expected, actual)

	empty := []rune{}
	c := FromString("")

	assert.Equal(t, empty, c.Slice())
	assert.Equal(t, empty, c.Slice())
	assert.Equal(t, empty, c.Slice())
}

func TestTakeNotEnough(t *testing.T) {
	expected := []int{1, 2}
	actual := Iota(1, 3, 1).Take(5).Slice()

	assert.Equal(t, expected, actual)
}

func TestUntil(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	actual := Iota(1, 10, 1).Until(func(val int) bool { return val > 5 }).Slice()

	assert.Equal(t, expected, actual)
}

func TestUntilNotEnough(t *testing.T) {
	expected := []int{1, 2}
	actual := Iota(1, 3, 1).Until(func(val int) bool { return val > 5 }).Slice()

	assert.Equal(t, expected, actual)
}

func TestMap(t *testing.T) {
	expected := "Yberz vcfhz qbybe fvg nzrg"
	actual := FromString("Lorem ipsum dolor sit amet").Map(func(val rune) rune {
		if (val >= 'A' && val <= 'M') || (val >= 'a' && val <= 'm') {
			return val + 13
		} else if (val >= 'N' && val <= 'Z') || (val >= 'n' && val <= 'z') {
			return val - 13
		} else {
			return val
		}
	}).String()

	assert.Equal(t, expected, actual)
}

func TestFilter(t *testing.T) {
	expected := []int{2, 4, 6, 8, 10}
	actual := Iota(1, 12, 1).Filter(func(val int) bool { return val%2 == 0 }).Slice()

	assert.Equal(t, expected, actual)
}

func TestReduce(t *testing.T) {
	expected := 45
	actual := Iota(1, 10, 1).Reduce(func(a, b int) int { return a + b }).Pop()

	assert.Equal(t, expected, actual)
}

func TestLast(t *testing.T) {
	expected := 9
	actual := Iota(1, 10, 1).Last().Pop()

	assert.Equal(t, expected, actual)
}

func TestChainVariadic(t *testing.T) {
	expected := "Lorem ipsum dolor sit amet"

	a := FromString("Lorem ipsum")
	b := FromString(" dolor")
	c := FromString(" sit amet")
	actual := a.Chain(b, c).String()

	assert.Equal(t, expected, actual)
}

func TestChain(t *testing.T) {
	expected := "Lorem ipsum dolor sit amet"

	a := FromString("Lorem ipsum")
	b := FromString(" dolor")
	c := FromString(" sit amet")
	actual := a.Chain(b).Chain(c).String()

	assert.Equal(t, expected, actual)
}

func TestRoundRobinVariadic1(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5fF6gG7hH8I9JKLMNOPQRSTUVWXYZ"

	a := FromString("0123456789")
	b := FromString("abcdefgh")
	c := FromString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	actual := a.RoundRobin(b, c).String()

	assert.Equal(t, expected, actual)
}

func TestRoundRobinVariadic2(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5f6g7h89"

	a := FromString("0123456789")
	b := FromString("abcdefgh")
	c := FromString("ABCDE")
	actual := a.RoundRobin(b, c).String()

	assert.Equal(t, expected, actual)
}

func TestRoundRobin(t *testing.T) {
	expected := "0AaB1CbD2EcF3GdH4IeJ5KfL6MgN7OhP8QiR9SjTkUlVmWnXoYpZqrstuvwxyz"

	a := FromString("0123456789")
	b := FromString("abcdefghijklmnopqrstuvwxyz")
	c := FromString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	actual := a.RoundRobin(b).RoundRobin(c).String()

	assert.Equal(t, expected, actual)
}

func TestChunk(t *testing.T) {
	expected := [][]int{{2, 4}, {6, 8}, {10}}
	i := 0
	for c := range Iota(2, 12, 2).Chunk(2) {
		actual := c.Slice()
		assert.Equal(t, expected[i], actual)
		i++
	}
}

func TestDrop(t *testing.T) {
	expected := []int{6, 7, 8, 9, 10}
	actual := Iota(1, 20, 1).Drop(5).Take(5).Slice()

	assert.Equal(t, expected, actual)

	expected = []int{6, 7, 8}
	empty := []int{}
	c := Iota(1, 10, 1).Take(8)

	assert.Equal(t, expected, c.Drop(5).Slice())
	assert.Equal(t, empty, c.Drop(5).Slice())
	assert.Equal(t, empty, c.Drop(5).Slice())
}

func TestStride(t *testing.T) {
	expected := []int{1, 4, 7, 10, 13, 16, 19, 22, 25, 28}
	actual := Iota(1, 100, 1).Stride(3).Take(10).Slice()

	assert.Equal(t, expected, actual)
}

func TestTail(t *testing.T) {
	expected := []int{17, 18, 19}
	actual := Iota(1, 20, 1).Tail(3).Slice()

	assert.Equal(t, expected, actual)
}

func TestZipVariadic1(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A'},
		{'1', 'b', 'B'},
		{'2', 'c', 'C'},
		{'3', 'd', 'D'},
	}
	a := FromString("0123") // Stops the zip
	b := FromString("abcdefg")
	c := FromString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	i := 0
	for c := range a.Zip(b, c) {
		actual := c.Slice()
		assert.Equal(t, expected[i], actual)
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
	a := FromString("0123456789")
	b := FromString("abcd") // Stops the zip
	c := FromString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	i := 0
	for c := range a.Zip(b, c) {
		actual := c.Slice()
		assert.Equal(t, expected[i], actual)
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
	a := FromString("0123456789")
	b := FromString("abcdefghi")
	c := FromString("ABCD") // Stops the zip
	i := 0
	for c := range a.Zip(b, c) {
		actual := c.Slice()
		assert.Equal(t, expected[i], actual)
		i++
	}
}

func TestPadRight(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 0, 0, 0}
	actual := Iota(1, 6, 1).PadRight(0, 8).Slice()

	assert.Equal(t, expected, actual)
}

func TestPadRightExceeded(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	actual := Iota(1, 10, 1).PadRight(0, 8).Slice()

	assert.Equal(t, expected, actual)
}

func TestPadLeft(t *testing.T) {
	expected := []int{0, 0, 0, 1, 2, 3, 4, 5}
	actual := Iota(1, 6, 1).PadLeft(0, 8).Slice()

	assert.Equal(t, expected, actual)
}

func TestPadLeftExceeded(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	actual := Iota(1, 10, 1).PadLeft(0, 8).Slice()

	assert.Equal(t, expected, actual)
}

func TestTee(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	tee := 0
	actual := Iota(1, 6, 1).Tee(func(val int) { tee = val }).Slice()

	assert.Equal(t, expected, actual)
	assert.Equal(t, 5, tee)
}
