package periodictask

import (
	"context"
	"errors"
	"time"
)

// TaskTimerFn represents task's actual workload.
type TaskTimerFn func(context.Context) error

type timerTask struct {
	taskFn       TaskTimerFn
	interval     time.Duration
	initialDelay time.Duration
}

// NewTimerTask creates an instance of a task that trigger provided task function
// every given interval, with possible initial delay before the first execution.
func NewTimerTask(interval time.Duration, taskFn TaskTimerFn, initialDelay time.Duration) *timerTask {
	return &timerTask{
		taskFn:       taskFn,
		interval:     interval,
		initialDelay: initialDelay,
	}
}

// Run starts task execution using provided context as a means for cancellation.
// The execution is synchronous, so it is the caller's responsibility to use
// goroutines if asynchronous execution is required.
func (t *timerTask) Run(ctx context.Context) error {
	var (
		errFinal error
	)

	start := time.Now()
	timer := time.NewTimer(t.initialDelay)

loop:
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			break loop
		case <-timer.C:
			errFinal = errors.Join(errFinal, t.taskFn(ctx))
			elapsedIntervals := int(time.Since(start) / t.interval)
			nextInvocation := start.Add(t.interval * time.Duration(elapsedIntervals+1))
			timer.Reset(time.Until(nextInvocation))
		}
	}

	return errFinal
}
