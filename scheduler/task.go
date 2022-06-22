package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Action func(ctx context.Context, args ...interface{})

type Task struct {
	id       uuid.UUID
	actionFn Action
	args     []interface{}
	timeout  time.Duration
	done     chan struct{}
}

func NewTask(actionFn Action, timeout time.Duration, args ...interface{}) Task {
	return Task{
		id:       uuid.New(),
		actionFn: actionFn,
		timeout:  timeout,
		args:     args,
		done:     make(chan struct{}),
	}
}

func (t Task) GetId() [16]byte {
	return t.id
}
