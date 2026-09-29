package stream

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Drain drains the main channel of all values and will return once the main
// channel is closed.
func (c Stream[T]) Drain() {
	for range c.All() {
	}
}

// Slice returns a slice containing the channel values once the main channel
// closes.
func (c Stream[T]) Slice() []T {
	output := make([]T, 0)

	for val := range c.All() {
		output = append(output, val)
	}

	return output
}

// Write writes the main channel values as bytes to the writer argument.
func (c Stream[T]) Write(w io.Writer) error {
	for v := range c.All() {
		switch val := any(v).(type) {
		case int:
			err := binary.Write(w, binary.LittleEndian, int64(val))
			if err != nil {
				return err
			}
		case *int:
			if val == nil {
				return fmt.Errorf("cannot write nil %T", val)
			}
			err := binary.Write(w, binary.LittleEndian, uint64(*val))
			if err != nil {
				return err
			}
		case uint:
			err := binary.Write(w, binary.LittleEndian, uint64(val))
			if err != nil {
				return err
			}
		case *uint:
			if val == nil {
				return fmt.Errorf("cannot write nil %T", val)
			}
			err := binary.Write(w, binary.LittleEndian, uint64(*val))
			if err != nil {
				return err
			}
		case string:
			err := binary.Write(w, binary.LittleEndian, []byte(val))
			if err != nil {
				return err
			}
		case *string:
			if val == nil {
				return fmt.Errorf("cannot write nil %T", val)
			}
			err := binary.Write(w, binary.LittleEndian, []byte(*val))
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

// ToString returns the string representation of the main channel values as a
// slice. A specialisation exists for rune channels, returning their string
// representation instead. This consumes the main channel.
func (c Stream[T]) ToString() string {
	slice := c.Slice()

	switch val := any(slice).(type) {
	case []rune:
		return string(val)

	default:
		return fmt.Sprintf("%v", val)
	}

}

// Pop will return one value from the main channel.
func (c Stream[T]) Pop() T {
	val, _ := c.recv()
	return val
}

// Print will output the string representation of the main channel values to
// stdout. This is useful for debugging.
func (c Stream[T]) Print() {
	fmt.Println(c.ToString())
}
