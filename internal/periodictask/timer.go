package periodictask

import (
	"context"
	"errors"
	"time"
)

type TaskTimerFn func(context.Context) error

type timerTask struct {
	context      context.Context
	taskFn       TaskTimerFn
	interval     time.Duration
	initialDelay time.Duration
}

func NewTimerTask(ctx context.Context, interval time.Duration, taskFn TaskTimerFn, initialDelay time.Duration) *timerTask {
	return &timerTask{
		context:      ctx,
		taskFn:       taskFn,
		interval:     interval,
		initialDelay: initialDelay,
	}
}

func (t *timerTask) Run() error {
	var (
		errFinal error
	)

	start := time.Now()
	timer := time.NewTimer(t.initialDelay)

loop:
	for {
		select {
		case <-t.context.Done():
			timer.Stop()
			break loop
		case <-timer.C:
			errFinal = errors.Join(errFinal, t.taskFn(t.context))
			next := start.Add(t.interval * ((time.Since(start) / t.interval) + 1))
			timer.Reset(time.Until(next))
		}
	}

	return errFinal
}
