package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShouldRunTaskOnce(t *testing.T) {
	ctx := context.Background()
	sdlr := NewScheduler()
	var counter atomic.Value
	task := NewTask(func(ctx context.Context, args ...interface{}) {
		fmt.Println("--task-- executing action")
		c := args[0].(*atomic.Value)
		c.Store(1)
	}, time.Millisecond, &counter)
	sdlr.Start(ctx, task)
	sdlr.Once(ctx, task.GetId())
	time.Sleep(time.Millisecond * 2)
	assert.Equal(t, 1, counter.Load())
}

func TestShouldStopTaskByContext(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	sdlr := NewScheduler()
	task := NewTask(func(ctx context.Context, args ...interface{}) {
		fmt.Println("--task-- executing action")
	}, time.Millisecond*3)
	sdlr.Start(ctx, task)
	cancel()
	time.Sleep(time.Millisecond * 1)
	_, ok := sdlr.tasks[task.id]
	assert.False(t, ok)
}

func TestShouldStopTaskByClose(t *testing.T) {
	ctx := context.Background()
	sdlr := NewScheduler()
	task := NewTask(func(ctx context.Context, args ...interface{}) {
		fmt.Println("--task-- executing action")
	}, time.Millisecond*3)
	sdlr.Start(ctx, task)
	sdlr.Close()
	_, ok := sdlr.tasks[task.id]
	assert.False(t, ok)
}

func TestShouldStopTaskByStop(t *testing.T) {
	ctx := context.Background()
	sdlr := NewScheduler()
	task := NewTask(func(ctx context.Context, args ...interface{}) {
		fmt.Println("--task-- executing action")
	}, time.Millisecond*3)
	sdlr.Start(ctx, task)
	sdlr.Stop(task.id)
	_, ok := sdlr.tasks[task.id]
	assert.False(t, ok)
}
