package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/internal/log"
)

type App[C any] struct {
	Name          string
	ParseArgs     func(*flag.FlagSet, []string, envparser.Parser) (C, error)
	LaunchProcess func(context.Context, log.Logger, C) error
}

func (a App[C]) Run(ctx context.Context, logger log.Logger, args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet(a.Name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	cfg, err := a.ParseArgs(fs, args, envparser.NewParser())
	if err != nil {
		return fmt.Errorf("unable to parse args: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err = a.LaunchProcess(ctx, logger, cfg)
	select {
	case <-ctx.Done():
		logger.Info().Msg("received termination signal")
	default:
	}

	return err
}
