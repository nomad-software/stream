# stream

**Generic stream processors written in Go**

---

## Description

This is a collection of generic, reusable, channel based stream processors that perform various operations on channels and their values. It's inspired by [component based programming](https://wiki.dlang.org/Component_programming_with_ranges) and [ranges](https://www.informit.com/articles/printerfriendly/1407357) popularised by the [D language](https://dlang.org/). Each operation is designed to be concurrent and if possible will execute in parallel. This is kind of an experiment to see how far I can leverage this. I've no idea if this is even useful.

## Documentation

https://pkg.go.dev/github.com/nomad-software/stream

## Example

```go
package main

import (
	"context"
	"os"

	"github.com/nomad-software/stream"
)

func join(a, b string) string {
	return a + " " + b
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	text := "Lorem adipiscing elit ipsum sed neque dolor non libero sit consequat magna amet placerat bibendum"

	stream.FromString(ctx, text, " ").
		Stride(3).
		Take(2).
		Reduce(join).
		Write(os.Stdout)

	// Output: Lorem ipsum
}
```

## Iterating

Use `All` to range over the values of a stream.

```go
for val := range stream.Iota(ctx, 0, 10, 1).All() {
	fmt.Println(val)
}
```

`Chunk`, `Zip` and `Enumerate` return plain channels, which can be ranged over
directly.
