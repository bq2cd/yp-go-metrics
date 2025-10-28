package retrymgr

import (
	"time"
)

//go:generate go tool mockgen -destination=retrymgrtest/mock_strategy.go -package retrymgrtest github.com/bq2cd/yp-go-metrics/pkg/retrymgr Strategy

// Strategy represents an algorithm to use with [Retrier] to provide
// a series of delays between executions.
type Strategy interface {
	Name() string
	NextDelay() (time.Duration, bool)
}

// CreateStrategyFn defines a function for [Strategy] creation.
type CreateStrategyFn func() Strategy

type strategy1s3s5s struct {
	steps   [3]int
	current int
}

// NewStrategy1s3s5s defines a simple static strategy that retries
// no more than 3 times, after 1 second, 3 seconds, 5 seconds.
func NewStrategy1s3s5s() Strategy {
	return &strategy1s3s5s{
		steps:   [3]int{1, 3, 5},
		current: 0,
	}
}

// Name returns the name of the strategy for logging purposes.
func (s *strategy1s3s5s) Name() string {
	return "1s3s5s"
}

// NextDelay returns current delay and `true` if the strategy is active, and `0` and `false` if the strategy has finished.
func (s *strategy1s3s5s) NextDelay() (time.Duration, bool) {
	if s.current >= len(s.steps) {
		return 0, false
	}
	delay := time.Duration(s.steps[s.current]) * time.Second
	s.current++
	return delay, true
}
