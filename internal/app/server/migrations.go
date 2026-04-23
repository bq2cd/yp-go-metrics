package server

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	"github.com/bq2cd/yp-go-metrics/migrations"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
)

type gooseLogger struct {
	logger log.Logger
}

func (l *gooseLogger) Fatalf(format string, v ...any) {
	l.logger.Error().Msg(fmt.Sprintf(format, v...))
}
func (l *gooseLogger) Printf(format string, v ...any) {
	l.logger.Info().Msg(fmt.Sprintf(format, v...))
}

func applyMigrationsWithRetries(ctx context.Context, logger log.Logger, cfg dbconfig.Config, retrierFactory retrymgr.RetrierFactory) error {
	_, err := retrymgr.NewRetrier[any](retrierFactory).Do(
		ctx, "apply_migrations",
		func(ctx context.Context) (any, error) {
			err := applyMigrations(ctx, logger, cfg)
			return nil, err
		},
		func(err error) bool {
			return true
		},
	)
	return err
}

func applyMigrations(ctx context.Context, logger log.Logger, cfg dbconfig.Config) error {
	if !cfg.Enabled() {
		return nil
	}
	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(&gooseLogger{logger: logger.With(log.Str("subsystem", "migrations"))})

	db, err := goose.OpenDBWithDriver(string(cfg.Driver()), cfg.DSN())
	if err != nil {
		return fmt.Errorf("cannot open DB: %w", err)
	}

	logger.Info().Str("dsn", cfg.DSN()).Msg("applying migrations")

	err = goose.UpContext(ctx, db, ".")
	if err != nil {
		return fmt.Errorf("cannot apply migrations: %w", err)
	}
	return nil
}
