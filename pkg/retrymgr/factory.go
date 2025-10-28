package retrymgr

import (
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

// RetrierFactory provides necessary parameters for [Retrier] creation.
type RetrierFactory interface {
	GetLogger() log.Logger
	GetSleeper() Sleeper
	GetStrategy() Strategy
}

type retrierFactory struct {
	logger          log.Logger
	sleeper         Sleeper
	strategyCreator CreateStrategyFn
}

// GetLogger returns an instance of the logger.
func (f *retrierFactory) GetLogger() log.Logger {
	return f.logger
}

// GetSleeper returns an instance of the sleeper to use between retries.
func (f *retrierFactory) GetSleeper() Sleeper {
	return f.sleeper
}

// GetStrategy returns an instance of retry strategy.
func (f *retrierFactory) GetStrategy() Strategy {
	return f.strategyCreator()
}

// NewRetrierFactory creates an instance of the factory to store
// provided logger, sleeper, and strategy creator function.
func NewRetrierFactory(logger log.Logger, sleeper Sleeper, strategyCreator CreateStrategyFn) *retrierFactory {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	return &retrierFactory{
		logger:          logger,
		sleeper:         sleeper,
		strategyCreator: strategyCreator,
	}
}
