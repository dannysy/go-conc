package conc

import (
	"context"
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
