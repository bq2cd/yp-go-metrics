package periodictask

import (
	"context"
	"errors"
)

type TaskChanFn[T any] func(context.Context, T) error

type chanTask[T any] struct {
	incoming <-chan T
	taskFn   TaskChanFn[T]
}

func NewChanTask[T any](incoming <-chan T, taskFn TaskChanFn[T]) *chanTask[T] {
	return &chanTask[T]{
		incoming: incoming,
		taskFn:   taskFn,
	}
}

func (t *chanTask[T]) Run(ctx context.Context) error {
	var (
		errFinal error
	)

loop:
	for {
		// extra check on context cancellation to ensure
		// we have not picked extra work
		select {
		case <-ctx.Done():
			break loop
		default:
		}

		// normal select
		select {
		case <-ctx.Done():
			break loop
		case v := <-t.incoming:
			errFinal = errors.Join(errFinal, t.taskFn(ctx, v))
		}
	}

	return errFinal
}
