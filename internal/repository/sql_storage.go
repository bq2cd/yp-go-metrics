package repository

import (
	"context"
	"database/sql"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type sqlStorage struct {
	db *sql.DB
}

// NewSQLStorage creates an instance of the storage backed by an SQL database.
// Currently, only PostgreSQL is supported.
func NewSQLStorage(cfg dbconfig.Config) (*sqlStorage, error) {
	db, err := sql.Open(string(cfg.Driver()), cfg.DSN())
	if err != nil {
		return nil, err
	}
	return &sqlStorage{db: db}, nil
}

// Ping returns an error if the underlying database is not reachable.
func (s *sqlStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close calls [sql.DB.Close] method under the hood.
func (s *sqlStorage) Close() error {
	return s.db.Close()
}
