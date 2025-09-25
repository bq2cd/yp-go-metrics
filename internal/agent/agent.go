package agent

import (
	"context"
	"errors"
	"sync"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
)

type agent struct {
	context   context.Context
	config    config.Config
	collector Collector
	reporter  Reporter
}

// NewAgent creates an instance of an agent worker.
func NewAgent(ctx context.Context, cfg config.Config, collector Collector, reporter Reporter) *agent {
	return &agent{context: ctx, config: cfg, collector: collector, reporter: reporter}
}

// RunPeriodicTask executes given task function at a given interval
// with optionally different initial delay before the first execution.
// This is a blocking call and is typically used in a goroutine.
// Supports cancellation via context.
func RunPeriodicTask(ctx context.Context, interval time.Duration, taskFunc func() error, initialDelay time.Duration) error {
	var (
		errFinal error
	)

	start := time.Now()
	timer := time.NewTimer(initialDelay)

loop:
	for {
		select {
		case <-timer.C:
			errFinal = errors.Join(errFinal, taskFunc())
			next := start.Add(interval * ((time.Since(start) / interval) + 1))
			timer.Reset(time.Until(next))
		case <-ctx.Done():
			timer.Stop()
			break loop
		}
	}

	return errFinal
}

func (a *agent) doReport() error {
	var errFinal error
	metrics, err := a.collector.Snapshot()
	errFinal = errors.Join(errFinal, err, a.reporter.Report(metrics))
	return errFinal
}

// Run launches main loop of the agent worker:
// collecting metrics and reporting them back to a server.
func (a *agent) Run() error {
	var (
		errRun error
		wg     sync.WaitGroup
	)

	errCh := make(chan error, 2)

	// launch poller (with first poll happening without delay)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- RunPeriodicTask(a.context, a.config.PollInterval, a.collector.Collect, 0)
	}()

	// launch reporter
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- RunPeriodicTask(a.context, a.config.ReportInterval, a.doReport, a.config.ReportInterval)
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
