package agent

import (
	"context"

	"github.com/bq2cd/yp-go-metrics/internal/agent"
	"github.com/bq2cd/yp-go-metrics/internal/agent/source"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/go-resty/resty/v2"
)

// Run is an app entry point to launch an agent process.
func Run(ctx context.Context, logger log.Logger, cfg config.Config) error {
	logger.Info().
		Dur("poll_interval", cfg.PollInterval).
		Dur("report_interval", cfg.ReportInterval).
		Str("upstream", cfg.UpstreamURL.String()).
		Msg("will send metrics at configured intervals")

	collector := agent.NewCollector(source.DefaultSources(), repository.NewMemStorage())

	client := resty.New().SetBaseURL(cfg.UpstreamURL.String()).SetTimeout(cfg.ReportInterval)
	sender := agent.NewSenderJSON(ctx, client)
	reporter := agent.NewReporter(sender, repository.NewMemStorage())

	ag := agent.New(ctx, cfg, collector, reporter)

	return ag.Run()
}
