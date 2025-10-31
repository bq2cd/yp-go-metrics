package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	tableNamePrefix = "metrics_"
)

var (
	ErrUnsupportedMetricType = errors.New("unsupported metric type")
	ErrMetricNotFound        = repository.ErrMetricNotFound
)

type sqlSelector interface {
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

type sqlExecer interface {
	sqlx.ExecerContext
}

type sqlStorage struct {
	db             *sqlx.DB
	retrierFactory retrymgr.RetrierFactory
}

// New creates an instance of the storage backed by an SQL database.
// Currently, only PostgreSQL is supported.
func New(cfg dbconfig.Config, retrierFactory retrymgr.RetrierFactory) (*sqlStorage, error) {
	db, err := sqlx.Open(string(cfg.Driver()), cfg.DSN())
	if err != nil {
		return nil, err
	}
	return &sqlStorage{db: db, retrierFactory: retrierFactory}, nil
}

// Ping returns an error if the underlying database is not reachable.
func (s *sqlStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close calls [sql.DB.Close] method under the hood.
func (s *sqlStorage) Close() error {
	return s.db.Close()
}

// Get returns a [model.Metric] by its key or [ErrMetricNotFound] is metric is not found.
func (s *sqlStorage) Get(ctx context.Context, key model.MetricKey) (model.Metric, error) {
	metrics, err := s.GetMulti(ctx, model.NewMetricKeySet(key))
	if err != nil {
		return model.Metric{}, err
	}
	if len(metrics) == 0 {
		return model.Metric{}, ErrMetricNotFound
	}
	return metrics[0], nil
}

// GetMulti returns multiple metrics by the given keys.
func (s *sqlStorage) GetMulti(ctx context.Context, keys model.MetricKeySet) ([]model.Metric, error) {
	idsByType := make(map[model.MetricType][]string, len(keys))
	for key := range keys {
		idsByType[key.Type] = append(idsByType[key.Type], key.ID)
	}

	return s.getMultiByTypeWithRetries(ctx, idsByType)
}

// GetAll returns all metrics in the database.
func (s *sqlStorage) GetAll(ctx context.Context) ([]model.Metric, error) {
	idsByType := make(map[model.MetricType][]string)
	for t := range sqlHandlersByMetricType {
		idsByType[t] = nil // get all metrics
	}

	return s.getMultiByTypeWithRetries(ctx, idsByType)
}

func (s *sqlStorage) getMultiByTypeWithRetries(ctx context.Context, idsByType map[model.MetricType][]string) ([]model.Metric, error) {
	return retrymgr.NewRetrier[[]model.Metric](s.retrierFactory).Do(
		ctx, "sql_storage_get_multi",
		func(ctx context.Context) ([]model.Metric, error) {
			return s.getMultiByType(ctx, idsByType)
		},
		func(err error) bool {
			return false
		},
	)
}

func (s *sqlStorage) getMultiByType(ctx context.Context, idsByType map[model.MetricType][]string) ([]model.Metric, error) {
	total := 0
	for _, ids := range idsByType {
		total += len(ids)
	}
	metrics := make([]model.Metric, 0, total)
	// FIXME: temporary hack to work around `sqlmock` limitations
	// when it is expecting the queries to be called in the same
	// order expectations are arranged.
	for _, mtype := range slices.Sorted(maps.Keys(idsByType)) {
		ids := idsByType[mtype]
		selected, err := s.getMultiForType(ctx, mtype, ids)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, selected...)
	}
	return metrics, nil
}

func (s *sqlStorage) getMultiForType(ctx context.Context, metricType model.MetricType, itemIds []string) ([]model.Metric, error) {
	metrics := make([]model.Metric, 0, len(itemIds))

	mgr, ok := sqlHandlersByMetricType[metricType]
	if !ok {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedMetricType, metricType)
	}

	items, err := mgr.Select(ctx, s.db, itemIds...)
	if err != nil {
		return nil, fmt.Errorf("cannot select metrics of type %v: %w", metricType, err)
	}
	for _, it := range items {
		metrics = append(metrics, it.ToMetric())
	}

	return metrics, nil
}

// Set stores given metric in the database.
func (s *sqlStorage) Set(ctx context.Context, metric model.Metric) error {
	return s.SetMulti(ctx, model.NewMetricSet(metric))
}

// SetMulti stores given metrics in the database.
func (s *sqlStorage) SetMulti(ctx context.Context, metrics model.MetricSet) error {
	return s.setMultiWithRetries(ctx, metrics)
}

func (s *sqlStorage) setMultiWithRetries(ctx context.Context, metrics model.MetricSet) error {
	_, err := retrymgr.NewRetrier[any](s.retrierFactory).Do(
		ctx, "sql_storage_set_multi",
		func(ctx context.Context) (any, error) {
			err := s.setMulti(ctx, metrics)
			return nil, err
		},
		func(err error) bool {
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				return false
			}
			return pgerrcode.IsConnectionException(pgErr.Code)
		},
	)
	return err
}

func (s *sqlStorage) setMulti(ctx context.Context, metrics model.MetricSet) error {
	// FIXME: temporary hack to work around `sqlmock` limitations
	// when it is expecting the queries to be called in the same
	// order expectations are arranged.
	mgroupByType := metrics.GroupByType()
	for _, mtype := range slices.Sorted(maps.Keys(mgroupByType)) {
		mgroup := mgroupByType[mtype]
		err := s.setMultiForType(ctx, mtype, mgroup)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sqlStorage) setMultiForType(ctx context.Context, metricType model.MetricType, metrics model.MetricSet) error {
	h, ok := sqlHandlersByMetricType[metricType]
	if !ok {
		return fmt.Errorf("%w: %v", ErrUnsupportedMetricType, metricType)
	}
	items := h.ConvertMetrics(metrics)
	if len(items) == 0 {
		return nil
	}

	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("cannot begin transaction for type %v: %w", metricType, err)
	}
	defer tx.Rollback()

	_, err = h.Insert(ctx, tx, items...)
	if err != nil {
		return fmt.Errorf("cannot insert metrics of type %v: %w", metricType, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot commit transaction for type %v: %w", metricType, err)
	}
	return nil
}
