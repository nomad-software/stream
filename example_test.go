package stream

import (
	"fmt"
)

func join(a, b string) string {
	return a + " " + b
}

func Example() {
	text := "Lorem adipiscing elit ipsum sed neque dolor non libero sit consequat magna amet placerat bibendum"

	result := FromString(text, " ").
		Stride(3).
		Take(2).
		Reduce(join).
		Pop()

	fmt.Println(result)
	// Output: Lorem ipsum
}
