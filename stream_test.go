package stream

import (
	"context"
	"fmt"
	"math"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStreamComparable(t *testing.T) {
	ctx := context.Background()

	a := Iota(ctx, 1, 10, 1).Chunk(2)
	b := Iota(ctx, 1, 10, 1).Chunk(2)

	assert.NotEqual(t, a, b)
	assert.True(t, a != b)

	c := FromChannel(ctx, a)
	d := FromChannel(ctx, b)

	assert.NotEqual(t, c, d)
	assert.True(t, c != d)
}

func TestTake(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	result := Iota(context.Background(), 1, 10, 1).Take(5).Slice()

	assert.Equal(t, expected, result)

	empty := []rune{}
	c := FromRunes(context.Background(), "")

	assert.Equal(t, empty, c.Slice())
	assert.Equal(t, empty, c.Slice())
	assert.Equal(t, empty, c.Slice())
}

func TestTakeNotEnough(t *testing.T) {
	expected := []int{1, 2}
	result := Iota(context.Background(), 1, 3, 1).Take(5).Slice()

	assert.Equal(t, expected, result)
}

func TestTakeWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Take(10)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func TestTakeZeroValue(t *testing.T) {
	var stream Stream[int]

	assert.NotPanics(t, func() {
		assert.Equal(t, []int{}, stream.Take(0).Slice())
	})
}

func ExampleStream_Take() {
	result := Iota(context.Background(), 1, 10, 1).Take(5).Slice()

	fmt.Println(result)
	// Output: [1 2 3 4 5]
}

func TestUntil(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	result := Iota(context.Background(), 1, 10, 1).Until(func(val int) bool { return val > 5 }).Slice()

	assert.Equal(t, expected, result)
}

func TestUntilNotEnough(t *testing.T) {
	expected := []int{1, 2}
	result := Iota(context.Background(), 1, 3, 1).Until(func(val int) bool { return val > 5 }).Slice()

	assert.Equal(t, expected, result)
}

func TestUntilWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Until(func(int) bool { return false })

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Until() {
	result := Iota(context.Background(), 1, 1000, 1).Until(func(val int) bool {
		return val > 5
	}).Slice()

	fmt.Println(result)
	// Output: [1 2 3 4 5]
}

func TestMap(t *testing.T) {
	expected := "Yberz vcfhz qbybe fvg nzrg"
	result := FromRunes(context.Background(), "Lorem ipsum dolor sit amet").Map(func(val rune) rune {
		if (val >= 'A' && val <= 'M') || (val >= 'a' && val <= 'm') {
			return val + 13
		} else if (val >= 'N' && val <= 'Z') || (val >= 'n' && val <= 'z') {
			return val - 13
		} else {
			return val
		}
	}).ToString()

	assert.Equal(t, expected, result)
}

func TestMapTypes(t *testing.T) {
	expected := []int{5, 5, 5, 3, 4}
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Map(func(val string) int {
		return len(val)
	}).Slice()

	assert.Equal(t, expected, result)
}

func TestMapWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Map(func(v int) int { return v * 2 })

		assert.Equal(t, 2, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Map() {
	rot13 := func(val rune) rune {
		if (val >= 'A' && val <= 'M') || (val >= 'a' && val <= 'm') {
			return val + 13
		} else if (val >= 'N' && val <= 'Z') || (val >= 'n' && val <= 'z') {
			return val - 13
		} else {
			return val
		}
	}

	result := FromRunes(context.Background(), "Lorem ipsum dolor sit amet").Map(rot13).ToString()

	fmt.Println(result)
	// Output: Yberz vcfhz qbybe fvg nzrg
}

func TestMapParallel(t *testing.T) {
	expected := []int{32, 76, 101, 105, 109, 109, 111, 112, 114, 115, 117}

	synctest.Test(t, func(*testing.T) {
		result := FromRunes(context.Background(), "Lorem ipsum").
			MapParallel(10, func(val rune) int {
				// time.Sleep(time.Second)
				// fmt.Printf("%s: %d\n", time.Now().Format(time.RFC3339), val)
				return int(val)
			}).
			Slice()

		synctest.Wait()
		slices.Sort(result)

		assert.Equal(t, expected, result)
	})
}

func TestMapParallelWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).MapParallel(4, func(v int) int { return v * 2 })

		assert.Equal(t, 2, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_MapParallel() {
	t := &testing.T{}

	synctest.Test(t, func(*testing.T) {
		FromRunes(context.Background(), "Lorem ipsum").
			MapParallel(10, func(val rune) int {
				fmt.Printf("%s\n", time.Now().Format(time.RFC3339))
				time.Sleep(time.Second)
				return int(val)
			}).
			Slice()

		synctest.Wait()

		// Output:
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:00Z
		// 2000-01-01T00:00:01Z
	})
}

func TestFilter(t *testing.T) {
	expected := []int{2, 4, 6, 8, 10}
	result := Iota(context.Background(), 1, 12, 1).Filter(func(val int) bool { return val%2 == 0 }).Slice()

	assert.Equal(t, expected, result)
}

func TestFilterWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Filter(func(int) bool { return true })

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Filter() {
	even := func(val int) bool {
		return val%2 == 0
	}

	result := Iota(context.Background(), 1, 12, 1).Filter(even).Slice()

	fmt.Println(result)
	// Output: [2 4 6 8 10]
}

func TestReduce(t *testing.T) {
	expected := 45
	result := Iota(context.Background(), 1, 10, 1).Reduce(func(a, b int) int { return a + b }).Pop()

	assert.Equal(t, expected, result)
}

func TestReduceEmpty(t *testing.T) {
	expected := []int{}
	result := FromSlice(context.Background(), []int{}).Reduce(func(a, b int) int { return a + b }).Slice()

	assert.Equal(t, expected, result)
}

func TestReduceWithContext(t *testing.T) {
	// Repeated because select chooses randomly between a waiting reader and a
	// cancelled context.
	for range 100 {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Sends some values but never closes.
			input := make(chan int)
			go func() {
				for i := range 3 {
					input <- i
				}
			}()

			stream := FromChannel(ctx, input).Reduce(func(a, b int) int { return a + b })

			received := make(chan bool)
			go func() {
				_, ok := <-stream.c
				received <- ok
			}()
			synctest.Wait()

			cancel()

			assert.False(t, <-received, "expected stream to be closed without a value")
		})
	}
}

func ExampleStream_Reduce() {
	sum := func(a, b int) int {
		return a + b
	}

	result := Iota(context.Background(), 1, 10, 1).Reduce(sum).Pop()

	fmt.Println(result)
	// Output: 45
}

func TestLast(t *testing.T) {
	expected := 9
	result := Iota(context.Background(), 1, 10, 1).Last().Pop()

	assert.Equal(t, expected, result)
}

func TestLastEmpty(t *testing.T) {
	expected := []int{}
	result := FromSlice(context.Background(), []int{}).Last().Slice()

	assert.Equal(t, expected, result)
}

func TestLastWithContext(t *testing.T) {
	// Repeated because select chooses randomly between a waiting reader and a
	// cancelled context.
	for range 100 {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Sends some values but never closes.
			input := make(chan int)
			go func() {
				for i := range 3 {
					input <- i
				}
			}()

			stream := FromChannel(ctx, input).Last()

			received := make(chan bool)
			go func() {
				_, ok := <-stream.c
				received <- ok
			}()
			synctest.Wait()

			cancel()

			assert.False(t, <-received, "expected stream to be closed without a value")
		})
	}
}

func ExampleStream_Last() {
	result := Iota(context.Background(), 1, 10, 1).Last().Pop()

	fmt.Println(result)
	// Output: 9
}

func TestChain(t *testing.T) {
	expected := "Lorem ipsum dolor sit amet"

	a := FromRunes(context.Background(), "Lorem ipsum")
	b := FromRunes(context.Background(), " dolor")
	c := FromRunes(context.Background(), " sit amet")
	result := a.Chain(b).Chain(c).ToString()

	assert.Equal(t, expected, result)
}

func TestChainVariadic(t *testing.T) {
	expected := "Lorem ipsum dolor sit amet"

	a := FromRunes(context.Background(), "Lorem ipsum")
	b := FromRunes(context.Background(), " dolor")
	c := FromRunes(context.Background(), " sit amet")
	result := a.Chain(b, c).ToString()

	assert.Equal(t, expected, result)
}

func TestChainWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		otherCtx, otherCancel := context.WithCancel(context.Background())
		defer otherCancel()

		// The passed stream never sends, so the stage can only close via the
		// main stream's context.
		a := FromSlice(ctx, []int{1})
		b := FromChannel(otherCtx, make(chan int))
		stream := a.Chain(b)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Chain() {
	a := FromRunes(context.Background(), "Lorem ipsum")
	b := FromRunes(context.Background(), " dolor")
	c := FromRunes(context.Background(), " sit amet")

	result := a.Chain(b, c).ToString()

	fmt.Println(result)
	// Output: Lorem ipsum dolor sit amet
}

func TestMerge(t *testing.T) {
	expected := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefghijklmnopqrstuvwxyz")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := a.Merge(b).Merge(c).ToString()
	runes := []rune(result)

	slices.Sort(runes)

	assert.Equal(t, expected, string(runes))
}

func TestMergeVariadic1(t *testing.T) {
	expected := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh"

	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefgh")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := a.Merge(b, c).ToString()
	runes := []rune(result)

	slices.Sort(runes)

	assert.Equal(t, expected, string(runes))
}

func TestMergeVariadic2(t *testing.T) {
	expected := "0123456789ABCDEabcdefgh"

	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefgh")
	c := FromRunes(context.Background(), "ABCDE")

	result := a.Merge(b, c).ToString()
	runes := []rune(result)

	slices.Sort(runes)

	assert.Equal(t, expected, string(runes))
}

func ExampleStream_Merge() {
	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefgh")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := a.Merge(b, c).ToString()
	runes := []rune(result)

	slices.Sort(runes)

	fmt.Println(string(runes))
	// Output: 0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh
}

func TestRoundRobin(t *testing.T) {
	expected := "0AaB1CbD2EcF3GdH4IeJ5KfL6MgN7OhP8QiR9SjTkUlVmWnXoYpZqrstuvwxyz"

	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefghijklmnopqrstuvwxyz")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := a.RoundRobin(b).RoundRobin(c).ToString()

	assert.Equal(t, expected, result)
}

func TestRoundRobinVariadic1(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5fF6gG7hH8I9JKLMNOPQRSTUVWXYZ"

	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefgh")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := a.RoundRobin(b, c).ToString()

	assert.Equal(t, expected, result)
}

func TestRoundRobinVariadic2(t *testing.T) {
	expected := "0aA1bB2cC3dD4eE5f6g7h89"

	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefgh")
	c := FromRunes(context.Background(), "ABCDE")
	result := a.RoundRobin(b, c).ToString()

	assert.Equal(t, expected, result)
}

func TestRoundRobinVariadic3(t *testing.T) {
	expected := "0aAx1bBy2z"

	a := FromRunes(context.Background(), "012")
	b := FromRunes(context.Background(), "ab")
	c := FromRunes(context.Background(), "AB")
	d := FromRunes(context.Background(), "xyz")
	result := a.RoundRobin(b, c, d).ToString()

	assert.Equal(t, expected, result)
}

func TestRoundRobinWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		otherCtx, otherCancel := context.WithCancel(context.Background())
		defer otherCancel()

		// The passed stream never sends, so the stage can only close via the
		// main stream's context.
		a := FromSlice(ctx, []int{1})
		b := FromChannel(otherCtx, make(chan int))
		stream := a.RoundRobin(b)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_RoundRobin() {
	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefgh")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := a.RoundRobin(b, c).ToString()

	fmt.Println(result)
	// Output: 0aA1bB2cC3dD4eE5fF6gG7hH8I9JKLMNOPQRSTUVWXYZ
}

func TestChunk(t *testing.T) {
	expected := [][]int{{2, 4}, {6, 8}, {10}}
	i := 0
	for c := range Iota(context.Background(), 2, 12, 2).Chunk(2) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestChunkZero(t *testing.T) {
	result := Iota(context.Background(), 2, 12, 2).Chunk(0)

	_, ok := <-result
	assert.False(t, ok, "expected channel to be closed")
}

func TestChunkWithContext(t *testing.T) {
	// Repeated because select chooses randomly between a waiting reader and a
	// cancelled context.
	for range 100 {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Sends some values but never closes.
			input := make(chan int)
			go func() {
				for i := range 3 {
					input <- i
				}
			}()

			stream := FromChannel(ctx, input).Chunk(5)

			received := make(chan bool)
			go func() {
				_, ok := <-stream
				received <- ok
			}()
			synctest.Wait()

			cancel()

			assert.False(t, <-received, "expected stream to be closed without a value")
		})
	}
}

func ExampleStream_Chunk() {
	for c := range Iota(context.Background(), 2, 20, 2).Chunk(4) {
		fmt.Println(c.Slice())
	}
	// Output:
	// [2 4 6 8]
	// [10 12 14 16]
	// [18]
}

func TestDrop(t *testing.T) {
	expected := []int{6, 7, 8, 9, 10}
	result := Iota(context.Background(), 1, 20, 1).Drop(5).Take(5).Slice()

	assert.Equal(t, expected, result)

	expected = []int{6, 7, 8}
	empty := []int{}
	c := Iota(context.Background(), 1, 10, 1).Take(8)

	assert.Equal(t, expected, c.Drop(5).Slice())
	assert.Equal(t, empty, c.Drop(5).Slice())
	assert.Equal(t, empty, c.Drop(5).Slice())
}

func TestDropWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Drop(1)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Drop() {
	result := Iota(context.Background(), 1, 20, 1).Drop(5).Take(5).Slice()

	fmt.Println(result)
	// Output: [6 7 8 9 10]
}

func TestStride(t *testing.T) {
	expected := []int{1, 4, 7, 10, 13, 16, 19, 22, 25, 28}
	result := Iota(context.Background(), 1, 100, 1).Stride(3).Take(10).Slice()

	assert.Equal(t, expected, result)
}

func TestStrideZero(t *testing.T) {
	expected := []int{}
	result := Iota(context.Background(), 1, 100, 1).Stride(0).Slice()

	assert.Equal(t, expected, result)
}

func TestStrideWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Stride(2)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Stride() {
	result := Iota(context.Background(), 1, 100, 1).Stride(3).Take(10).Slice()

	fmt.Println(result)
	// Output: [1 4 7 10 13 16 19 22 25 28]
}

func TestTail(t *testing.T) {
	expected := []int{17, 18, 19}
	result := Iota(context.Background(), 1, 20, 1).Tail(3).Slice()

	assert.Equal(t, expected, result)
}

func TestTailZero(t *testing.T) {
	expected := []int{}
	result := Iota(context.Background(), 1, 20, 1).Tail(0).Slice()

	assert.Equal(t, expected, result)
}

func TestTailWithContext(t *testing.T) {
	// Repeated because select chooses randomly between a waiting reader and a
	// cancelled context.
	for range 100 {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Sends some values but never closes.
			input := make(chan int)
			go func() {
				for i := range 3 {
					input <- i
				}
			}()

			stream := FromChannel(ctx, input).Tail(2)

			received := make(chan bool)
			go func() {
				_, ok := <-stream.c
				received <- ok
			}()
			synctest.Wait()

			cancel()

			assert.False(t, <-received, "expected stream to be closed without a value")
		})
	}
}

func ExampleStream_Tail() {
	result := Iota(context.Background(), 1, 20, 1).Tail(3).Slice()

	fmt.Println(result)
	// Output: [17 18 19]
}

func TestZipVariadic1(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A'},
		{'1', 'b', 'B'},
		{'2', 'c', 'C'},
		{'3', 'd', 'D'},
	}
	a := FromRunes(context.Background(), "0123") // Stops the zip
	b := FromRunes(context.Background(), "abcdefg")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	i := 0
	for c := range a.Zip(b, c) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestZipVariadic2(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A'},
		{'1', 'b', 'B'},
		{'2', 'c', 'C'},
		{'3', 'd', 'D'},
	}
	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcd") // Stops the zip
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	i := 0
	for c := range a.Zip(b, c) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestZipVariadic3(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A'},
		{'1', 'b', 'B'},
		{'2', 'c', 'C'},
		{'3', 'd', 'D'},
	}
	a := FromRunes(context.Background(), "0123456789")
	b := FromRunes(context.Background(), "abcdefghi")
	c := FromRunes(context.Background(), "ABCD") // Stops the zip
	i := 0
	for c := range a.Zip(b, c) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
}

func TestZipVariadic4(t *testing.T) {
	expected := [][]rune{
		{'0', 'a', 'A', 'x'},
		{'1', 'b', 'B', 'y'},
	}
	a := FromRunes(context.Background(), "0123")
	b := FromRunes(context.Background(), "abcd")
	c := FromRunes(context.Background(), "ABCD")
	d := FromRunes(context.Background(), "xy") // Stops the zip
	i := 0
	for c := range a.Zip(b, c, d) {
		result := c.Slice()
		assert.Equal(t, expected[i], result)
		i++
	}
	assert.Equal(t, len(expected), i)
}

func TestZipWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		otherCtx, otherCancel := context.WithCancel(context.Background())
		defer otherCancel()

		// The passed stream never sends, so the stage can only close via the
		// main stream's context.
		a := FromSlice(ctx, []int{1})
		b := FromChannel(otherCtx, make(chan int))
		stream := a.Zip(b)
		synctest.Wait()

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Zip() {
	a := FromRunes(context.Background(), "0123")
	b := FromRunes(context.Background(), "abcdefg")
	c := FromRunes(context.Background(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	for c := range a.Zip(b, c) {
		fmt.Println(c.ToString())
	}
	// Output:
	// 0aA
	// 1bB
	// 2cC
	// 3dD
}

func TestPadRight(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 0, 0, 0}
	result := Iota(context.Background(), 1, 6, 1).PadRight(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestPadRightExceeded(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := Iota(context.Background(), 1, 10, 1).PadRight(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestPadRightWithContext(t *testing.T) {
	// Repeated because select chooses randomly between a waiting reader and a
	// cancelled context.
	for range 100 {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Never sends or closes.
			input := make(chan int)

			stream := FromChannel(ctx, input).PadRight(0, 5)

			received := make(chan bool)
			go func() {
				_, ok := <-stream.c
				received <- ok
			}()
			synctest.Wait()

			cancel()

			assert.False(t, <-received, "expected stream to be closed without a value")
		})
	}
}

func ExampleStream_PadRight() {
	result := Iota(context.Background(), 1, 6, 1).PadRight(0, 8).Slice()

	fmt.Println(result)
	// Output: [1 2 3 4 5 0 0 0]
}

func TestPadLeft(t *testing.T) {
	expected := []int{0, 0, 0, 1, 2, 3, 4, 5}
	result := Iota(context.Background(), 1, 6, 1).PadLeft(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestPadLeftExceeded(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := Iota(context.Background(), 1, 10, 1).PadLeft(0, 8).Slice()

	assert.Equal(t, expected, result)
}

func TestPadLeftNegative(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	result := Iota(context.Background(), 1, 6, 1).PadLeft(0, -1).Slice()

	assert.Equal(t, expected, result)
}

func TestPadLeftLarge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expected := []int{0, 0, 0}
	result := Iota(ctx, 1, 6, 1).PadLeft(0, math.MaxInt).Take(3).Slice()

	assert.Equal(t, expected, result)
}

func TestPadLeftWithContext(t *testing.T) {
	// Repeated because select chooses randomly between a waiting reader and a
	// cancelled context.
	for range 100 {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Sends some values but never closes.
			input := make(chan int)
			go func() {
				for i := range 3 {
					input <- i
				}
			}()

			stream := FromChannel(ctx, input).PadLeft(0, 5)

			received := make(chan bool)
			go func() {
				_, ok := <-stream.c
				received <- ok
			}()
			synctest.Wait()

			cancel()

			assert.False(t, <-received, "expected stream to be closed without a value")
		})
	}
}

func ExampleStream_PadLeft() {
	result := Iota(context.Background(), 1, 6, 1).PadLeft(0, 8).Slice()

	fmt.Println(result)
	// Output: [0 0 0 1 2 3 4 5]
}

func TestTee(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	tee := 0
	result := Iota(context.Background(), 1, 6, 1).Tee(func(val int) { tee = val }).Slice()

	assert.Equal(t, expected, result)
	assert.Equal(t, 5, tee)
}

func TestTeeWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Tee(func(int) {})

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Tee() {
	count := 0

	result := Iota(context.Background(), 1, 6, 1).Tee(func(val int) {
		count++
	}).Slice()

	fmt.Printf("values: %v count: %d\n", result, count)
	// Output: values: [1 2 3 4 5] count: 5
}

func TestEnumerate(t *testing.T) {
	expected := []Enum[string]{
		{
			Index: 1,
			Val:   "Lorem",
		},
		{
			Index: 2,
			Val:   "ipsum",
		},
	}

	i := 0
	for val := range FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Take(2).Enumerate(1) {
		assert.Equal(t, expected[i], val)
		i++
	}
}

func TestEnumerateWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, "Lorem").Enumerate(1)

		assert.Equal(t, Enum[string]{Index: 1, Val: "Lorem"}, <-stream)

		cancel()
		synctest.Wait()

		_, ok := <-stream
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Enumerate() {
	text := "Lorem ipsum dolor sit amet"

	for v := range FromString(context.Background(), text, " ").Enumerate(1) {
		fmt.Printf("%d: %v\n", v.Index, v.Val)
	}
	// Output:
	// 1: Lorem
	// 2: ipsum
	// 3: dolor
	// 4: sit
	// 5: amet
}

func TestFind(t *testing.T) {
	expected := []string{"dolor", "sit", "amet"}
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Find("dolor").Slice()

	assert.Equal(t, expected, result)
}

func TestFindWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Find(1)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Find() {
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Find("dolor").Slice()

	fmt.Println(result)
	// Output: [dolor sit amet]
}

func TestSubstitute(t *testing.T) {
	expected := []string{"Lorem", "ipsum", "lectus", "sit", "amet"}
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Substitute("dolor", "lectus").Slice()

	assert.Equal(t, expected, result)
}

func TestSubstituteWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Substitute(1, 2)

		assert.Equal(t, 2, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Substitute() {
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Substitute("dolor", "lectus").Slice()

	fmt.Println(result)
	// Output: [Lorem ipsum lectus sit amet]
}

func TestSkip(t *testing.T) {
	expected := []int{1, 3, 1, 3, 1, 3}
	result := FromSlice(context.Background(), []int{1, 2, 3, 1, 2, 3, 1, 2, 3}).Skip(2).Slice()

	assert.Equal(t, expected, result)
}

func TestSkipWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Skip(2)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Skip() {
	result := FromSlice(context.Background(), []int{1, 2, 3, 1, 2, 3, 1, 2, 3}).Skip(2).Slice()

	fmt.Println(result)
	// Output: [1 3 1 3 1 3]
}

func TestThrottle(t *testing.T) {
	expected := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}
	result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").Throttle(10, time.Second).Slice()

	assert.Equal(t, expected, result)
}

func TestThrottleWindow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		input := make(chan int)
		go func() {
			defer close(input)
			// Arrive just before the first tick of a ticker started at zero.
			time.Sleep(900 * time.Millisecond)
			for i := range 6 {
				input <- i
			}
		}()

		times := make([]time.Time, 0)
		FromChannel(context.Background(), input).
			Throttle(2, time.Second).
			Tee(func(int) { times = append(times, time.Now()) }).
			Drain()

		assert.Len(t, times, 6)
		for i := 2; i < len(times); i++ {
			assert.GreaterOrEqual(t, times[i].Sub(times[i-2]), time.Second, "more than 2 values within a second")
		}
	})
}

func TestThrottleWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Throttle(1, time.Hour)

		assert.Equal(t, 1, stream.Pop())

		cancel()
		synctest.Wait()

		start := time.Now()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
		assert.Equal(t, start, time.Now(), "expected stream to close without waiting")
	})
}

func ExampleStream_Throttle() {
	t := &testing.T{}

	synctest.Test(t, func(t *testing.T) {
		result := FromString(context.Background(), "Lorem ipsum dolor sit amet", " ").
			Throttle(1, time.Second).
			Tee(func(val string) {
				fmt.Printf("%s: %s\n", time.Now().Format(time.RFC3339), val)
			}).
			ToString()

		fmt.Println(result)
		// Output:
		// 2000-01-01T00:00:00Z: Lorem
		// 2000-01-01T00:00:01Z: ipsum
		// 2000-01-01T00:00:02Z: dolor
		// 2000-01-01T00:00:03Z: sit
		// 2000-01-01T00:00:04Z: amet
		// [Lorem ipsum dolor sit amet]
	})
}

func TestDistinct(t *testing.T) {
	expected := []string{"Lorem", "ipsum", "dolor", "sit", "amet"}
	result := FromString(context.Background(), "Lorem ipsum ipsum dolor Lorem sit amet dolor amet sit", " ").Distinct().Slice()

	assert.Equal(t, expected, result)
}

func TestDistinctWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Iota(ctx, 0, 10, 1).Distinct()

		assert.Equal(t, 0, stream.Pop())

		cancel()
		synctest.Wait()

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Distinct() {
	result := FromString(context.Background(), "Lorem ipsum ipsum dolor Lorem sit amet dolor amet sit", " ").Distinct().ToString()

	fmt.Println(result)
	// Output: [Lorem ipsum dolor sit amet]
}

func TestBuffer(t *testing.T) {
	count := 0

	synctest.Test(t, func(*testing.T) {
		FromSlice(context.Background(), []int{1, 2, 3, 4, 5}).
			Tee(func(val int) {
				count++
			}).
			Buffer(5)

		synctest.Wait()

		assert.Equal(t, 5, count)
	})
}

func TestBufferNegative(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5}
	result := FromSlice(context.Background(), []int{1, 2, 3, 4, 5}).Buffer(-1).Slice()

	assert.Equal(t, expected, result)
}

func TestBufferWithContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		stream := Repeat(ctx, 1).Buffer(2)
		synctest.Wait()

		cancel()
		synctest.Wait()

		// Values already buffered are still received.
		assert.Equal(t, 1, <-stream.c)
		assert.Equal(t, 1, <-stream.c)

		_, ok := <-stream.c
		assert.False(t, ok, "expected stream to be closed")
	})
}

func ExampleStream_Buffer() {
	t := &testing.T{}

	synctest.Test(t, func(t *testing.T) {
		FromSlice(context.Background(), []int{1, 2, 3, 4, 5}).
			Tee(func(val int) {
				fmt.Println(val)
			}).
			Buffer(5)

		synctest.Wait()

		// Output:
		// 1
		// 2
		// 3
		// 4
		// 5
	})
}

func TestNotStringer(t *testing.T) {
	_, ok := any(Stream[int]{}).(fmt.Stringer)

	assert.False(t, ok, "formatting a stream must not consume it")
}
