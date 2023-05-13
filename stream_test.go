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

func TestZipVariadic1(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5fF6gG7hH8"

	a := FromString("0123456789")
	b := FromString("abcdefgh") // Stops the zip
	c := FromString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	actual := a.Zip(b, c).String()

	assert.Equal(t, expected, actual)
}

func TestZipVariadic2(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5f"

	a := FromString("0123456789")
	b := FromString("abcdefgh")
	c := FromString("ABCDE") // Stops the zip
	actual := a.Zip(b, c).String()

	assert.Equal(t, expected, actual)
}

func TestZip(t *testing.T) {
	expected := "0AaB1CbD2EcF3GdH4IeJ5KfL6MgN7OhP8QiR9SjT"

	a := FromString("0123456789") // Stops the zip
	b := FromString("abcdefghijklmnopqrstuvwxyz")
	c := FromString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	actual := a.Zip(b).Zip(c).String()

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
