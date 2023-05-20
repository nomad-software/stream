package stream

import "os"

func join(a, b string) string {
	return a + " " + b
}

func Example() {
	text := "Lorem adipiscing elit ipsum sed neque dolor non libero sit consequat magna amet placerat bibendum"

	FromString(text, " ").
		Stride(3).
		Take(2).
		Reduce(join).
		WriteTo(os.Stdout)

	// Output: Lorem ipsum
}
