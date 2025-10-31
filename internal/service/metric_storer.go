package service

import (
	"context"
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

var (
	// ErrMetricNotFound wraps [repository.ErrMetricNotFound] to avoid exposure of [repository]
	// to a caller.
	ErrMetricNotFound = errors.New("metric not found")
)

//go:generate go tool mockgen -destination=servicetest/mock_metric_storer.go -package servicetest github.com/bq2cd/yp-go-metrics/internal/service MetricStorer

// MetricStorer defines an interface to work with metrics.
// E.g. store, retrieve, delete, etc.
type MetricStorer interface {
	StoreSingle(ctx context.Context, metric model.Metric) error
	StoreBatch(ctx context.Context, metrics []model.Metric) error
	RetrieveSingle(ctx context.Context, key model.MetricKey) (model.Metric, error)
	RetrieveBatch(ctx context.Context, keys []model.MetricKey) ([]model.Metric, error)
	RetrieveAll(ctx context.Context) ([]model.Metric, error)
}

type metricStorer struct {
	reader repository.StorageMultiReader
	writer StorageBatchWriter
}

// NewMetricStorer creates an instance of the metrics service.
func NewMetricStorer(reader repository.StorageMultiReader, writer StorageBatchWriter) *metricStorer {
	return &metricStorer{reader: reader, writer: writer}
}

// StoreSingle updates or replaces a single metric.
func (s *metricStorer) StoreSingle(ctx context.Context, metric model.Metric) error {
	if metric.Empty() {
		return nil
	}
	return s.StoreBatch(ctx, []model.Metric{metric})
}

// StoreBatch updates or replaces a slice of metrics.
// This method essentially calls StoreSingle method for each metric.
func (s *metricStorer) StoreBatch(ctx context.Context, metrics []model.Metric) error {
	tx := s.writer.WriteBatch(ctx, metrics)
	err := <-tx.Result()
	return err
}

// RetrieveSingle obtains a metric from the underlying storage by given key
// or returns error if metric is not found or storage has failed.
func (s *metricStorer) RetrieveSingle(ctx context.Context, key model.MetricKey) (model.Metric, error) {
	metrics, err := s.RetrieveBatch(ctx, []model.MetricKey{key})
	if err != nil {
		return model.Metric{}, err
	}
	if len(metrics) == 0 {
		return model.Metric{}, ErrMetricNotFound
	}
	return metrics[0], nil
}

// RetrieveBatch obtains a slice of metrics from the underlying storage by given keys
// or returns an error if storage has failed.
// This method essentially wraps RetrieveSingle while skipping non-existent metrics.
// NB. Number of metrics returned might be smaller than the number of the keys requested.
func (s *metricStorer) RetrieveBatch(ctx context.Context, keys []model.MetricKey) ([]model.Metric, error) {
	if len(keys) == 0 {
		return []model.Metric{}, nil
	}
	return s.reader.GetMulti(ctx, model.NewMetricKeySet(keys...))
}

// RetrieveAll returns all metrics from the underlying storage.
func (s *metricStorer) RetrieveAll(ctx context.Context) ([]model.Metric, error) {
	return s.reader.GetAll(ctx)
}
