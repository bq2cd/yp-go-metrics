// Binary server launches an HTTP server that accepts incoming metrics and stores them in configured database.
// The server also provides endpoints to access stored metrics.
package main

import (
	"context"
	"flag"
	"fmt"
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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

type cliOptions struct {
	ListenAddress            string `env:"ADDRESS"`
	ShutdownTimeout          uint   `env:"SHUTDOWN_TIMEOUT"`
	MetricStoreInterval      uint   `env:"STORE_INTERVAL"`
	MetricStoreFilePath      string `env:"FILE_STORAGE_PATH"`
	MetricStoreLoadOnStartup bool   `env:"RESTORE"`
	DatabaseDSN              string `env:"DATABASE_DSN"`
	HMACSecretkey            string `env:"KEY"`
	AuditFilePath            string `env:"AUDIT_FILE"`
	AuditURL                 string `env:"AUDIT_URL"`
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (config.Config, error) {
	var opts cliOptions

	fs.StringVar(&opts.ListenAddress, "a", defaultAddress, "listen address in the format [HOST]:PORT")
	fs.UintVar(&opts.ShutdownTimeout, "t", defaultShutdownTimeoutSec, "graceful shutdown timeout in seconds")
	fs.UintVar(&opts.MetricStoreInterval, "i", defaultMetricStoreIntervalSec, "dump metrics on disk each interval (in seconds)")
	fs.StringVar(&opts.MetricStoreFilePath, "f", defaultMetricStoreFilePath, "path to file for dumping metrics")
	fs.BoolVar(&opts.MetricStoreLoadOnStartup, "r", defaultMetricStoreLoadOnStartup, "restore metrics from file on startup")
	fs.StringVar(&opts.DatabaseDSN, "d", "", "database dsn (only postgres is supported)")
	fs.StringVar(&opts.HMACSecretkey, "k", "", "secret key for HMAC calculation")
	fs.StringVar(&opts.AuditFilePath, "audit-file", "", "path to file for writing audit events")
	fs.StringVar(&opts.AuditURL, "audit-url", "", "remote endpoint URL for sending audit events")

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
		config.DatabaseURL(opts.DatabaseDSN),
		config.HMACSecretKey(opts.HMACSecretkey),
		config.AuditFilePath(opts.AuditFilePath),
		config.AuditURL(opts.AuditURL),
	)
	if err != nil {
		return config.Config{}, fmt.Errorf("unable to construct config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return config.Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return *cfg, nil
}

func run(ctx context.Context, args []string, terminalConfig cli.TerminalConfig) error {
	app := cli.App[config.Config]{
		Name:          "server",
		ParseArgs:     parseArgs,
		LaunchProcess: launcher.Run,
		BuildInfo: cli.BuildInfo{
			Version: buildVersion,
			Date:    buildDate,
			Commit:  buildCommit,
		},
		TerminalConfig: terminalConfig,
	}

	return app.Run(ctx, logger.NewProduction(), args)
}

func main() {
	err := run(context.Background(), os.Args[1:], cli.TerminalConfig{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
