package stream

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

// benchSize is the number of values pushed through a stream per iteration.
const benchSize = 10_000

var (
	benchInts   = makeBenchInts(benchSize)
	benchText   = strings.Repeat("lorem ipsum dolor sit amet ", benchSize/5)
	benchBytes  = bytes.Repeat([]byte("0123456789abcdef"), benchSize/16)
	benchIntSum int
)

func makeBenchInts(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

// perValue reports the time taken per streamed value.
func perValue(b *testing.B, n int) {
	b.Helper()
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/value")
}

func inc(n int) int    { return n + 1 }
func even(n int) bool  { return n%2 == 0 }
func add(a, b int) int { return a + b }
func never(int) bool   { return false }
func discard(int)      {}
func ints(ctx context.Context) Stream[int] {
	return FromSlice(ctx, benchInts)
}

// ---------------------------------------------------------------------------
// Generators
// ---------------------------------------------------------------------------

func BenchmarkFromSlice(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		FromSlice(ctx, benchInts).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkCycle(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		Cycle(ctx, benchInts[:16]).Take(benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkGenerate(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		i := 0
		Generate(ctx, func() int { i++; return i }).Take(benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkRepeat(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		Repeat(ctx, 1).Take(benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkFromChannel(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		c := make(chan int, benchSize)
		for _, v := range benchInts {
			c <- v
		}
		close(c)
		FromChannel(ctx, c).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkFromString(b *testing.B) {
	ctx := context.Background()
	n := strings.Count(benchText, " ") + 1
	b.ReportAllocs()
	for b.Loop() {
		FromString(ctx, benchText, " ").Drain()
	}
	perValue(b, n)
}

func BenchmarkFromRunes(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		FromRunes(ctx, benchText[:benchSize]).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkFromReader(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		FromReader(ctx, bytes.NewReader(benchBytes)).Drain()
	}
	perValue(b, len(benchBytes))
}

func BenchmarkIota(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		Iota(ctx, 0, benchSize, 1).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkFibonacci(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		Fibonacci(ctx).Take(1000).Drain()
	}
	perValue(b, 1000)
}

func BenchmarkPrimes(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		Primes(ctx).Take(benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkRandInt(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		RandInt(ctx).Take(benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkRandFloat64(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		RandFloat64(ctx).Take(benchSize).Drain()
	}
	perValue(b, benchSize)
}

// ---------------------------------------------------------------------------
// Stages. Each benchmark is a slice source, the stage under test and a drain.
// ---------------------------------------------------------------------------

func BenchmarkAll(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		sum := 0
		for v := range ints(ctx).All() {
			sum += v
		}
		benchIntSum = sum
	}
	perValue(b, benchSize)
}

func BenchmarkTake(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Take(benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkUntil(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Until(never).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkMap(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Map(inc).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkMapParallel(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).MapParallel(4, inc).Drain()
	}
	perValue(b, benchSize)
}

// BenchmarkMapParallelWork maps with a function that does real work, which is
// the case MapParallel exists for.
func BenchmarkMapParallelWork(b *testing.B) {
	ctx := context.Background()
	work := func(n int) int {
		for i := range 2000 {
			n = n*31 + i
		}
		return n
	}
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).MapParallel(4, work).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkFilter(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Filter(even).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkReduce(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = ints(ctx).Reduce(add).Pop()
	}
	perValue(b, benchSize)
}

func BenchmarkLast(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = ints(ctx).Last().Pop()
	}
	perValue(b, benchSize)
}

func BenchmarkChain(b *testing.B) {
	ctx := context.Background()
	half := benchSize / 2
	b.ReportAllocs()
	for b.Loop() {
		FromSlice(ctx, benchInts[:half]).Chain(FromSlice(ctx, benchInts[half:])).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkMerge(b *testing.B) {
	ctx := context.Background()
	half := benchSize / 2
	b.ReportAllocs()
	for b.Loop() {
		FromSlice(ctx, benchInts[:half]).Merge(FromSlice(ctx, benchInts[half:])).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkRoundRobin(b *testing.B) {
	ctx := context.Background()
	half := benchSize / 2
	b.ReportAllocs()
	for b.Loop() {
		FromSlice(ctx, benchInts[:half]).RoundRobin(FromSlice(ctx, benchInts[half:])).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkChunk(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		for chunk := range ints(ctx).Chunk(10) {
			chunk.Drain()
		}
	}
	perValue(b, benchSize)
}

func BenchmarkDrop(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Drop(benchSize / 2).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkStride(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Stride(4).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkTail(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Tail(100).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkZip(b *testing.B) {
	ctx := context.Background()
	half := benchSize / 2
	b.ReportAllocs()
	for b.Loop() {
		for zip := range FromSlice(ctx, benchInts[:half]).Zip(FromSlice(ctx, benchInts[half:])) {
			zip.Drain()
		}
	}
	perValue(b, benchSize)
}

func BenchmarkPadRight(b *testing.B) {
	ctx := context.Background()
	half := benchSize / 2
	b.ReportAllocs()
	for b.Loop() {
		FromSlice(ctx, benchInts[:half]).PadRight(0, benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkPadLeft(b *testing.B) {
	ctx := context.Background()
	half := benchSize / 2
	b.ReportAllocs()
	for b.Loop() {
		FromSlice(ctx, benchInts[:half]).PadLeft(0, benchSize).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkTee(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Tee(discard).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkEnumerate(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		for range ints(ctx).Enumerate(0) {
		}
	}
	perValue(b, benchSize)
}

func BenchmarkFind(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Find(benchSize / 2).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkSubstitute(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Substitute(1, 2).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkSkip(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Skip(1).Drain()
	}
	perValue(b, benchSize)
}

// BenchmarkThrottle uses a limit high enough that the stream never waits, so
// it measures the bookkeeping rather than the sleeping.
func BenchmarkThrottle(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Throttle(benchSize, time.Nanosecond).Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkDistinct(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Distinct().Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkDistinctDuplicates(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		Cycle(ctx, benchInts[:16]).Take(benchSize).Distinct().Drain()
	}
	perValue(b, benchSize)
}

func BenchmarkBuffer(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Buffer(256).Drain()
	}
	perValue(b, benchSize)
}

// ---------------------------------------------------------------------------
// Sinks
// ---------------------------------------------------------------------------

func BenchmarkSlice(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = len(ints(ctx).Slice())
	}
	perValue(b, benchSize)
}

func BenchmarkWriteInt(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		ints(ctx).Write(io.Discard)
	}
	perValue(b, benchSize)
}

func BenchmarkWriteByte(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		FromSlice(ctx, benchBytes).Write(io.Discard)
	}
	perValue(b, len(benchBytes))
}

func BenchmarkWriteString(b *testing.B) {
	ctx := context.Background()
	n := strings.Count(benchText, " ") + 1
	b.ReportAllocs()
	for b.Loop() {
		FromString(ctx, benchText, " ").Write(io.Discard)
	}
	perValue(b, n)
}

func BenchmarkWriteRune(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		FromRunes(ctx, benchText[:benchSize]).Write(io.Discard)
	}
	perValue(b, benchSize)
}

func BenchmarkToStringRunes(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = len(FromRunes(ctx, benchText[:benchSize]).ToString())
	}
	perValue(b, benchSize)
}

func BenchmarkPop(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		s := ints(ctx)
		sum := 0
		for range benchSize {
			sum += s.Pop()
		}
		benchIntSum = sum
	}
	perValue(b, benchSize)
}

// ---------------------------------------------------------------------------
// Pipelines
// ---------------------------------------------------------------------------

// BenchmarkPipeline measures how the cost of a value grows with the number of
// stages it passes through.
func BenchmarkPipeline(b *testing.B) {
	for _, stages := range []int{1, 4, 16} {
		b.Run("stages="+itoa(stages), func(b *testing.B) {
			ctx := context.Background()
			b.ReportAllocs()
			for b.Loop() {
				s := ints(ctx)
				for range stages {
					s = s.Map(inc)
				}
				s.Drain()
			}
			perValue(b, benchSize)
		})
	}
}

// BenchmarkPipelineMixed is a realistic chain of different stages.
func BenchmarkPipelineMixed(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = Iota(ctx, 0, benchSize, 1).
			Filter(even).
			Map(inc).
			Skip(7).
			Substitute(9, 10).
			Stride(2).
			Reduce(add).
			Pop()
	}
	perValue(b, benchSize)
}

// BenchmarkEuler1 is the Project Euler problem from the tests.
func BenchmarkEuler1(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = Iota(ctx, 0, 1000, 1).
			Filter(func(n int) bool { return n%3 == 0 || n%5 == 0 }).
			Reduce(add).
			Pop()
	}
	perValue(b, 1000)
}

// BenchmarkReadme is the example from the readme.
func BenchmarkReadme(b *testing.B) {
	ctx := context.Background()
	text := "Lorem adipiscing elit ipsum sed neque dolor non libero sit consequat magna amet placerat bibendum"
	b.ReportAllocs()
	for b.Loop() {
		FromString(ctx, text, " ").
			Stride(3).
			Take(2).
			Reduce(join).
			Write(io.Discard)
	}
}

// BenchmarkSmallStream measures the fixed cost of building and running a short
// stream, where setup dominates.
func BenchmarkSmallStream(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = FromSlice(ctx, benchInts[:8]).Map(inc).Filter(even).Reduce(add).Pop()
	}
}

// BenchmarkCancellable runs a pipeline under a cancellable context, which is
// how streams are normally used.
func BenchmarkCancellable(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.ReportAllocs()
	for b.Loop() {
		benchIntSum = ints(ctx).Map(inc).Filter(even).Reduce(add).Pop()
	}
	perValue(b, benchSize)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
