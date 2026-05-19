package e2e

import (
	"context"
	"time"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
)

// NoopRetrierStrategy implements [retrymgr.Strategy] that does not retry at all.
type NoopRetrierStrategy struct{}

// NoopSleeper implements [retrymgr.Sleeper] that returns immediately.
type NoopSleeper struct{}

// Name returns name of the [NoopRetrierStrategy].
func (s *NoopRetrierStrategy) Name() string {
	return "noop_strategy"
}

// NextDelay returns `0` and `false`, implying that the strategy has already finished.
func (s *NoopRetrierStrategy) NextDelay() (time.Duration, bool) {
	return 0, false
}

// Sleep returns `nil` immediately, no actual sleeping occurs.
func (s *NoopSleeper) Sleep(_ context.Context, _ time.Duration) error {
	return nil
}

// NewNoopRetrierFactory creates a [retrymgr.RetrierFactory] that returns [NoopRetrierStrategy] and [NoopSleeper],
// which effectively do nothing (no retries, no sleep).
func NewNoopRetrierFactory() retrymgr.RetrierFactory {
	return retrymgr.NewRetrierFactory(log.NewNoopLogger(), new(NoopSleeper), func() retrymgr.Strategy { return new(NoopRetrierStrategy) })
}
