[![Go](https://github.com/dannysy/go-conc/actions/workflows/go.yml/badge.svg)](https://github.com/dannysy/go-conc/actions/workflows/go.yml)
![GoCoverage](https://img.shields.io/badge/Coverage-80.7%25-green)
# go-conc
Concurrency patterns realization in Golang

# Worker Pool

Collect Tasks and executes them into routines concurrently. 
Has a watchdog routine that looks after worker routines and tasks queue (channel) length. 
If task queue accumulating tasks then watchdog creates new workers but no more than pool size.
On the over side if where is idle workers which size is more than allowed in pool settings 
than watchdog kills exceed idlers. 

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