package periodictask

import (
	"context"
	"errors"
)

type TaskChanFn[T any] func(context.Context, T) error

type chanTask[T any] struct {
	context  context.Context
	incoming <-chan T
	taskFn   TaskChanFn[T]
}

func NewChanTask[T any](ctx context.Context, incoming <-chan T, taskFn TaskChanFn[T]) *chanTask[T] {
	return &chanTask[T]{
		context:  ctx,
		incoming: incoming,
		taskFn:   taskFn,
	}
}

func (t *chanTask[T]) Run() error {
	var (
		errFinal error
	)

loop:
	for {
		// extra check on context cancellation to ensure
		// we have not picked extra work
		select {
		case <-t.context.Done():
			break loop
		default:
		}

		// normal select
		select {
		case <-t.context.Done():
			break loop
		case v := <-t.incoming:
			errFinal = errors.Join(errFinal, t.taskFn(t.context, v))
		}
	}

	return errFinal
}
