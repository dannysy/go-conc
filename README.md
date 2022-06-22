[![Go](https://github.com/dannysy/go-conc/actions/workflows/go.yml/badge.svg)](https://github.com/dannysy/go-conc/actions/workflows/go.yml)
![GoCoverage](https://img.shields.io/badge/Coverage-80.7%25-green)
# go-conc
Concurrency patterns realization in Golang

# Worker Pool

Collect Tasks and executes them into routines concurrently. Has a watchdog routine that looks after worker routines 
 and tasks queue (channel) length. If task queue accumulating tasks then watchdog creates new workers but no more 
 than pool size. On the over side if where is idle workers which quantity is more than allowed in pool settings 
 then watchdog kills excesed idlers.
```go
type Options struct {
	Size          uint32 // Pool size, max workers count
	TaskChSize    uint32 // Max task queue (channel) length
	ResultChSize  uint32 // Max result channel length
	Idle          uint32 // Max idle workers count
	RecoveryFn    func() // Custom recovery func to use in workers
	WatcherPeriod time.Duration // Watchdog action timeout
}
```
## Usage
```go
...
        ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)
	opts := wp.GetDefaultOptions()
	opts.RecoveryFn = func() {
		if msg := recover(); msg != nil {
			fmt.Println(msg)
			cancel()
		}
	}
	pool := wp.NewWorkerPool(opts)
	pool.Run(ctx)
	for i := 0; i < 10; i++ {
		pool.Add(wp.NewTask(
			func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return args[0].(int), nil
			}, i))
	}

	pool.Add(wp.NewTask(func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return nil, fmt.Errorf("error")
	}))
	pool.Add(wp.NewTask(func(ctx context.Context, args ...interface{}) (interface{}, error) {
		panic("oops! panic!")
	}))
	for {
		select {
		case <-pool.Done():
			return
		case result := <-pool.Result():
			if result.GetErr() != nil {
				fmt.Println(result.GetErr())
				continue
			}
			fmt.Println(result.GetValue())
		}
	}
...
```

# Scheduler

Collect Tasks and executes them periodically. Can remove Task from Scheduler by cancelling Task **Context** or
manually by calling Scheduler's **Stop** method.

## Usage
```go
    ...
    ctx := context.Background()
    sdlr := scheduler.NewScheduler()
    task := NewTask(func(ctx context.Context, args ...interface{}) {
    fmt.Println("task with 3ms execution period")
    }, time.Millisecond*3)
    sdlr.Start(ctx, task) // Add Task to Schedule, Scheduler will run Task after it's period
    sdlr.Once(ctx, task.GetId()) // Run Task once off schedule
    sdlr.Stop(ctx, task.GetId()) // Stop removes Task from schedule
    ctx, cancel := context.WithCancel(ctx)
    task2 := scheduler.NewTask(func(ctx context.Context, args ...interface{}) {
    fmt.Println("task with cancellable Context")
    }, time.Millisecond*3)
    sdlr.Start(ctx, task2)
    cancel() // Cancelling the Context will stop all Task that is Run with that Context
    sdlr.Close() // Stops all periodic Tasks started by the Scheduler
    ...
```