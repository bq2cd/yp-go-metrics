package agent

import (
	"context"
	"errors"
	"sync"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/periodictask"
)

type agent struct {
	config    config.Config
	collector Collector
	reporter  Reporter
}

// New creates an instance of an agent process which runs
// metrics collector and reporter.
func New(cfg config.Config, collector Collector, reporter Reporter) *agent {
	return &agent{config: cfg, collector: collector, reporter: reporter}
}

// runPeriodicTask executes given task function at a given interval
// with optionally different initial delay before the first execution.
// This is a blocking call and is typically used in a goroutine.
// Supports cancellation via context.
func runPeriodicTask(ctx context.Context, interval time.Duration, taskFn func(context.Context) error, initialDelay time.Duration) error {
	t := periodictask.NewTimerTask(interval, taskFn, initialDelay)
	return t.Run(ctx)
}

func (a *agent) doReport(ctx context.Context) error {
	var errFinal error
	metrics, err := a.collector.Snapshot(ctx)
	errFinal = errors.Join(errFinal, err, a.reporter.Report(ctx, metrics))
	return errFinal
}

// Run launches main loop of the agent worker:
// collecting metrics and reporting them back to a server.
func (a *agent) Run(ctx context.Context) error {
	var (
		errRun error
		wg     sync.WaitGroup
	)

	errCh := make(chan error, 2)

	// launch poller (with first poll happening without delay)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- runPeriodicTask(ctx, a.config.PollInterval, a.collector.Collect, 0)
	}()

	// launch reporter
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- runPeriodicTask(ctx, a.config.ReportInterval, a.doReport, a.config.ReportInterval)
	}()

	// wait for poller and reporter in a goroutine
	// in order to consume from errCh in the main thread.
	// NB. both goroutines will stop when a.context is cancelled.
	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		errRun = errors.Join(errRun, err)
	}

	return errRun
}
