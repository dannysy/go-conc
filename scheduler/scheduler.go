package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	wg    sync.WaitGroup
	tasks map[uuid.UUID]Task
}

func NewScheduler() *Scheduler {
	return &Scheduler{tasks: map[uuid.UUID]Task{}, wg: sync.WaitGroup{}}
}

func (s *Scheduler) Start(ctx context.Context, task Task) {
	s.wg.Add(1)
	ticker := time.NewTicker(task.timeout)
	s.tasks[task.id] = task
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-ticker.C:
				task.actionFn(ctx, task.args...)
			case <-ctx.Done():
				delete(s.tasks, task.id)
				return
			case <-task.done:
				return
			}
		}
	}()
}

func (s *Scheduler) Once(ctx context.Context, taskId [16]byte) (ok bool) {
	task, ok := s.tasks[taskId]
	if !ok {
		return false
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		task.actionFn(ctx, task.args...)
	}()
	return true
}

func (s *Scheduler) Stop(taskId [16]byte) (ok bool) {
	task, ok := s.tasks[taskId]
	if !ok {
		return false
	}
	close(task.done)
	delete(s.tasks, taskId)
	return true
}

func (s *Scheduler) Close() {
	for _, task := range s.tasks {
		close(task.done)
		delete(s.tasks, task.id)
	}
	s.wg.Wait()
}
