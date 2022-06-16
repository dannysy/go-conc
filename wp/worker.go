package wp

import (
	"context"
)

type worker struct {
	ctx        context.Context
	id         int64
	taskCh     <-chan Task
	resultCh   chan<- Result
	recoveryFn func()
	idle       bool
	done       chan struct{}
}

func (w *worker) run() {
	defer w.recoveryFn()
	w.done = make(chan struct{})
	for {
		select {
		case t := <-w.taskCh:
			w.idle = false
			result, err := t.actionFn(w.ctx, t.args...)
			w.resultCh <- Result{
				value: result,
				err:   err,
			}
			w.idle = true
		case <-w.ctx.Done():
			return
		case <-w.done:
			return
		}
	}
}

func (w *worker) stop() {
	if w.done == nil {
		return
	}
	close(w.done)
}
