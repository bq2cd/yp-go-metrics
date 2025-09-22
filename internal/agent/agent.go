package agent

import (
	"context"
	"errors"
	"sync"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
)

type agentWorker struct {
	context   context.Context
	config    config.Config
	collector Collector
	reporter  Reporter
}

// NewAgent creates an instance of an agent worker.
func NewAgent(ctx context.Context, cfg config.Config, collector Collector, reporter Reporter) *agentWorker {
	return &agentWorker{context: ctx, config: cfg, collector: collector, reporter: reporter}
}

// Run launches main loop of the agent worker:
// collecting metrics and reporting them back to a server.
func (a *agentWorker) Run() error {
	var (
		errRun error
		wg     sync.WaitGroup
		mu     sync.Mutex
	)

	doPoll := func() {
		errRun = errors.Join(errRun, a.collector.Collect())
	}

	pollTicker := time.NewTicker(a.config.PollInterval)
	reportTicker := time.NewTicker(a.config.ReportInterval)

	errCh := make(chan error, 2)

	// Perform first poll immediately rather than waiting for
	// the poll interval to pass.
	doPoll()

loop:
	for {
		select {
		case <-pollTicker.C:
			doPoll()
		case <-reportTicker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				if mu.TryLock() {
					defer mu.Unlock()
				} else {
					return
				}
				metrics, err := a.collector.Snapshot()
				errCh <- err
				errCh <- a.reporter.Report(metrics)
			}()
		case err := <-errCh:
			errRun = errors.Join(errRun, err)
		case <-a.context.Done():
			go func() {
				wg.Wait()
				close(errCh)
			}()
			for err := range errCh {
				errRun = errors.Join(errRun, err)
			}
			break loop
		}
	}

	return errRun
}
