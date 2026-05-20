// Binary server launches an HTTP server that accepts incoming metrics and stores them in configured database.
// The server also provides endpoints to access stored metrics.
package main

import (
	"context"
	"encoding/json"
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
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

type cliOptions struct {
	ConfigFile               string `env:"CONFIG" json:"-"`
	ListenAddress            string `env:"ADDRESS" json:"address"`
	ShutdownTimeout          uint   `env:"SHUTDOWN_TIMEOUT" json:"shutdown_timeout"`
	MetricStoreInterval      uint   `env:"STORE_INTERVAL" json:"store_interval"`
	MetricStoreFilePath      string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	MetricStoreLoadOnStartup bool   `env:"RESTORE" json:"restore"`
	DatabaseDSN              string `env:"DATABASE_DSN" json:"database_dsn"`
	HMACSecretkey            string `env:"KEY" json:"key"`
	DecryptionPrivateKeyFile string `env:"CRYPTO_KEY" json:"crypto_key"`
	AuditFilePath            string `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL                 string `env:"AUDIT_URL" json:"audit_url"`
}

func defineArgs(fs *flag.FlagSet, opts *cliOptions) {
	fs.StringVar(&opts.ConfigFile, "c", "", "path to config file in JSON format (e.g. config.json)")
	fs.StringVar(&opts.ListenAddress, "a", defaultAddress, "listen address in the format [HOST]:PORT")
	fs.UintVar(&opts.ShutdownTimeout, "t", defaultShutdownTimeoutSec, "graceful shutdown timeout in seconds")
	fs.UintVar(&opts.MetricStoreInterval, "i", defaultMetricStoreIntervalSec, "dump metrics on disk each interval (in seconds)")
	fs.StringVar(&opts.MetricStoreFilePath, "f", defaultMetricStoreFilePath, "path to file for dumping metrics")
	fs.BoolVar(&opts.MetricStoreLoadOnStartup, "r", defaultMetricStoreLoadOnStartup, "restore metrics from file on startup")
	fs.StringVar(&opts.DatabaseDSN, "d", "", "database dsn (only postgres is supported)")
	fs.StringVar(&opts.HMACSecretkey, "k", "", "secret key for HMAC calculation")
	fs.StringVar(&opts.AuditFilePath, "audit-file", "", "path to file for writing audit events")
	fs.StringVar(&opts.AuditURL, "audit-url", "", "remote endpoint URL for sending audit events")
	fs.StringVar(&opts.DecryptionPrivateKeyFile, "crypto-key", "", "path to a file with server's X25519 private key (for decryption)")
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (config.Config, error) {
	var (
		cfg  config.Config
		opts cliOptions
		err  error
	)

	defineArgs(fs, &opts)

	err = populateOptions(fs, args, envParser, &opts)
	if err != nil {
		return cfg, err
	}

	err = loadConfigFile(&opts)
	if err != nil {
		return cfg, err
	}

	// Override options from config file with options from CLI flags and environment variables.
	// Precedence order: config file -> CLI flags -> environment variables
	err = populateOptions(fs, args, envParser, &opts)
	if err != nil {
		return cfg, err
	}

	cfg, err = createConfig(&opts)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func populateOptions(fs *flag.FlagSet, args []string, envParser envparser.Parser, opts *cliOptions) error {
	// parse flags
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("invalid args: %w", err)
	}

	// parse env vars (take precedence over flags)
	if err := envParser.Parse(opts); err != nil {
		return fmt.Errorf("invalid env vars: %w", err)
	}

	return nil
}

func loadConfigFile(opts *cliOptions) error {
	if opts.ConfigFile == "" {
		return nil
	}

	cfg, err := os.Open(opts.ConfigFile)
	if err != nil {
		return fmt.Errorf("cannot open config file at %s: %w", opts.ConfigFile, err)
	}
	defer cfg.Close()

	err = json.NewDecoder(cfg).Decode(opts)
	if err != nil {
		return fmt.Errorf("cannot JSON-decode config file: %w", err)
	}

	return nil
}

func createConfig(opts *cliOptions) (config.Config, error) {
	cfg, err := config.New(
		config.ListenAddress(opts.ListenAddress),
		config.ShutdownTimeout(opts.ShutdownTimeout),
		config.MetricStoreInterval(opts.MetricStoreInterval),
		config.MetricStoreFilePath(opts.MetricStoreFilePath),
		config.MetricStoreLoadOnStartup(opts.MetricStoreLoadOnStartup),
		config.DatabaseURL(opts.DatabaseDSN),
		config.HMACSecretKey(opts.HMACSecretkey),
		config.DecryptionPrivateKey(opts.DecryptionPrivateKeyFile),
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
