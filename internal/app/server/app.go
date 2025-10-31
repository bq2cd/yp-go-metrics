package server

import (
	"context"
	"fmt"
	"time"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/handler/router"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Run is an app entry point to launch a server process.
func Run(ctx context.Context, logger log.Logger, cfg config.Config) error {
	storage, pinger, err := initStorage(ctx, logger, cfg)
	if err != nil {
		return fmt.Errorf("cannot init storage: %w", err)
	}

	writer := service.NewStorageBatchWriter(storage)
	storer := service.NewMetricStorer(storage, writer)
	snapshotter := service.NewMetricSnapshotter(storer, service.NewMetricJSONEncoder(), service.NewMetricJSONDecoder())

	handlers := handler.NewRegistry(logger, snapshotter, pinger)
	router, err := router.New(logger, handlers)
	if err != nil {
		return fmt.Errorf("cannot create router: %w", err)
	}

	srv := New(logger, cfg, router, snapshotter, writer)

	return srv.Run(ctx)
}

func initStorage(ctx context.Context, logger log.Logger, cfg config.Config) (repository.StorageMulti, service.StoragePinger, error) {
	dbCfg, err := dbconfig.New(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create DB config: %w", err)
	}

	retrierFactory := retrymgr.NewRetrierFactory(logger, retrymgr.NewSleeper(), retrymgr.NewStrategy1s3s5s)

	if err := applyMigrationsWithRetries(ctx, logger, dbCfg, retrierFactory); err != nil {
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
