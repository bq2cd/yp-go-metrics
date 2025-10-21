package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bq2cd/yp-go-metrics/internal/app/cli"
	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	logger "github.com/bq2cd/yp-go-metrics/internal/app/logger"
	launcher "github.com/bq2cd/yp-go-metrics/internal/app/server"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
)

const (
	defaultAddress                  = "localhost:8080"
	defaultShutdownTimeoutSec       = 1
	defaultMetricStoreIntervalSec   = 300
	defaultMetricStoreFilePath      = "metrics.json"
	defaultMetricStoreLoadOnStartup = false
)

type cliOptions struct {
	ListenAddress            string `env:"ADDRESS"`
	ShutdownTimeout          uint   `env:"SHUTDOWN_TIMEOUT"`
	MetricStoreInterval      uint   `env:"STORE_INTERVAL"`
	MetricStoreFilePath      string `env:"FILE_STORAGE_PATH"`
	MetricStoreLoadOnStartup bool   `env:"RESTORE"`
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (config.Config, error) {
	var opts cliOptions

	fs.StringVar(&opts.ListenAddress, "a", defaultAddress, "listen address in the format [HOST]:PORT")
	fs.UintVar(&opts.ShutdownTimeout, "t", defaultShutdownTimeoutSec, "graceful shutdown timeout in seconds")
	fs.UintVar(&opts.MetricStoreInterval, "i", defaultMetricStoreIntervalSec, "dump metrics on disk each interval (in seconds)")
	fs.StringVar(&opts.MetricStoreFilePath, "f", defaultMetricStoreFilePath, "path to file for dumping metrics")
	fs.BoolVar(&opts.MetricStoreLoadOnStartup, "r", defaultMetricStoreLoadOnStartup, "restore metrics from file on startup")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, fmt.Errorf("invalid args: %w", err)
	}

	if err := envParser.Parse(&opts); err != nil {
		return config.Config{}, fmt.Errorf("invalid env vars: %w", err)
	}

	cfg, err := config.New(
		config.ListenAddress(opts.ListenAddress),
		config.ShutdownTimeout(opts.ShutdownTimeout),
		config.MetricStoreInterval(opts.MetricStoreInterval),
		config.MetricStoreFilePath(opts.MetricStoreFilePath),
		config.MetricStoreLoadOnStartup(opts.MetricStoreLoadOnStartup),
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
	app := cli.App[config.Config]{
		Name:          "server",
		ParseArgs:     parseArgs,
		LaunchProcess: launcher.Run,
	}
	return app.Run(ctx, logger.NewProduction(), args, stderr)
}

func main() {
	err := run(context.Background(), os.Args[1:], os.Stderr)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
