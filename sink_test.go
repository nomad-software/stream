package stream

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrain(t *testing.T) {
	c := FromString("Lorem ipsum dolor sit amet")

	assert.Equal(t, 'L', <-c)
	assert.Equal(t, 'o', <-c)

	c.Drain()

	assert.Equal(t, int32(0), <-c)
}

func TestSlice(t *testing.T) {
	expected := []int{1, 2, 3, 4}
	actual := Iota(1, 5, 1).Slice()

	assert.Equal(t, expected, actual)
}

func TestWriteToInt(t *testing.T) {
	expected := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0}

	buf := new(bytes.Buffer)
	err := Iota(1, 3, 1).WriteTo(buf)
	assert.NoError(t, err)

	actual := buf.Bytes()

	assert.Equal(t, expected, actual)
}

func TestWriteToString(t *testing.T) {
	expected := []byte{76, 0, 0, 0, 111, 0, 0, 0}

	buf := new(bytes.Buffer)
	err := FromString("Lorem ipsum dolor sit amet").Take(2).WriteTo(buf)
	assert.NoError(t, err)

	actual := buf.Bytes()

	assert.Equal(t, expected, actual)
}

func TestString(t *testing.T) {
	expected := "[1 2 3 4 5 6 7 8 9]"
	actual := Iota(1, 10, 1).String()

	assert.Equal(t, expected, actual)
}

func TestStringRunes(t *testing.T) {
	expected := "Lorem ipsum"
	actual := FromString("Lorem ipsum dolor sit amet").Take(11).String()

	assert.Equal(t, expected, actual)
}

func TestPop(t *testing.T) {
	expected := 2
	actual := Iota(2, 10, 2).Pop()

	assert.Equal(t, expected, actual)
}

func TestPrint(t *testing.T) {
	Iota(2, 10, 2).Print()
}
