# stream

**Generic stream processors written in Go**

---

## Description

This is a collection of generic, reusable, channel based stream processors that perform various operations on channels and their values. It's inspired by [component based programming](https://wiki.dlang.org/Component_programming_with_ranges) and [ranges](https://www.informit.com/articles/printerfriendly/1407357) popularised by the [D language](https://dlang.org/). This is kind of an experiment to see how far I can leverage this. I've no idea if this is even useful.

## Example

```go
package main

import (
	"os"
)

func join(a, b string) string {
	return a + " " + b
}

func main() {
	text := "Lorem adipiscing elit ipsum sed neque dolor non libero sit consequat magna amet placerat bibendum"

	FromString(text, " ").
		Stride(3).
		Take(2).
		Reduce(join).
		WriteTo(os.Stdout)

	// Output: Lorem ipsum
}
```

## Documentation

https://pkg.go.dev/github.com/nomad-software/stream
