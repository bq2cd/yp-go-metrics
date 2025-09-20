package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/agent"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/go-resty/resty/v2"
)

const (
	defaultUpstreamURL       = "http://localhost:8080"
	defaultPollIntervalSec   = 2
	defaultReportIntervalSec = 10
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

func run(args []string) error {
	_ = args

	upstreamURL, err := url.Parse(defaultUpstreamURL)
	if err != nil {
		log.Fatalf("failed to parse upstream url: %v", err)
	}

	cfg := config.Config{
		UpstreamURL:    *upstreamURL,
		PollInterval:   defaultPollIntervalSec * time.Second,
		ReportInterval: defaultReportIntervalSec * time.Second,
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
