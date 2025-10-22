package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrUnsupportedDBType = errors.New("unsupported database type")
)

type sqlStorage struct {
	db *sql.DB
}

// NewSQLStorage creates an instance of the storage backed by an SQL database.
// Currently, only PostgreSQL is supported.
func NewSQLStorage(dbURL url.URL) (*sqlStorage, error) {
	dsn := dbURL.String()
	isPostgres := dbURL.Scheme == "postgres" || dbURL.Scheme == "postgresql"
	if dsn != "" && !isPostgres {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDBType, dbURL.Scheme)
	}
	db, err := sql.Open("pgx", dsn)
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
