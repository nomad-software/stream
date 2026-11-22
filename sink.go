package stream

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
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
	write := writer[T](w)

	for v := range c.All() {
		if err := write(v); err != nil {
			return err
		}
	}
	return nil
}

// writer returns a function that writes a value of type T to the passed writer
// in little endian byte order. The encoding is chosen once, from the type,
// rather than by inspecting every value.
func writer[T comparable](w io.Writer) func(T) error {
	var write func(T) error

	// One buffer shared by every write. Writers must not retain it.
	buf := make([]byte, 8)
	le := binary.LittleEndian

	put16 := func(v uint16) error { le.PutUint16(buf, v); _, err := w.Write(buf[:2]); return err }
	put32 := func(v uint32) error { le.PutUint32(buf, v); _, err := w.Write(buf[:4]); return err }
	put64 := func(v uint64) error { le.PutUint64(buf, v); _, err := w.Write(buf[:8]); return err }

	// The address of a func(T) can only be asserted to the address of the
	// function type that T makes it, which selects the encoding without
	// boxing any values.
	switch f := any(&write).(type) {
	case *func(int) error:
		*f = func(v int) error { return put64(uint64(v)) }
	case *func(*int) error:
		*f = func(v *int) error {
			if v == nil {
				return fmt.Errorf("cannot write nil %T", v)
			}
			return put64(uint64(*v))
		}
	case *func(uint) error:
		*f = func(v uint) error { return put64(uint64(v)) }
	case *func(*uint) error:
		*f = func(v *uint) error {
			if v == nil {
				return fmt.Errorf("cannot write nil %T", v)
			}
			return put64(uint64(*v))
		}
	case *func(string) error:
		*f = func(v string) error { _, err := io.WriteString(w, v); return err }
	case *func(*string) error:
		*f = func(v *string) error {
			if v == nil {
				return fmt.Errorf("cannot write nil %T", v)
			}
			_, err := io.WriteString(w, *v)
			return err
		}
	case *func(uint8) error:
		*f = func(v uint8) error { buf[0] = v; _, err := w.Write(buf[:1]); return err }
	case *func(int8) error:
		*f = func(v int8) error { buf[0] = uint8(v); _, err := w.Write(buf[:1]); return err }
	case *func(uint16) error:
		*f = put16
	case *func(int16) error:
		*f = func(v int16) error { return put16(uint16(v)) }
	case *func(uint32) error:
		*f = put32
	case *func(int32) error:
		*f = func(v int32) error { return put32(uint32(v)) }
	case *func(uint64) error:
		*f = put64
	case *func(int64) error:
		*f = func(v int64) error { return put64(uint64(v)) }
	case *func(float32) error:
		*f = func(v float32) error { return put32(math.Float32bits(v)) }
	case *func(float64) error:
		*f = func(v float64) error { return put64(math.Float64bits(v)) }
	default:
		// Anything else of a fixed size, such as an array or a struct.
		write = func(v T) error { return binary.Write(w, le, v) }
	}

	return write
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
