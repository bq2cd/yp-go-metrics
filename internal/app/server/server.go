package server

import (
	"context"
	"fmt"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/handler/router"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/server"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

// Run is an app entry point to launch a server process.
func Run(ctx context.Context, logger log.Logger, cfg config.Config) error {
	storage := repository.NewMemStorage()
	storer := service.NewMetricStorer(storage)
	snapshotter := service.NewMetricSnapshotter(storer, service.NewMetricJSONEncoder(), service.NewMetricJSONDecoder())
	handlers := handler.NewRegistry(logger, snapshotter)

	router, err := router.New(logger, handlers)
	if err != nil {
		return fmt.Errorf("cannot create router: %w", err)
	}

	srv := server.New(ctx, logger, cfg, router, snapshotter)

	return srv.Run()
}
