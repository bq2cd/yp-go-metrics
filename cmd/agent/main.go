// Binary agent launches a process that collects its runtime metrics
// (e.g. memory consumption, number of goroutines, etc.) and uploads them to a remote server.
package main

import (
	"context"
	"flag"
	"fmt"
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

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

type cliOptions struct {
	ConfigFilePath      string `env:"CONFIG" json:"-"`
	UpstreamURL         string `env:"ADDRESS" json:"address"`
	PollInterval        uint   `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval      uint   `env:"REPORT_INTERVAL" json:"report_interval"`
	HMACSecretKey       string `env:"KEY" json:"key"`
	SenderPoolSize      uint   `env:"RATE_LIMIT" json:"rate_limit"`
	ServerPublicKeyFile string `env:"CRYPTO_KEY" json:"crypto_key"`
}

func defineArgs(fs *flag.FlagSet, opts *cliOptions) {
	fs.StringVar(&opts.ConfigFilePath, "c", "", "path to config file in JSON format (e.g. config.json)")
	fs.StringVar(&opts.UpstreamURL, "a", defaultUpstreamURL, "upstream url in the format [http://]HOST[:PORT]")
	fs.UintVar(&opts.PollInterval, "p", defaultPollIntervalSec, "poll interval in seconds")
	fs.UintVar(&opts.ReportInterval, "r", defaultReportIntervalSec, "report interval in seconds")
	fs.StringVar(&opts.HMACSecretKey, "k", "", "secret key for HMAC calculation")
	fs.UintVar(&opts.SenderPoolSize, "l", defaultSenderPoolSize, "sender pool size (aka rate limit)")
	fs.StringVar(&opts.ServerPublicKeyFile, "crypto-key", "", "path to a file with server's X25519 public key")
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (config.Config, error) {
	parser := cli.Parser[cliOptions, config.Config]{
		DefineArgs:        defineArgs,
		GetConfigFilePath: func(opts *cliOptions) string { return opts.ConfigFilePath },
		CreateConfig:      createConfig,
	}

	return parser.Parse(fs, args, envParser)
}

func createConfig(opts *cliOptions) (config.Config, error) {
	cfg, err := config.New(
		config.UpstreamURL(opts.UpstreamURL),
		config.PollInterval(opts.PollInterval),
		config.ReportInterval(opts.ReportInterval),
		config.HMACSecretKey(opts.HMACSecretKey),
		config.ServerPublicKey(opts.ServerPublicKeyFile),
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

func run(ctx context.Context, args []string, terminalConfig cli.TerminalConfig) error {
	app := cli.App[config.Config]{
		Name:          "agent",
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
		log.Fatalf("failed to start agent: %v", err)
	}
}
