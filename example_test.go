package stream

import (
	"fmt"
)

func ExampleChan() {
	text := "Lorem adipiscing elit ipsum sed neque dolor non libero sit consequat magna amet placerat bibendum"

	result := FromString(text, " ").
		Stride(3).
		Take(2).
		Reduce(func(a, b string) string {
			return a + " " + b
		}).
		Pop()

	fmt.Println(result)
	// Output: Lorem ipsum
}
