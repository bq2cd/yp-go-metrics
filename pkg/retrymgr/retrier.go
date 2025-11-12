package retrymgr

import (
	"context"
	"errors"
	"fmt"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

// RetryableFn defines a function that is called by [Retrier.Do].
type RetryableFn[T any] func(context.Context) (T, error)

// ShouldRetryFn defines a function to determine if further retries are necessary based on the given error.
type ShouldRetryFn func(error) bool

// Retrier provides a way to retry given task with a conditional
// [ShouldRetryFn] gateway.
type Retrier[T any] interface {
	Do(ctx context.Context, taskName string, taskFn RetryableFn[T], shouldRetryFn ShouldRetryFn) (T, error)
}

type retrier[T any] struct {
	logger   log.Logger
	strategy Strategy
	sleeper  Sleeper
}

// New creates an instance of [Retrier] with given [Strategy] and [Sleeper].
func NewRetrier[T any](factory RetrierFactory) *retrier[T] {
	strategy := factory.GetStrategy()
	return &retrier[T]{
		logger:   factory.GetLogger().With(log.Str("component", "retrier"), log.Str("strategy", strategy.Name())),
		strategy: strategy,
		sleeper:  factory.GetSleeper(),
	}
}

func (r *retrier[T]) Do(ctx context.Context, taskName string, taskFn RetryableFn[T], shouldRetryFn ShouldRetryFn) (result T, err error) {
	logger := r.logger.With(log.Str("task", taskName))

	attempts := 0

	for {
		result, err = taskFn(ctx)
		if err == nil {
			return
		}

		if !shouldRetryFn(err) {
			return
		}

		delay, active := r.strategy.NextDelay()
		if !active {
			return
		}

		attempts++

		l := logger.With(log.Int("attempt", attempts), log.Dur("delay", delay))
		l.Info().Err("previous_error", err).Msg("retrying task")

		errSleep := r.sleeper.Sleep(ctx, delay)
		if errSleep != nil {
			err = errors.Join(err, fmt.Errorf("sleeper error (attempt=%d, delay=%v): %w", attempts, delay, errSleep))
			l.Info().WithErr(errSleep).Msg("aborting execution due to sleeper error")
			return
		}
	}
}
