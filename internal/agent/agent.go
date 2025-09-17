package agent

import (
	"context"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
)

type agentWorker struct {
	context   context.Context
	config    config.Config
	collector Collector
	storer    Storer
	reporter  Reporter
}

// NewAgent creates an instance of an agent worker.
func NewAgent(ctx context.Context, cfg config.Config, collector Collector, storer Storer, reporter Reporter) *agentWorker {
	return &agentWorker{context: ctx, config: cfg, collector: collector, storer: storer, reporter: reporter}
}

// Run launches main loop of the agent worker.
func (a *agentWorker) Run() error {
	return nil
}
