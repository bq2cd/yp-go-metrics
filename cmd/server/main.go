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

func runServer(ctx context.Context, cfg config.Config) error {
	storage := repository.NewMemStorage()
	svc := service.NewMetrics(storage)
	router := handler.NewRouter(svc, nil)

	srv := server.NewServer(ctx, cfg, router)

	return srv.Run()
}

func parseArgs(fs *flag.FlagSet, args []string) (config.Config, error) {
	var (
		listenAddress   string
		shutdownTimeout uint
	)

	fs.StringVar(&listenAddress, "a", defaultAddress, "listen address in the format [HOST]:PORT")
	fs.UintVar(&shutdownTimeout, "t", defaultShutdownTimeoutSec, "graceful shutdown timeout in seconds")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, fmt.Errorf("invalid args: %w", err)
	}

	cfg, err := config.New(
		config.ListenAddress(listenAddress),
		config.ShutdownTimeout(shutdownTimeout),
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

	cfg, err := parseArgs(fs, args)
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
