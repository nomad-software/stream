package stream

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

// Drain drains the main channel of all values.
func (c Chan[T]) Drain() {
	for range c {
	}
}

// Slice returns a slice containing the channel values once the channel closes.
func (c Chan[T]) Slice() []T {
	output := make([]T, 0)

	for val := range c {
		output = append(output, val)
	}

	return output
}

// WriteTo writes the channel values as bytes to the writer argument.
func (c Chan[T]) WriteTo(w io.Writer) error {
	for v := range c {
		switch val := any(v).(type) {
		case int:
			err := binary.Write(w, binary.LittleEndian, int64(val))
			if err != nil {
				return err
			}
		case *int:
			err := binary.Write(w, binary.LittleEndian, uint64(uintptr(unsafe.Pointer(val))))
			if err != nil {
				return err
			}
		case uint:
			err := binary.Write(w, binary.LittleEndian, uint64(val))
			if err != nil {
				return err
			}
		case *uint:
			err := binary.Write(w, binary.LittleEndian, uint64(uintptr(unsafe.Pointer(val))))
			if err != nil {
				return err
			}
		case uintptr:
			err := binary.Write(w, binary.LittleEndian, uint64(val))
			if err != nil {
				return err
			}
		case string:
			err := binary.Write(w, binary.LittleEndian, []byte(val))
			if err != nil {
				return err
			}
		case *string:
			err := binary.Write(w, binary.LittleEndian, uint64(uintptr(unsafe.Pointer(val))))
			if err != nil {
				return err
			}
		default:
			err := binary.Write(w, binary.LittleEndian, val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// String returns the string representation of the channel values as a slice.
func (c Chan[T]) String() string {
	slice := c.Slice()

	switch val := any(slice).(type) {
	case []rune:
		return string(val)

	default:
		return fmt.Sprintf("%v", val)
	}

}

// Pop will return one value from the channel.
func (c Chan[T]) Pop() T {
	return <-c
}

// Print will print the string representation of the channel values to stdout.
// This is useful for debugging.
func (c Chan[T]) Print() {
	fmt.Println(c.String())
}
