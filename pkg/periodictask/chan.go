package periodictask

import (
	"context"
	"errors"
)

// TaskChanFn represents task's actual workload.
type TaskChanFn[T any] func(context.Context, T) error

type chanTask[T any] struct {
	incoming <-chan T
	taskFn   TaskChanFn[T]
}

// NewChanTask creates an instance of a task that triggers provided task function
// on incoming values to a channel.
func NewChanTask[T any](incoming <-chan T, taskFn TaskChanFn[T]) *chanTask[T] {
	return &chanTask[T]{
		incoming: incoming,
		taskFn:   taskFn,
	}
}

// Run starts task execution using provided context as a means for cancellation.
// The execution is synchronous, so it is the caller's responsibility to use
// goroutines if asynchronous execution is required.
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
