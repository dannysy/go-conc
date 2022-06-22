package wp

import "context"

type Action func(ctx context.Context, args ...interface{}) (interface{}, error)

type Task struct {
	actionFn Action
	args     []interface{}
}

func NewTask(actionFn Action, args ...interface{}) Task {
	return Task{
		actionFn: actionFn,
		args:     args,
	}
}
