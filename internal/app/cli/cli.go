package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

// App represents an application entry point with given name and hook functions
// to parse CLI flags and launch app's main process.
type App[C any] struct {
	Name          string
	ParseArgs     func(*flag.FlagSet, []string, envparser.Parser) (C, error)
	LaunchProcess func(context.Context, log.Logger, C) error
}

// Run parses CLI flags and environment variables, populates app's config, conditionally enables
// profiling, and start app's main process.
// It is also configures OS signal handling (SIGINT, SIGTERM) via [context.Context] - the app is
// expected to handle context cancellation and finish its execution gracefully.
// Currently, there is no enforcement for apps that might ignore context cancellation or
// not handle it properly.
func (a App[C]) Run(baseCtx context.Context, logger log.Logger, args []string, stderr io.Writer) error {
	profiler := newProfiler(logger)

	cfg, err := a.populateConfig(profiler, args, stderr)
	if err != nil {
		return err
	}

	stopProfiling, err := profiler.MaybeStartProfiling()
	if err != nil {
		return fmt.Errorf("unable to start profiling: %w", err)
	}
	defer stopProfiling()

	ctx, stop := signal.NotifyContext(baseCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return a.run(ctx, logger, cfg)
}

func (a App[C]) populateConfig(profiler *profiler, args []string, stderr io.Writer) (C, error) {
	fs := flag.NewFlagSet(a.Name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	profiler.AddProfilingArgs(fs)

	cfg, err := a.ParseArgs(fs, args, envparser.NewParser())
	if err != nil {
		return cfg, fmt.Errorf("unable to parse args: %w", err)
	}

	return cfg, err
}

func (a App[C]) run(ctx context.Context, logger log.Logger, cfg C) error {
	err := a.LaunchProcess(ctx, logger, cfg)
	select {
	case <-ctx.Done():
		logger.Info().Msg("received termination signal")
	default:
	}

	return err
}
