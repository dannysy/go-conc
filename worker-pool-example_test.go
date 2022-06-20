package conc

import (
	"context"
	"math/rand"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dannysy/go-conc/wp"
)

// Merge sort by work-pool example
func TestWorkerPoolExample(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	in := []int{45, 1, 23, 56, 11, 777, 22, -12, 233}
	parts := 4
	chunkLen := len(in) / parts
	tail := len(in) % parts
	options := wp.GetDefaultOptions()
	if runtime.GOMAXPROCS(0) == 1 {
		options.Size = 2
	}
	pool := wp.NewWorkerPool(options)

	for i := 0; i < parts; i++ {
		startIdx := i * chunkLen
		endIdx := (i + 1) * chunkLen
		if i == parts-1 {
			endIdx += tail
		}
		pool.Add(wp.NewTask(
			func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return mSort(args[0].([]int)), nil
			}, in[startIdx:endIdx]))
	}
	pool.Run(ctx)
	var got []int
	for i := 0; i < parts; i++ {
		result := <-pool.Result()
		got = mrg(got, result.GetValue().([]int))
	}
	cancel()
	<-pool.Done()
	assert.ElementsMatch(t, []int{-12, 1, 11, 22, 23, 45, 56, 233, 777}, got)
	stats := pool.Stats()
	assert.Equal(t, uint32(0), stats.WorkersCount)
	assert.Equal(t, uint32(0), stats.Idlers)
	assert.Equal(t, uint32(0), stats.TasksInQueue)
}

func BenchmarkWorkerPoolMergeSort(b *testing.B) {
	in := genIntSlice()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		parts := 2
		chunkLen := len(in) / parts
		tail := len(in) % parts
		options := wp.GetDefaultOptions()
		if runtime.GOMAXPROCS(0) == 1 {
			options.Size = 2
		}
		options.Idle = uint32(parts)
		pool := wp.NewWorkerPool(options)

		for j := 0; j < parts; j++ {
			startIdx := j * chunkLen
			endIdx := (j + 1) * chunkLen
			if j == parts-1 {
				endIdx += tail
			}
			pool.Add(wp.NewTask(
				func(ctx context.Context, args ...interface{}) (interface{}, error) {
					return mSort(args[0].([]int)), nil
				}, in[startIdx:endIdx]))
		}
		pool.Run(ctx)
		var got []int
		for j := 0; j < parts; j++ {
			result := <-pool.Result()
			got = mrg(got, result.GetValue().([]int))
		}
		cancel()
		<-pool.Done()
	}
}

func BenchmarkMergeSort(b *testing.B) {
	in := genIntSlice()
	for i := 0; i < b.N; i++ {
		mSort(in)
	}
}

func genIntSlice() (out []int) {
	size := 1000
	out = make([]int, size)
	rand.Seed(1111)
	for i := 0; i < size; i++ {
		out[i] = rand.Intn(1000000)
	}
	return out
}

func mSort(slice []int) []int {

	if len(slice) < 2 {
		return slice
	}
	mid := (len(slice)) / 2
	return mrg(mSort(slice[:mid]), mSort(slice[mid:]))
}

func mrg(left, right []int) []int {

	size, i, j := len(left)+len(right), 0, 0
	slice := make([]int, size)

	for k := 0; k < size; k++ {
		if i > len(left)-1 && j <= len(right)-1 {
			slice[k] = right[j]
			j++
		} else if j > len(right)-1 && i <= len(left)-1 {
			slice[k] = left[i]
			i++
		} else if left[i] < right[j] {
			slice[k] = left[i]
			i++
		} else {
			slice[k] = right[j]
			j++
		}
	}
	return slice
}
