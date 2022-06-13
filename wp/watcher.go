package wp

import (
	"context"
	"sync/atomic"
	"time"
)

type watcher struct {
	ctx  context.Context
	pool *WorkerPool
}

func (w *watcher) watch() {
	ticker := time.NewTicker(w.pool.opts.WatcherPeriod)
	for {
		select {
		case <-ticker.C:
			w.addMoreWorkers()
			w.stopExcessIdlers()
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *watcher) addMoreWorkers() {
	workersCount := atomic.LoadUint32(&w.pool.workersCount)
	tasksCount := len(w.pool.taskCh)
	poolSize := w.pool.opts.Size
	if tasksCount != 0 && workersCount <= poolSize {
		canAdd := poolSize - workersCount
		if int(canAdd) > tasksCount {
			w.pool.addWorkers(w.ctx, tasksCount)
		} else {
			w.pool.addWorkers(w.ctx, int(canAdd))
		}
	}
}

func (w *watcher) stopExcessIdlers() {
	idle := make([]int64, 0, w.pool.opts.Size)
	w.pool.workers.Range(func(key, value any) bool {
		wrk := value.(*worker)
		if wrk.idle {
			idle = append(idle, wrk.id)
		}
		return true
	})
	excessCount := len(idle) - int(w.pool.opts.Idle)
	for i := 0; i < excessCount; i++ {
		value, ok := w.pool.workers.LoadAndDelete(idle[i])
		if ok {
			wrk := value.(*worker)
			wrk.stop()
		}
	}
}
