package stream

import (
	"fmt"
)

func ExampleChan() {
	text := "lorem adipiscing elit ipsum sed neque dolor non libero sit consequat magna amet placerat bibendum"

	output := FromString(text, " ").
		Stride(3).
		Take(2).
		Reduce(func(a, b string) string {
			return a + " " + b
		}).
		Pop()

	fmt.Println(output)
	// Output: lorem ipsum
}
