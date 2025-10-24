package server

import (
	"context"
	"fmt"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/handler/router"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/server"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Run is an app entry point to launch a server process.
func Run(ctx context.Context, logger log.Logger, cfg config.Config) error {
	dbCfg, err := dbconfig.New(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("cannot create DB config: %w", err)
	}

	if err := applyMigrations(ctx, logger, dbCfg); err != nil {
		return fmt.Errorf("cannot apply DB migrations: %w", err)
	}

	sqlStorage, err := repository.NewSQLStorage(dbCfg)
	if err != nil {
		return fmt.Errorf("cannot create SQL storage: %w", err)
	}

	memStorage := repository.NewMemStorage()
	storer := service.NewMetricStorer(memStorage)
	snapshotter := service.NewMetricSnapshotter(storer, service.NewMetricJSONEncoder(), service.NewMetricJSONDecoder())

	handlers := handler.NewRegistry(logger, snapshotter, sqlStorage)

	router, err := router.New(logger, handlers)
	if err != nil {
		return fmt.Errorf("cannot create router: %w", err)
	}

	srv := server.New(ctx, logger, cfg, router, snapshotter)

	return srv.Run()
}
