package agent

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/periodictask"
)

type agent struct {
	logger    log.Logger
	config    config.Config
	collector Collector
	reporter  Reporter
}

// New creates an instance of an agent process which runs
// metrics collector and reporter.
func New(logger log.Logger, cfg config.Config, collector Collector, reporter Reporter) *agent {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	return &agent{
		logger:    logger.With(log.Str("subsystem", "agent")),
		config:    cfg,
		collector: collector,
		reporter:  reporter,
	}
}

// runPeriodicTask executes given task function at a given interval
// with optionally different initial delay before the first execution.
// This is a blocking call and is typically used in a goroutine.
// Supports cancellation via context.
func runPeriodicTask(ctx context.Context, interval time.Duration, taskFn func(context.Context) error, initialDelay time.Duration) error {
	t := periodictask.NewTimerTask(interval, taskFn, initialDelay)
	return t.Run(ctx)
}

func (a *agent) doReport(baseCtx context.Context) error {
	ctx, cancel := context.WithTimeout(baseCtx, a.config.ReportInterval)
	defer cancel()

	outCh, err := a.collector.Snapshot(ctx)
	if err != nil {
		return err
	}
	return a.reporter.Report(ctx, outCh)
}

func (a *agent) launchCollector(ctx context.Context, grp *errgroup.Group, errCh chan<- error) {
	grp.Go(func() error {
		err := runPeriodicTask(ctx, a.config.PollInterval, a.collector.Collect, 0)
		errCh <- err

		return err
	})
}

func (a *agent) launchReporter(ctx context.Context, grp *errgroup.Group, errCh chan<- error) {
	grp.Go(func() error {
		err := runPeriodicTask(ctx, a.config.ReportInterval, a.doReport, a.config.ReportInterval)
		errCh <- err

		return err
	})
}

// Run launches main loop of the agent worker:
// collecting metrics and reporting them back to a server.
func (a *agent) Run(ctx context.Context) error {
	grp := new(errgroup.Group)
	errCh := make(chan error, 2)

	a.launchCollector(ctx, grp, errCh)
	a.launchReporter(ctx, grp, errCh)

	// we need to collect individual errors for tests
	_ = grp.Wait()
	close(errCh)

	var errFinal error
	for err := range errCh {
		errFinal = errors.Join(errFinal, err)
	}

	return errFinal
}
