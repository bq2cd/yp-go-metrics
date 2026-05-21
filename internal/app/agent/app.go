package agent

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/repository"

	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
)

// Run is an app entry point to launch an agent process.
func Run(ctx context.Context, logger log.Logger, cfg config.Config) error {
	encryptor, err := initEncryptor(cfg)
	if err != nil {
		return err
	}

	collector := NewCollector(source.DefaultSources(), repository.NewMemStorage())

	client := resty.New().SetBaseURL(cfg.UpstreamURL.String()).SetTimeout(cfg.ReportInterval)
	retrierFactory := retrymgr.NewRetrierFactory(logger, retrymgr.NewSleeper(), retrymgr.NewStrategy1s3s5s)
	hmacSigner := hmacsigner.NewHMACSigner(cfg.HMACSecretKey)
	sender := NewSenderJSON(client, retrierFactory, hmacSigner, encryptor)
	reporter := NewReporter(sender, repository.NewMemStorage(), cfg.SenderPoolSize)

	logger.Info().
		Dur("poll_interval", cfg.PollInterval).
		Dur("report_interval", cfg.ReportInterval).
		Str("upstream", cfg.UpstreamURL.String()).
		Msg("will send metrics at configured intervals")

	ag := New(logger, cfg, collector, reporter)

	return ag.Run(ctx)
}

func initEncryptor(cfg config.Config) (asymcrypt.Encryptor, error) {
	if len(cfg.ServerPublicKey) == 0 {
		return nil, nil
	}

	pubkey, err := asymcrypt.ParsePublicKey(cfg.ServerPublicKey)
	if err != nil {
		return nil, fmt.Errorf("cannot parse public key: %w", err)
	}

	encryptor, err := asymcrypt.NewEncryptor(pubkey)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryptor: %w", err)
	}

	return encryptor, nil
}
