package wp

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Task func() interface{}

type Stats struct {
	WorkersCount uint32
	Idlers       uint32
	TasksInQueue uint32
}

type WorkerPool struct {
	opts         Options
	taskCh       chan Task
	done         chan struct{}
	workersCount uint32
	workers      sync.Map
	wg           sync.WaitGroup
}

func NewWorkerPool(opts Options) *WorkerPool {
	if opts.Idle == 0 {
		opts.Idle = 1
	}
	if opts.Size == 0 {
		runtime.GOMAXPROCS(0)
	}
	wp := WorkerPool{
		done:    make(chan struct{}),
		opts:    opts,
		taskCh:  make(chan Task, opts.TaskChSize),
		workers: sync.Map{},
		wg:      sync.WaitGroup{},
	}
	return &wp
}

func (p *WorkerPool) Run(ctx context.Context) {
	go p.run(ctx)
}

func (p *WorkerPool) Add(task Task) {
	p.taskCh <- task
}

func (p *WorkerPool) SAdd(tasks ...Task) {
	for _, task := range tasks {
		p.Add(task)
	}
}

func (p *WorkerPool) Done() <-chan struct{} {
	return p.done
}

func (p *WorkerPool) Stats() (out Stats) {
	out.WorkersCount = atomic.LoadUint32(&p.workersCount)
	out.TasksInQueue = uint32(len(p.taskCh))
	p.workers.Range(func(key, value any) bool {
		wkr := value.(*worker)
		if wkr.idle {
			out.Idlers++
		}
		return true
	})
	return out
}

func (p *WorkerPool) run(ctx context.Context) {
	workersCount := len(p.taskCh)
	if workersCount > int(p.opts.Size) {
		workersCount = int(p.opts.Size)
	}
	if workersCount == 0 {
		workersCount = int(p.opts.Idle)
	}
	p.addWorkers(ctx, workersCount)
	p.addWatcher(ctx)
	p.wg.Wait()
	close(p.done)
}

func (p *WorkerPool) addWorkers(ctx context.Context, count int) {
	p.wg.Add(count)
	for i := 0; i < count; i++ {
		w := worker{
			ctx:        ctx,
			id:         time.Now().UnixNano(),
			taskCh:     p.taskCh,
			outFn:      p.opts.OutFn,
			recoveryFn: p.opts.RecoveryFn,
		}
		p.workers.Store(w.id, &w)
		atomic.AddUint32(&p.workersCount, 1)
		go func() {
			defer func() {
				atomic.AddUint32(&p.workersCount, ^uint32(0))
				p.workers.Delete(w.id)
				p.wg.Done()
			}()
			w.run()
		}()
	}
}

func (p *WorkerPool) addWatcher(ctx context.Context) {
	p.wg.Add(1)
	w := watcher{
		ctx:  ctx,
		pool: p,
	}
	go func() {
		defer p.wg.Done()
		w.watch()
	}()
}
