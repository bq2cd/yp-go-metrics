package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/agent"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/go-resty/resty/v2"
)

const (
	defaultUpstreamURL       = "localhost:8080"
	defaultPollIntervalSec   = 2
	defaultReportIntervalSec = 10
)

var (
	reURLCheck = regexp.MustCompile("^[^:/]+://.+")
)

func runAgent(ctx context.Context, cfg config.Config) error {
	if cfg.PollInterval >= cfg.ReportInterval {
		return fmt.Errorf("poll interval must be less than report interval (got %v >= %v)", cfg.PollInterval, cfg.ReportInterval)
	}

	log.Printf("sending metrics to %s every %v (poll interval %v)", cfg.UpstreamURL.String(), cfg.ReportInterval, cfg.PollInterval)

	collector := agent.NewDefaultCollector()

	client := resty.New().SetBaseURL(cfg.UpstreamURL.String()).SetTimeout(cfg.ReportInterval)
	reporter := agent.NewDefaultReporter(ctx, client)

	ag := agent.NewAgent(ctx, cfg, collector, reporter)

	return ag.Run()
}

func parseArgs(args []string) (config.Config, error) {
	var (
		upstreamURL    string
		pollInterval   uint
		reportInterval uint
	)

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	fs.StringVar(&upstreamURL, "a", defaultUpstreamURL, "upstream url in the format [http://]HOST[:PORT]")
	fs.UintVar(&pollInterval, "p", defaultPollIntervalSec, "poll interval in seconds")
	fs.UintVar(&reportInterval, "r", defaultReportIntervalSec, "report interval in seconds")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, fmt.Errorf("invalid args: %w", err)
	}

	cfg := config.Config{}

	// validate upstream url
	{
		if !reURLCheck.MatchString(upstreamURL) {
			upstreamURL = "http://" + upstreamURL
		}

		parsed, err := url.Parse(upstreamURL)
		if err != nil {
			return config.Config{}, fmt.Errorf("invalid upstream url (%v)", upstreamURL)
		}

		cfg.UpstreamURL = *parsed
	}

	// validate poll interval
	{
		if pollInterval == 0 {
			return config.Config{}, fmt.Errorf("poll interval must be non-zero")
		}

		cfg.PollInterval = time.Duration(pollInterval) * time.Second
	}

	// validate report interval
	{
		if reportInterval == 0 {
			return config.Config{}, fmt.Errorf("report interval must be non-zero")
		}
		if reportInterval < pollInterval {
			return config.Config{}, fmt.Errorf("report interval must be greater than poll interval")
		}

		cfg.ReportInterval = time.Duration(reportInterval) * time.Second
	}

	return cfg, nil
}

func run(args []string) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return fmt.Errorf("unable to parse args: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = runAgent(ctx, cfg)
	<-ctx.Done()

	return err
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to start agent: %v", err)
	}
}
