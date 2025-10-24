package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/huandu/go-sqlbuilder"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	tableNameCounter = "metrics_counter"
	tableNameGauge   = "metrics_gauge"
)

var (
	ErrUnsupportedMetricType = errors.New("unsupported metric type")
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

func (s *sqlStorage) prepareGetBuilder(key model.MetricKey) (sqlbuilder.Builder, error) {
	var tableName string
	switch key.Type {
	case model.MetricTypeCounter:
		tableName = tableNameCounter
	case model.MetricTypeGauge:
		tableName = tableNameGauge
	default:
		return nil, ErrUnsupportedMetricType
	}

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("value").
		From(tableName).
		Where(sb.EQ("metric_id", key.ID))

	return sb, nil
}

// Get returns a [model.Metric] by its key or [ErrMetricNotFound] is metric is not found.
func (s *sqlStorage) Get(ctx context.Context, key model.MetricKey) (model.Metric, error) {
	var (
		metric model.Metric
		err    error
	)
	sb, err := s.prepareGetBuilder(key)
	if err != nil {
		return model.Metric{}, err
	}

	query, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	row := s.db.QueryRowContext(ctx, query, args...)
	switch key.Type {
	case model.MetricTypeCounter:
		var delta int64
		err = row.Scan(&delta)
		metric = model.NewCounterMetric(key.ID, delta)
	case model.MetricTypeGauge:
		var value float64
		err = row.Scan(&value)
		metric = model.NewGaugeMetric(key.ID, value)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return model.Metric{}, ErrMetricNotFound
	}
	if err != nil {
		return model.Metric{}, fmt.Errorf("cannot query metric: %w", err)
	}
	return metric, nil
}

func (s *sqlStorage) prepareGetAllBuilder() sqlbuilder.Builder {
	sbCounter := sqlbuilder.NewSelectBuilder()
	sbCounter.
		Select(sbCounter.As(`'counter'`, "metric_type"), "metric_id", "value::NUMERIC").
		From(`metrics_counter`)

	sbGauge := sqlbuilder.NewSelectBuilder()
	sbGauge.
		Select(sbGauge.As(`'gauge'`, "metric_type"), "metric_id", "value::NUMERIC").
		From(`metrics_gauge`)

	sb := sqlbuilder.NewUnionBuilder()
	sb.UnionAll(sbCounter, sbGauge)

	return sb
}

// GetAll returns all metrics in the database.
func (s *sqlStorage) GetAll(ctx context.Context) ([]model.Metric, error) {
	sb := s.prepareGetAllBuilder()

	query, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.db.QueryContext(ctx, query, args...)

	metrics := make([]model.Metric, 0)
	if errors.Is(err, sql.ErrNoRows) {
		return []model.Metric{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot query database: %w", err)
	}
	for rows.Next() {
		var (
			metricType, metricID string
			value                float64
		)
		err := rows.Scan(&metricType, &metricID, &value)
		if err != nil {
			return nil, fmt.Errorf("cannot scan rows: %w", err)
		}
		switch model.MetricType(metricType) {
		case model.MetricTypeCounter:
			metrics = append(metrics, model.NewCounterMetric(metricID, int64(value)))
		case model.MetricTypeGauge:
			metrics = append(metrics, model.NewGaugeMetric(metricID, value))
		}
	}
	if err := rows.Err(); err != nil {
		return metrics, fmt.Errorf("cannon iterate rows: %w", err)
	}
	if err := rows.Close(); err != nil {
		return metrics, fmt.Errorf("cannot close rows: %w", err)
	}
	return metrics, nil
}

func (s *sqlStorage) prepareSetBuilder(metric model.Metric) (sqlbuilder.Builder, error) {
	var (
		tableName string
		value     any
	)
	switch metric.Type {
	case model.MetricTypeCounter:
		tableName = tableNameCounter
		value = *metric.Delta
	case model.MetricTypeGauge:
		tableName = tableNameGauge
		value = *metric.Value
	default:
		return nil, ErrUnsupportedMetricType
	}

	sb := sqlbuilder.NewInsertBuilder()
	sb.InsertInto(tableName).
		Cols("metric_id", "value").
		Values(metric.ID, value).
		SQL(`ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value`)

	return sb, nil
}

// Set stores given metric in the database.
func (s *sqlStorage) Set(ctx context.Context, metric model.Metric) error {
	if metric.Empty() {
		return nil
	}

	sb, err := s.prepareSetBuilder(metric)
	if err != nil {
		return err
	}

	query, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("cannot update database: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("cannot get affected rows: %w", err)
	}
	if n != 1 {
		return fmt.Errorf("nothing was updated in database")
	}
	return nil
}
