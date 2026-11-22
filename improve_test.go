package stream

import (
	"bytes"
	"context"
	"encoding/binary"
	"runtime"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

// A stage that's blocked sending, with nothing receiving, must stop when the
// context is cancelled, as must every stage before it.
func TestCancelStopsEveryStage(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		before := runtime.NumGoroutine()

		stream := Repeat(ctx, 1).
			Map(func(n int) int { return n + 1 }).
			Filter(func(int) bool { return true }).
			Tee(func(int) {}).
			Skip(0).
			Substitute(2, 7).
			Merge(Repeat(ctx, 3), Repeat(ctx, 4)).
			RoundRobin(Repeat(ctx, 5)).
			Chain(Repeat(ctx, 6)).
			MapParallel(3, func(n int) int { return n })
		chunks := Repeat(ctx, 1).Chunk(3)
		zips := Repeat(ctx, 1).Zip(Repeat(ctx, 2))
		enums := Repeat(ctx, 1).Enumerate(0)
		synctest.Wait()

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
		_, ok = <-chunks
		assert.False(t, ok, "expected chunks to be closed")
		_, ok = <-zips
		assert.False(t, ok, "expected zips to be closed")
		_, ok = <-enums
		assert.False(t, ok, "expected enums to be closed")
		assert.Equal(t, before, runtime.NumGoroutine(), "expected no goroutines to be left running")
	})
}

// A stream that finishes must not leave anything registered with its context.
func TestFinishedStreamLeavesNothingBehind(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		before := runtime.NumGoroutine()

		assert.Equal(t, []int{2, 3, 4}, FromSlice(ctx, []int{1, 2, 3}).Map(func(n int) int { return n + 1 }).Slice())
		synctest.Wait()

		cancel()
		synctest.Wait()
		assert.Equal(t, before, runtime.NumGoroutine(), "expected cancelling to start nothing")
	})
}

// Receiving must stop when the context is cancelled, even if a stage is stuck
// in a callback and so cannot close its stream.
func TestCancelWhileSourceIsStuck(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		release := make(chan struct{})
		stuck := func() int { <-release; return 1 }

		time.AfterFunc(time.Second, cancel)

		assert.Equal(t, []int{}, Generate(ctx, stuck).Map(func(n int) int { return n }).Slice())
		assert.Equal(t, 0, Generate(ctx, stuck).Take(1).Pop())
		_, ok := <-Generate(ctx, stuck).Chunk(2)
		assert.False(t, ok, "expected chunks to be closed")
		_, ok = <-Generate(ctx, stuck).Zip(Generate(ctx, stuck))
		assert.False(t, ok, "expected zips to be closed")
		_, ok = <-Generate(ctx, stuck).Enumerate(0)
		assert.False(t, ok, "expected enums to be closed")

		close(release)
	})
}

// Streams bound to another context are read using the main stream's context.
func TestCancelWithForeignStreams(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		other := context.Background()
		var zero Stream[int]

		chain := FromSlice(ctx, []int{1}).Chain(zero)
		foreign := make(chan int)
		defer close(foreign)
		merge := FromSlice(ctx, []int{1}).Merge(FromChannel(other, foreign))
		robin := FromSlice(ctx, []int{1}).RoundRobin(zero)

		assert.Equal(t, 1, chain.Pop())
		assert.Equal(t, 1, merge.Pop())
		assert.Equal(t, 1, robin.Pop())
		synctest.Wait()

		cancel()
		synctest.Wait()

		for _, s := range []Stream[int]{chain, merge, robin} {
			_, ok := <-s.c
			assert.False(t, ok, "expected stream to be closed")
		}
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

	check := func(t *testing.T, write func(*bytes.Buffer) error, vals any) {
		t.Helper()
		expected := new(bytes.Buffer)
		assert.NoError(t, binary.Write(expected, binary.LittleEndian, vals))
		result := new(bytes.Buffer)
		assert.NoError(t, write(result))
		assert.Equal(t, expected.Bytes(), result.Bytes())
	}

	ctx := context.Background()
	t.Run("uint8", func(t *testing.T) {
		v := []uint8{0, 1, 255}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("int8", func(t *testing.T) {
		v := []int8{0, -1, 127, -128}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("uint16", func(t *testing.T) {
		v := []uint16{0, 1, 0xfffe}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("int16", func(t *testing.T) {
		v := []int16{0, -2, 0x7ffe}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("uint32", func(t *testing.T) {
		v := []uint32{0, 1, 0xfffffffe}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("int32", func(t *testing.T) {
		v := []int32{0, -2, 0x7ffffffe}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("uint64", func(t *testing.T) {
		v := []uint64{0, 1, 0xfffffffffffffffe}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("int64", func(t *testing.T) {
		v := []int64{0, -2, 0x7ffffffffffffffe}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("float32", func(t *testing.T) {
		v := []float32{0, -1.5, 3.25e10}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("float64", func(t *testing.T) {
		v := []float64{0, -1.5, 3.25e100}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("bool", func(t *testing.T) {
		v := []bool{true, false, true}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("struct", func(t *testing.T) {
		v := []pair{{1, -2}, {3, 4}}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, v)
	})
	t.Run("int", func(t *testing.T) {
		v := []int{0, -2, 1 << 40}
		check(t, func(b *bytes.Buffer) error { return FromSlice(ctx, v).Write(b) }, []int64{0, -2, 1 << 40})
	})
}

func TestWriteUnsupported(t *testing.T) {
	type named struct{ S string }

	err := FromSlice(context.Background(), []named{{"a"}}).Write(new(bytes.Buffer))

	assert.Error(t, err)
}
