package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	launcher "github.com/bq2cd/yp-go-metrics/internal/app/agent"
	"github.com/bq2cd/yp-go-metrics/internal/app/cli"
	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/internal/app/logger"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
)

const (
	defaultUpstreamURL       = "localhost:8080"
	defaultPollIntervalSec   = 2
	defaultReportIntervalSec = 10
	defaultSenderPoolSize    = 0 // sender pool is disabled, sending done serially
)

type cliOptions struct {
	UpstreamURL    string `env:"ADDRESS"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	HMACSecretKey  string `env:"KEY"`
	SenderPoolSize uint   `env:"RATE_LIMIT"`
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (config.Config, error) {
	var opts cliOptions

	fs.StringVar(&opts.UpstreamURL, "a", defaultUpstreamURL, "upstream url in the format [http://]HOST[:PORT]")
	fs.UintVar(&opts.PollInterval, "p", defaultPollIntervalSec, "poll interval in seconds")
	fs.UintVar(&opts.ReportInterval, "r", defaultReportIntervalSec, "report interval in seconds")
	fs.StringVar(&opts.HMACSecretKey, "k", "", "secret key for HMAC calculation")
	fs.UintVar(&opts.SenderPoolSize, "l", defaultSenderPoolSize, "sender pool size (aka rate limit)")

	// parse flags
	if err := fs.Parse(args); err != nil {
		return config.Config{}, fmt.Errorf("invalid args: %w", err)
	}

	// parse env vars (take precedence over flags)
	if err := envParser.Parse(&opts); err != nil {
		return config.Config{}, fmt.Errorf("invalid env vars: %w", err)
	}

	cfg, err := config.New(
		config.UpstreamURL(opts.UpstreamURL),
		config.PollInterval(opts.PollInterval),
		config.ReportInterval(opts.ReportInterval),
		config.HMACSecretKey(opts.HMACSecretKey),
		config.SenderPoolSize(opts.SenderPoolSize),
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
		Name:          "agent",
		ParseArgs:     parseArgs,
		LaunchProcess: launcher.Run,
	}
	return app.Run(ctx, logger.NewProduction(), args, stderr)
}

func main() {
	err := run(context.Background(), os.Args[1:], os.Stderr)
	if err != nil {
		log.Fatalf("failed to start agent: %v", err)
	}
}
