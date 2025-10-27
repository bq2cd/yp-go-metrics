package periodictask

import (
	"context"
	"errors"
	"time"
)

type TaskTimerFn func(context.Context) error

type timerTask struct {
	taskFn       TaskTimerFn
	interval     time.Duration
	initialDelay time.Duration
}

func NewTimerTask(interval time.Duration, taskFn TaskTimerFn, initialDelay time.Duration) *timerTask {
	return &timerTask{
		taskFn:       taskFn,
		interval:     interval,
		initialDelay: initialDelay,
	}
}

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
			next := start.Add(t.interval * ((time.Since(start) / t.interval) + 1))
			timer.Reset(time.Until(next))
		}
	}

	return errFinal
}
