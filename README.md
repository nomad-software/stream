# stream

**Generic stream processors written in Go**

---

## Description

This is a collection of generic, reusable, iterator based stream processors that perform various operations on sequences of values. It's inspired by [component based programming](https://wiki.dlang.org/Component_programming_with_ranges) and [ranges](https://www.informit.com/articles/printerfriendly/1407357) popularised by the [D language](https://dlang.org/).

Streams are lazy and single pass. Nothing happens until a value is asked for, and then every stage runs on the calling goroutine, one value at a time. Each value is handed out once, like a channel, so a stream carries on from wherever it was left. Cancelling a stream's context stops it.

## Installation

```
go get github.com/nomad-software/stream/v2
```

## Documentation

https://pkg.go.dev/github.com/nomad-software/stream/v2

## Example

```go
package main

import (
	"context"
	"os"

	"github.com/nomad-software/stream/v2"
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

`Chunk`, `Zip` and `Enumerate` return iterators, which can be ranged over
directly.

Taking a value from a stream moves it on, and whatever uses the stream next
carries on from there.

```go
s := stream.Iota(ctx, 1, 6, 1)

s.Pop()           // 1
s.Take(2).Slice() // [2 3]
s.Slice()         // [4 5]
s.Slice()         // []
```

A stream must not be used by more than one goroutine at a time.

## Concurrency

Stages are not concurrent unless asked to be.

- `MapParallel` calls its function from a number of workers.
- `Merge` runs each of its streams in its own goroutine.
- `Buffer` runs the stages before it in their own goroutine, letting them get
  ahead of the stages after it.

The goroutines these start run until their streams are exhausted or the context
is cancelled, so cancel the context if a stream is abandoned part way through.

```go
stream.FromReader(ctx, file).
	Map(decode).
	Buffer(64). // Reading and decoding happen on another goroutine.
	MapParallel(8, process).
	Drain()
```
