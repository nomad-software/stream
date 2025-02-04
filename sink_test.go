package stream

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrain(t *testing.T) {
	c := FromRunes("Lorem ipsum dolor sit amet")

	assert.Equal(t, 'L', <-c)
	assert.Equal(t, 'o', <-c)

	c.Drain()

	assert.Equal(t, int32(0), <-c)
}

func ExampleChan_Drain() {
	c := FromRunes("Lorem ipsum dolor sit amet")

	c.Drain()

	fmt.Println(<-c)
	// Output: 0
}

func TestSlice(t *testing.T) {
	expected := []int{1, 2, 3, 4}
	result := Iota(1, 5, 1).Slice()

	assert.Equal(t, expected, result)
}

func ExampleChan_Slice() {
	result := Iota(1, 5, 1).Slice()

	fmt.Println(result)
	// Output: [1 2 3 4]
}

func TestWriteToInt(t *testing.T) {
	expected := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0}

	buf := new(bytes.Buffer)
	err := Iota(1, 3, 1).WriteTo(buf)
	assert.NoError(t, err)

	result := buf.Bytes()

	assert.Equal(t, expected, result)
}

func TestWriteToIntPointer(t *testing.T) {
	expected := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0}

	buf := new(bytes.Buffer)
	one := 1
	two := 2
	err := FromSlice([]*int{&one, &two}).WriteTo(buf)
	assert.NoError(t, err)

	result := buf.Bytes()

	assert.Equal(t, expected, result)
}

func TestWriteToUint(t *testing.T) {
	expected := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0}

	buf := new(bytes.Buffer)
	err := FromSlice([]uint{1, 2}).WriteTo(buf)
	assert.NoError(t, err)

	result := buf.Bytes()

	assert.Equal(t, expected, result)
}

func TestWriteToUintPointer(t *testing.T) {
	expected := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0}

	buf := new(bytes.Buffer)
	one := uint(1)
	two := uint(2)
	err := FromSlice([]*uint{&one, &two}).WriteTo(buf)
	assert.NoError(t, err)

	result := buf.Bytes()

	assert.Equal(t, expected, result)
}

func TestWriteToString(t *testing.T) {
	expected := []byte{76, 111, 114, 101, 109}

	buf := new(bytes.Buffer)
	err := FromString("Lorem", "").WriteTo(buf)
	assert.NoError(t, err)

	result := buf.Bytes()

	assert.Equal(t, expected, result)
}

func TestWriteToStringPointer(t *testing.T) {
	expected := []byte{76, 111, 114, 101, 109}

	buf := new(bytes.Buffer)
	str := "Lorem"
	err := FromSlice([]*string{&str}).WriteTo(buf)
	assert.NoError(t, err)

	result := buf.Bytes()

	assert.Equal(t, expected, result)
}

func TestWriteToRunes(t *testing.T) {
	expected := []byte{76, 0, 0, 0, 111, 0, 0, 0}

	buf := new(bytes.Buffer)
	err := FromRunes("Lorem ipsum dolor sit amet").Take(2).WriteTo(buf)
	assert.NoError(t, err)

	result := buf.Bytes()

	assert.Equal(t, expected, result)
}

func ExampleChan_WriteTo() {
	buf := new(bytes.Buffer)
	FromRunes("Lorem ipsum dolor sit amet").Take(2).WriteTo(buf)

	fmt.Println(buf.Bytes())
	// Output: [76 0 0 0 111 0 0 0]
}

func TestString(t *testing.T) {
	expected := "[1 2 3 4 5 6 7 8 9]"
	result := Iota(1, 10, 1).String()

	assert.Equal(t, expected, result)
}

func ExampleChan_String() {
	result := Iota(1, 10, 1).String()

	fmt.Println(result)
	// Output: [1 2 3 4 5 6 7 8 9]
}

func TestStringRunes(t *testing.T) {
	expected := "Lorem ipsum"
	result := FromRunes("Lorem ipsum dolor sit amet").Take(11).String()

	assert.Equal(t, expected, result)
}

func TestPop(t *testing.T) {
	expected := 2
	result := Iota(2, 10, 2).Pop()

	assert.Equal(t, expected, result)
}

func ExampleChan_Pop() {
	result := Iota(2, 20, 2).Drop(6).Pop()

	fmt.Println(result)
	// Output: 14
}

func TestPrint(t *testing.T) {
	Iota(2, 10, 2).Print()
}

func ExampleChan_Print() {
	Primes().Take(10).Print()
	// Output: [2 3 5 7 11 13 17 19 23 29]
}
