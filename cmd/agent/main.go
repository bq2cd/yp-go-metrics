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

	"github.com/bq2cd/yp-go-metrics/internal/agent"
	"github.com/bq2cd/yp-go-metrics/internal/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/go-resty/resty/v2"
)

const (
	defaultUpstreamURL       = "localhost:8080"
	defaultPollIntervalSec   = 2
	defaultReportIntervalSec = 10
)

type cliOptions struct {
	UpstreamURL    string `env:"ADDRESS"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
}

func runAgent(ctx context.Context, cfg config.Config) error {
	log.Printf("sending metrics to %s every %v (poll interval %v)", cfg.UpstreamURL.String(), cfg.ReportInterval, cfg.PollInterval)

	collector := agent.NewCollector(source.DefaultSources(), repository.NewMemStorage())

	client := resty.New().SetBaseURL(cfg.UpstreamURL.String()).SetTimeout(cfg.ReportInterval)
	reporter := agent.NewReporter(ctx, client, repository.NewMemStorage())

	ag := agent.NewAgent(ctx, cfg, collector, reporter)

	return ag.Run()
}

func parseArgs(fs *flag.FlagSet, args []string, envParser envparser.Parser) (config.Config, error) {
	var opts cliOptions

	fs.StringVar(&opts.UpstreamURL, "a", defaultUpstreamURL, "upstream url in the format [http://]HOST[:PORT]")
	fs.UintVar(&opts.PollInterval, "p", defaultPollIntervalSec, "poll interval in seconds")
	fs.UintVar(&opts.ReportInterval, "r", defaultReportIntervalSec, "report interval in seconds")

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
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.SetOutput(stderr)

	cfg, err := parseArgs(fs, args, envparser.NewParser())
	if err != nil {
		return fmt.Errorf("unable to parse args: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = runAgent(ctx, cfg)
	<-ctx.Done()

	return err
}

func main() {
	err := run(context.Background(), os.Args[1:], os.Stderr)
	if err != nil {
		log.Fatalf("failed to start agent: %v", err)
	}
}
