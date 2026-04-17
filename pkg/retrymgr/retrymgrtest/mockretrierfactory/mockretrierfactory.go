package mockretrierfactory

import (
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest"
)

type MockRetrierFactory struct {
	Logger   log.TestLogger
	Sleeper  *retrymgrtest.MockSleeper
	Strategy *retrymgrtest.MockStrategy
}

func NewMockRetrierFactory(ctrl *gomock.Controller) *MockRetrierFactory {
	return &MockRetrierFactory{
		Logger:   log.NewTestLogger(),
		Sleeper:  retrymgrtest.NewMockSleeper(ctrl),
		Strategy: retrymgrtest.NewMockStrategy(ctrl),
	}
}

func (m *MockRetrierFactory) GetLogger() log.Logger {
	return m.Logger
}

func (m *MockRetrierFactory) GetSleeper() retrymgr.Sleeper {
	return m.Sleeper
}

func (m *MockRetrierFactory) GetStrategy() retrymgr.Strategy {
	return m.Strategy
}
