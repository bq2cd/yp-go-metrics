package service

import (
	"errors"
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

var (
	// ErrMetricNotFound wraps [repository.ErrMetricNotFound] to avoid exposure of [repository]
	// to a caller.
	ErrMetricNotFound = errors.New("metric not found")
)

// MetricStorer defines an interface to work with metrics.
// E.g. store, retrieve, delete, etc.
type MetricStorer interface {
	StoreSingle(metric model.Metric) error
	StoreBatch(metrics []model.Metric) error
	RetrieveSingle(key model.MetricKey) (model.Metric, error)
	RetrieveBatch(keys []model.MetricKey) ([]model.Metric, error)
	RetrieveAll() ([]model.Metric, error)
}

type metricStorer struct {
	storage repository.Storage
}

// NewMetricStorer creates an instance of the metrics service.
func NewMetricStorer(storage repository.Storage) *metricStorer {
	return &metricStorer{storage: storage}
}

// StoreSingle updates or replaces a single metric.
func (s *metricStorer) StoreSingle(metric model.Metric) error {
	if metric.Empty() {
		return nil
	}
	switch metric.Type {
	case model.MetricTypeCounter:
		prev, err := s.storage.Get(metric.Key())
		if err == repository.ErrMetricNotFound {
			return s.storage.Set(metric)
		}
		if err != nil {
			return fmt.Errorf("failed to retrieve existing metric: %w", err)
		}
		metric = metric.Copy()
		*metric.Delta += *prev.Delta
	default:
		// pass
	}
	return s.storage.Set(metric)
}

// StoreBatch updates or replaces a slice of metrics.
// This method essentially calls StoreSingle method for each metric.
func (s *metricStorer) StoreBatch(metrics []model.Metric) error {
	var errFinal error

	for i := range metrics {
		err := s.StoreSingle(metrics[i])
		errFinal = errors.Join(errFinal, err)
	}

	return errFinal
}

// RetrieveSingle obtains a metric from the underlying storage by given key
// or returns error if metric is not found or storage has failed.
func (s *metricStorer) RetrieveSingle(key model.MetricKey) (model.Metric, error) {
	m, err := s.storage.Get(key)
	if err == nil && m.Empty() {
		// avoid returning empty metrics from underlying storage as
		// the storage should not store such metrics in the first place,
		// but in case it did, we will intercept such cases here.
		return model.Metric{}, ErrMetricNotFound
	}
	if errors.Is(err, repository.ErrMetricNotFound) {
		// wrap not found error to avoid exposing repository layer to the caller.
		return m, ErrMetricNotFound
	}
	return m, err
}

// RetrieveBatch obtains a slice of metrics from the underlying storage by given keys
// or returns an error if storage has failed.
// This method essentially wraps RetrieveSingle while skipping non-existent metrics.
// NB. Number of metrics returned might be smaller than the number of the keys requested.
func (s *metricStorer) RetrieveBatch(keys []model.MetricKey) ([]model.Metric, error) {
	var errFinal error

	metrics := make([]model.Metric, 0, len(keys))

	for _, k := range keys {
		m, err := s.storage.Get(k)
		switch err {
		case nil:
			metrics = append(metrics, m)
		case repository.ErrMetricNotFound:
			continue
		default:
			errFinal = errors.Join(errFinal, err)
		}
	}

	return metrics, errFinal
}

// RetrieveAll returns all metrics from the underlying storage.
func (s *metricStorer) RetrieveAll() ([]model.Metric, error) {
	return s.storage.GetAll()
}
