package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/server"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

const (
	defaultAddress         = "localhost:8080"
	defaultShutdownTimeout = 1 * time.Second
)

func runServer(ctx context.Context, cfg config.Config) error {
	storage := repository.NewMemStorage()
	svc := service.NewMetrics(storage)
	router := handler.NewRouter(svc, nil)

	srv := server.NewServer(ctx, cfg, router)

	return srv.Run()
}

func parseArgs(args []string) (config.Config, error) {
	var (
		listenAddress string
	)

	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	fs.StringVar(&listenAddress, "a", defaultAddress, "listen address in the format [HOST]:PORT")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, fmt.Errorf("invalid args: %w", err)
	}

	cfg := config.Config{}

	// validate listen address
	{
		parts := strings.Split(listenAddress, ":")
		if len(parts) > 2 {
			return config.Config{}, fmt.Errorf("invalid listen address")
		}
		cfg.ListenAddress = listenAddress
	}

	cfg.ShutdownTimeout = defaultShutdownTimeout

	return cfg, nil
}

func run(ctx context.Context, args []string) error {
	cfg, err := parseArgs(args)
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
	err := run(context.Background(), os.Args[1:])
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
