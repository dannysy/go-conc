package wp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	wp := NewWorkerPool(DefaultOptions)
	wp.Run(ctx)
	wp.Add(func() error {
		panic("oops! panic!")
	})
	wp.Add(sleepSecTask(2, 1))
	wp.Add(sleepSecTask(3, 2))
	wp.Add(sleepSecTask(4, 3))
	wp.Add(sleepSecTask(5, 4))

	printStats(wp)
	time.Sleep(5 * time.Second)
	printStats(wp)
	time.Sleep(1 * time.Second)
	printStats(wp)
	cancel()
	fmt.Println("context cancelled")
	<-wp.Done()
	stats := printStats(wp)
	assert.Equal(t, uint32(0), stats.WorkersCount)
	assert.Equal(t, uint32(0), stats.Idlers)
	assert.Equal(t, uint32(0), stats.TasksInQueue)
}

func TestWorkerPool_ShouldUseCustomRecover(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	opts := DefaultOptions
	var got string
	opts.RecoveryFn = func() {
		if msg := recover(); msg != nil {
			got = fmt.Sprintf("%v", msg)
		}
	}
	wp := NewWorkerPool(opts)
	wp.Run(ctx)
	wp.Add(func() error {
		panic("oops! panic!")
	})
	time.Sleep(1 * time.Second)
	cancel()
	<-wp.Done()
	assert.Equal(t, "oops! panic!", got)
}

func sleepSecTask(id, sec int) func() error {
	return func() error {
		fmt.Printf("--task(%d)-- starts to sleep %d sec in %s\n", id, sec, time.Now().Format("15:04:05"))
		time.Sleep(time.Duration(sec) * time.Second)
		fmt.Printf("--task(%d)-- stops to sleep %d sec in %s\n", id, sec, time.Now().Format("15:04:05"))
		return nil
	}
}

func printStats(wp *WorkerPool) Stats {
	stats := wp.Stats()
	fmt.Printf("--stats-- workers count %d\n", stats.WorkersCount)
	fmt.Printf("--stats-- idle workers %d\n", stats.Idlers)
	fmt.Printf("--stats-- tasks in queue %d\n", stats.TasksInQueue)
	return stats
}
