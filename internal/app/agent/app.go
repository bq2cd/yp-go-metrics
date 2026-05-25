package agent

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbmetrics "github.com/bq2cd/yp-go-metrics/api/gen/metrics/v1"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
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

	hmacSigner := hmacsigner.NewHMACSigner(cfg.HMACSecretKey)

	realIP, err := prepareRealIPHeader(cfg.UpstreamURL.Host, hmacSigner)
	if err != nil {
		return err
	}

	sender, err := initSender(logger, cfg, hmacSigner, encryptor, realIP)
	if err != nil {
		return err
	}

	collector := NewCollector(source.DefaultSources(), repository.NewMemStorage())
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

func initSender(logger log.Logger, cfg config.Config, hmacSigner hmacsigner.HMACSigner, encryptor asymcrypt.Encryptor, realIP httpheaders.XRealIP) (SenderBatch, error) {
	switch cfg.UpstreamURL.Scheme {
	case "grpc":
		conn, err := grpc.NewClient(cfg.UpstreamURL.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("cannot create new GRPC channel for %s: %w", cfg.UpstreamURL.Host, err)
		}

		return NewSenderGRPC(logger, pbmetrics.NewMetricsClient(conn), realIP), nil
	default:
		client := resty.New().SetBaseURL(cfg.UpstreamURL.String()).SetTimeout(cfg.ReportInterval)
		retrierFactory := retrymgr.NewRetrierFactory(logger, retrymgr.NewSleeper(), retrymgr.NewStrategy1s3s5s)

		return NewSenderJSON(client, retrierFactory, hmacSigner, encryptor, realIP), nil
	}
}
