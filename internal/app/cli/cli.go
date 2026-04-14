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

type App[C any] struct {
	Name          string
	ParseArgs     func(*flag.FlagSet, []string, envparser.Parser) (C, error)
	LaunchProcess func(context.Context, log.Logger, C) error
}

func (a App[C]) Run(baseCtx context.Context, logger log.Logger, args []string, stderr io.Writer) error {
	cfg, profiler, err := a.parseArgs(args, stderr)
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

func (a App[C]) parseArgs(args []string, stderr io.Writer) (C, *profiler, error) {
	fs := flag.NewFlagSet(a.Name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	profiler := newProfiler()
	profiler.AddProfilingArgs(fs)

	cfg, err := a.ParseArgs(fs, args, envparser.NewParser())
	if err != nil {
		return cfg, profiler, fmt.Errorf("unable to parse args: %w", err)
	}

	return cfg, profiler, err
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
