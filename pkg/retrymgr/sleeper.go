package retrymgr

import (
	"context"
	"time"
)

//go:generate go tool mockgen -destination=retrymgrtest/mock_sleeper.go -package retrymgrtest github.com/bq2cd/yp-go-metrics/pkg/retrymgr Sleeper

// Sleeper provides a context-aware way to wait for a given delay.
type Sleeper interface {
	Sleep(ctx context.Context, delay time.Duration) error
}

type sleeper struct{}

// NewSleeper creates an instance of the sleeper.
func NewSleeper() *sleeper {
	return &sleeper{}
}

// Sleep waits for the specified delay and returns nil unless
// a context was canceled. In this case an error returned.
func (s *sleeper) Sleep(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
