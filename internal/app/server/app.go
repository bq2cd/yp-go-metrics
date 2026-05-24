package server

import (
	"context"
	"fmt"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/handler/router"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/repository/auditsink"
	"github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
)

// Run is an app entry point to launch a server process.
func Run(ctx context.Context, logger log.Logger, cfg config.Config) error {
	storage, pinger, err := initStorage(ctx, logger, cfg)
	if err != nil {
		return fmt.Errorf("cannot init storage: %w", err)
	}

	auditProcessor, err := initAuditProcessor(ctx, logger, cfg)
	if err != nil {
		return fmt.Errorf("cannot init metric auditor: %w", err)
	}

	decryptor, err := initDecryptor(cfg)
	if err != nil {
		return fmt.Errorf("cannot init decryptor: %w", err)
	}

	writer := service.NewStorageBatchWriter(storage)
	storer := service.NewMetricStorer(storage, writer)
	snapshotter := service.NewMetricSnapshotter(storer, service.NewMetricJSONEncoder(), service.NewMetricJSONDecoder())
	auditor := service.NewMetricAuditor(logger, auditProcessor)

	handlers := handler.NewRegistry(logger, snapshotter, pinger, auditor)
	hmacSigner := hmacsigner.NewHMACSigner(cfg.HMACSecretKey)

	router, err := router.New(logger, handlers, hmacSigner, decryptor, cfg.TrustedSubnet)
	if err != nil {
		return fmt.Errorf("cannot create router: %w", err)
	}

	srv := New(logger, cfg, router, snapshotter, writer, auditProcessor)

	return srv.Run(ctx)
}

func initStorage(ctx context.Context, logger log.Logger, cfg config.Config) (repository.StorageMulti, service.StoragePinger, error) {
	dbCfg, err := dbconfig.New(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create DB config: %w", err)
	}

	retrierFactory := retrymgr.NewRetrierFactory(logger, retrymgr.NewSleeper(), retrymgr.NewStrategy1s3s5s)

	err = applyMigrationsWithRetries(ctx, logger, dbCfg, retrierFactory)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot apply DB migrations: %w", err)
	}

	sqlStorage, err := sqlstorage.New(dbCfg, retrierFactory)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create SQL storage: %w", err)
	}

	if !dbCfg.Enabled() {
		return repository.NewMemStorage(), sqlStorage, nil
	}

	pingCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	if err := sqlStorage.Ping(pingCtx); err != nil {
		return nil, nil, fmt.Errorf("cannot ping SQL storage: %w", err)
	}

	return sqlStorage, sqlStorage, nil
}

func initAuditProcessor(_ context.Context, logger log.Logger, cfg config.Config) (service.AuditEventProcessor, error) {
	processor := service.NewAuditEventProcessor(logger)

	err := maybeRegisterAuditFileSink(logger, processor, cfg.AuditFilePath)
	if err != nil {
		return nil, err
	}

	err = maybeRegisterAuditHTTPSink(logger, processor, cfg.AuditURL)
	if err != nil {
		return nil, err
	}

	return processor, nil
}

func initDecryptor(cfg config.Config) (asymcrypt.Decryptor, error) {
	if len(cfg.DecryptionPrivateKey) == 0 {
		return nil, nil
	}

	key, err := asymcrypt.ParsePrivateKey(cfg.DecryptionPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key: %w", err)
	}

	decryptor := asymcrypt.NewDecryptor(key)

	return decryptor, nil
}

func maybeRegisterAuditFileSink(logger log.Logger, processor service.AuditEventProcessor, path string) error {
	if path == "" {
		return nil
	}

	sink, err := auditsink.NewFileSink(path)
	if err != nil {
		return fmt.Errorf("cannot create audit file sink: %w", err)
	}

	processor.RegisterSink("file:"+path, sink)

	logger.Info().Str("path", path).Msg("registered audit file sink")

	return nil
}

func maybeRegisterAuditHTTPSink(logger log.Logger, processor service.AuditEventProcessor, remote url.URL) error {
	if remote.String() == "" {
		return nil
	}

	sink, err := auditsink.NewHTTPSink(remote)
	if err != nil {
		return fmt.Errorf("cannot create audit http sink: %w", err)
	}

	processor.RegisterSink("http:"+remote.String(), sink)

	logger.Info().Str("remote", remote.String()).Msg("registered audit http sink")

	return nil
}
