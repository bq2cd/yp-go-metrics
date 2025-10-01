package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/server"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

const (
	defaultAddress            = "localhost:8080"
	defaultShutdownTimeoutSec = 1
)

type cliOptions struct {
	ListenAddress   string `env:"ADDRESS"`
	ShutdownTimeout uint   `env:"SHUTDOWN_TIMEOUT"`
}

func runServer(ctx context.Context, cfg config.Config) error {
	storage := repository.NewMemStorage()
	svc := service.NewMetrics(storage)
	router := handler.NewRouter(svc, nil)

	srv := server.NewServer(ctx, cfg, router)

	return srv.Run()
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (config.Config, error) {
	var opts cliOptions

	fs.StringVar(&opts.ListenAddress, "a", defaultAddress, "listen address in the format [HOST]:PORT")
	fs.UintVar(&opts.ShutdownTimeout, "t", defaultShutdownTimeoutSec, "graceful shutdown timeout in seconds")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, fmt.Errorf("invalid args: %w", err)
	}

	if err := envParser.Parse(&opts); err != nil {
		return config.Config{}, fmt.Errorf("invalid env vars: %w", err)
	}

	cfg, err := config.New(
		config.ListenAddress(opts.ListenAddress),
		config.ShutdownTimeout(opts.ShutdownTimeout),
	)
	if err != nil {
		return config.Config{}, fmt.Errorf("unable to construct config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return config.Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return *cfg, nil
}

func run(ctx context.Context, args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(stderr)

	cfg, err := parseArgs(fs, args, envparser.NewParser())
	if err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = runServer(ctx, cfg)
	<-ctx.Done()

	return err
}

func main() {
	err := run(context.Background(), os.Args[1:], os.Stderr)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
