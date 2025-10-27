package agent

import (
	"context"

	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/go-resty/resty/v2"
)

// Run is an app entry point to launch an agent process.
func Run(ctx context.Context, logger log.Logger, cfg config.Config) error {
	logger.Info().
		Dur("poll_interval", cfg.PollInterval).
		Dur("report_interval", cfg.ReportInterval).
		Str("upstream", cfg.UpstreamURL.String()).
		Msg("will send metrics at configured intervals")

	collector := NewCollector(source.DefaultSources(), repository.NewMemStorage())

	client := resty.New().SetBaseURL(cfg.UpstreamURL.String()).SetTimeout(cfg.ReportInterval)
	sender := NewSenderJSON(client)
	reporter := NewReporter(sender, repository.NewMemStorage())

	ag := New(cfg, collector, reporter)

	return ag.Run(ctx)
}
