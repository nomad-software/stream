package stream

import (
	"bytes"
	"context"
	"encoding/binary"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

// Nothing runs until a value is asked for, and then only as much as is needed.
func TestLazy(t *testing.T) {
	calls := 0
	stream := Generate(context.Background(), func() int { calls++; return calls }).
		Map(func(n int) int { return n * 2 }).
		Take(3)

	assert.Equal(t, 0, calls)
	assert.Equal(t, []int{2, 4, 6}, stream.Slice())
	assert.Equal(t, 3, calls)
}

// Each value is handed out once, to whichever stage or sink asks for it next.
func TestSinglePass(t *testing.T) {
	ctx := context.Background()

	stream := Iota(ctx, 1, 11, 1)
	assert.Equal(t, 1, stream.Pop())
	assert.Equal(t, 2, stream.Pop())
	assert.Equal(t, []int{3, 4}, stream.Take(2).Slice())
	for val := range stream.All() {
		assert.Equal(t, 5, val)
		break
	}
	assert.Equal(t, 6, stream.Map(func(n int) int { return n }).Pop())
	for enum := range stream.Enumerate(0) {
		assert.Equal(t, Enum[int]{Index: 0, Val: 7}, enum)
		break
	}
	for chunk := range stream.Chunk(2) {
		assert.Equal(t, []int{8, 9}, chunk.Slice())
		break
	}
	assert.Equal(t, []int{10}, stream.Slice())
	assert.Equal(t, []int{}, stream.Slice())
	assert.Equal(t, 0, stream.Pop())

	// Copies of a stream share its position.
	a := FromString(ctx, "a b c d", " ")
	b := a
	assert.Equal(t, "a", a.Pop())
	assert.Equal(t, "b", b.Pop())
	assert.Equal(t, []string{"c", "d"}, a.Slice())
	assert.Equal(t, []string{}, b.Slice())

	// A stage keeps its own position too.
	evens := Iota(ctx, 0, 10, 1).Stride(2)
	assert.Equal(t, 0, evens.Pop())
	assert.Equal(t, 2, evens.Pop())
	assert.Equal(t, []int{4, 6, 8}, evens.Slice())

	reader := FromReader(ctx, bytes.NewReader([]byte("abcdef")))
	assert.Equal(t, byte('a'), reader.Pop())
	assert.Equal(t, []byte("bc"), reader.Take(2).Slice())
	assert.Equal(t, []byte("def"), reader.Slice())

	sum := Iota(ctx, 1, 5, 1).Reduce(func(a, b int) int { return a + b })
	assert.Equal(t, 10, sum.Pop())
	assert.Equal(t, 0, sum.Pop())

	tail := Iota(ctx, 1, 10, 1).Tail(3)
	assert.Equal(t, 7, tail.Pop())
	assert.Equal(t, []int{8, 9}, tail.Slice())

	padded := Iota(ctx, 1, 3, 1).PadLeft(0, 4)
	assert.Equal(t, 0, padded.Pop())
	assert.Equal(t, []int{0, 1, 2}, padded.Slice())
}

// The goroutines started by the concurrent stages must stop when the context
// is cancelled, even if their values are never asked for again. The test fails
// if any are still running when it ends.
func TestConcurrentStagesStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		streams := []Stream[int]{
			Repeat(ctx, 1).Buffer(4),
			Repeat(ctx, 1).MapParallel(3, func(n int) int { return n }),
			Repeat(ctx, 1).Merge(Repeat(ctx, 1), Repeat(ctx, 1)),
		}

		for _, stream := range streams {
			assert.Equal(t, []int{1, 1, 1, 1, 1}, stream.Take(5).Slice())
		}

		cancel()
	})
}

// The goroutines started by the concurrent stages must stop when their streams
// are exhausted. The test fails if any are still running when it ends.
func TestConcurrentStagesFinish(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()

		assert.Len(t, Iota(ctx, 0, 50, 1).Buffer(4).Slice(), 50)
		assert.Len(t, Iota(ctx, 0, 50, 1).MapParallel(3, func(n int) int { return n }).Slice(), 50)
		assert.Len(t, Iota(ctx, 0, 50, 1).Merge(Iota(ctx, 0, 50, 1), Iota(ctx, 0, 50, 1)).Slice(), 150)
	})
}

// A consumer must stop when the context is cancelled, even if a concurrent
// stage is stuck in a callback.
func TestCancelWhileCallbackIsStuck(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		release := make(chan struct{})
		stuck := func(n int) int { <-release; return n }

		time.AfterFunc(time.Second, cancel)

		assert.Empty(t, Repeat(ctx, 1).MapParallel(2, stuck).Slice())
		assert.Empty(t, Repeat(ctx, 1).Map(stuck).Buffer(2).Slice())
		assert.Empty(t, Repeat(ctx, 1).Map(stuck).Merge(Repeat(ctx, 1).Map(stuck)).Slice())

		close(release)
	})
}

// Streams bound to another context are consumed using the main stream's
// context.
func TestForeignStreamsUseMainContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		other := context.Background()
		blocked := make(chan int)

		time.AfterFunc(time.Second, cancel)

		assert.Equal(t, []int{1}, FromSlice(ctx, []int{1}).Chain(FromChannel(other, blocked)).Slice())
		assert.Empty(t, FromSlice(ctx, []int{1}).Merge(FromChannel(other, blocked)).Slice())
		assert.Empty(t, FromSlice(ctx, []int{1}).RoundRobin(FromChannel(other, blocked)).Slice())
		assert.Empty(t, slices.Collect(FromSlice(ctx, []int{1}).Zip(FromChannel(other, blocked))))
	})
}

func TestMapParallelRunsConcurrently(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var running, peak atomic.Int32

		start := time.Now()
		result := Iota(context.Background(), 0, 8, 1).MapParallel(4, func(n int) int {
			peak.Store(max(peak.Load(), running.Add(1)))
			time.Sleep(time.Second)
			running.Add(-1)
			return n
		}).Slice()

		slices.Sort(result)
		assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7}, result)
		assert.Equal(t, int32(4), peak.Load())
		assert.Equal(t, 2*time.Second, time.Since(start))
	})
}

func TestTailWraps(t *testing.T) {
	ctx := context.Background()

	for n := 1; n <= 12; n++ {
		for size := 0; size <= 12; size++ {
			expected := makeBenchInts(size)
			expected = expected[max(0, size-n):]
			assert.Equal(t, expected, Iota(ctx, 0, size, 1).Tail(n).Slice(), "tail %d of %d", n, size)
		}
	}
}

func TestPadLeftLengths(t *testing.T) {
	ctx := context.Background()

	for n := -1; n <= 6; n++ {
		for size := 0; size <= 6; size++ {
			values := makeBenchInts(size)
			expected := append(slices.Repeat([]int{-1}, max(0, n-size)), values...)
			assert.Equal(t, expected, FromSlice(ctx, values).PadLeft(-1, n).Slice(), "pad %d of %d", n, size)
		}
	}
}

func TestStrideAndTakeCounts(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, []int{0, 3, 6, 9}, Iota(ctx, 0, 10, 1).Stride(3).Slice())
	assert.Equal(t, []int{0, 1, 2, 3}, Iota(ctx, 0, 4, 1).Stride(1).Slice())

	// Take must not ask for a value it doesn't need.
	calls := 0
	Generate(ctx, func() int { calls++; return calls }).Take(2).Drain()
	assert.Equal(t, 2, calls)
}

func TestPrimesAgainstSieve(t *testing.T) {
	const limit = 20_000

	composite := make([]bool, limit)
	expected := make([]int, 0)
	for n := 2; n < limit; n++ {
		if composite[n] {
			continue
		}
		expected = append(expected, n)
		for m := n * n; m < limit; m += n {
			composite[m] = true
		}
	}

	result := Primes(context.Background()).Until(func(n int) bool { return n >= limit }).Slice()

	assert.Equal(t, expected, result)
}

func TestFromStringEdgeCases(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, []string{""}, FromString(ctx, "", " ").Slice())
	assert.Equal(t, []string{"", "a", "", "b", ""}, FromString(ctx, " a  b ", " ").Slice())
	assert.Equal(t, []string{"a", "é", "b"}, FromString(ctx, "aéb", "").Slice())
	assert.Equal(t, []string{"abc"}, FromString(ctx, "abc", "::").Slice())
}

// Write must produce the same bytes as encoding/binary for fixed size types.
func TestWriteMatchesBinary(t *testing.T) {
	type pair struct{ A, B int16 }

	t.Run("uint8", func(t *testing.T) { checkWrite(t, []uint8{0, 1, 255}) })
	t.Run("int8", func(t *testing.T) { checkWrite(t, []int8{0, -1, 127, -128}) })
	t.Run("uint16", func(t *testing.T) { checkWrite(t, []uint16{0, 1, 0xfffe}) })
	t.Run("int16", func(t *testing.T) { checkWrite(t, []int16{0, -2, 0x7ffe}) })
	t.Run("uint32", func(t *testing.T) { checkWrite(t, []uint32{0, 1, 0xfffffffe}) })
	t.Run("int32", func(t *testing.T) { checkWrite(t, []int32{0, -2, 0x7ffffffe}) })
	t.Run("uint64", func(t *testing.T) { checkWrite(t, []uint64{0, 1, 0xfffffffffffffffe}) })
	t.Run("int64", func(t *testing.T) { checkWrite(t, []int64{0, -2, 0x7ffffffffffffffe}) })
	t.Run("float32", func(t *testing.T) { checkWrite(t, []float32{0, -1.5, 3.25e10}) })
	t.Run("float64", func(t *testing.T) { checkWrite(t, []float64{0, -1.5, 3.25e100}) })
	t.Run("bool", func(t *testing.T) { checkWrite(t, []bool{true, false, true}) })
	t.Run("struct", func(t *testing.T) { checkWrite(t, []pair{{1, -2}, {3, 4}}) })
}

func checkWrite[T comparable](t *testing.T, vals []T) {
	t.Helper()

	expected := new(bytes.Buffer)
	assert.NoError(t, binary.Write(expected, binary.LittleEndian, vals))

	result := new(bytes.Buffer)
	assert.NoError(t, FromSlice(context.Background(), vals).Write(result))

	assert.Equal(t, expected.Bytes(), result.Bytes())
}

func TestWriteUnsupported(t *testing.T) {
	type named struct{ S string }

	err := FromSlice(context.Background(), []named{{"a"}}).Write(new(bytes.Buffer))

	assert.Error(t, err)
}
