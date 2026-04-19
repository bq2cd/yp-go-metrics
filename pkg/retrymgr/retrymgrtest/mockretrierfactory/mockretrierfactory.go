package mockretrierfactory

import (
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest"
)

// MockRetrierFactory is a dummy [retrymgr.RetrierFactory] that contains only mock
// implementations.
type MockRetrierFactory struct {
	Logger   log.TestLogger
	Sleeper  *retrymgrtest.MockSleeper
	Strategy *retrymgrtest.MockStrategy
}

// NewMockRetrierFactory creates an instance of [MockRetrierFactor].
func NewMockRetrierFactory(ctrl *gomock.Controller) *MockRetrierFactory {
	return &MockRetrierFactory{
		Logger:   log.NewTestLogger(),
		Sleeper:  retrymgrtest.NewMockSleeper(ctrl),
		Strategy: retrymgrtest.NewMockStrategy(ctrl),
	}
}

// GetLogger returns test logger instance ([log.TestLogger]).
func (m *MockRetrierFactory) GetLogger() log.Logger {
	return m.Logger
}

// GetSleeper returns mock implementation of [retrymgr.Sleeper] ([retrymgrtest.MockSleeper]).
func (m *MockRetrierFactory) GetSleeper() retrymgr.Sleeper {
	return m.Sleeper
}

// GetStrategy returns mock implementation of [retrymgr.Strategy] ([retrymgrtest.MockStrategy]).
func (m *MockRetrierFactory) GetStrategy() retrymgr.Strategy {
	return m.Strategy
}
