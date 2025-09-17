package service

import (
	"errors"
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

// Metrics defines an interface to work with metrics.
// E.g. store, retrieve, delete, etc.
type Metrics interface {
	Store(metric model.Metric, metrics ...model.Metric) error
	Retrieve(hash model.MetricHash, hashes ...model.MetricHash) ([]model.Metric, error)
}

type metricService struct {
	storage repository.Storage
}

// NewMetrics creates an instance of the metrics service.
func NewMetrics(storage repository.Storage) *metricService {
	return &metricService{storage: storage}
}

func (s *metricService) storeSingle(metric model.Metric) error {
	switch metric.Type {
	case model.MetricTypeCounter:
		prev, err := s.storage.Get(metric.Hash)
		if err == repository.ErrMetricNotFound {
			return s.storage.Set(metric)
		}
		if err != nil {
			return fmt.Errorf("failed to retrieve existing metric: %w", err)
		}
		*metric.Delta += *prev.Delta
	default:
		// pass
	}
	return s.storage.Set(metric)
}

// Store implements a mechanism to update or replace a slice of metrics.
func (s *metricService) Store(metric model.Metric, metrics ...model.Metric) error {
	var errFinal error

	// performing a separate call to avoid allocating another slice
	// for a subsequent loop
	err := s.storeSingle(metric)
	errFinal = errors.Join(errFinal, err)

	for i := range metrics {
		err := s.storeSingle(metrics[i])
		errFinal = errors.Join(errFinal, err)
	}

	return errFinal
}

// Retrieve implements a mechanism to retrive a slice of metrics by their hashes.
func (s *metricService) Retrieve(hash model.MetricHash, hashes ...model.MetricHash) ([]model.Metric, error) {
	var errFinal error

	metrics := make([]model.Metric, 0, len(hashes)+1)

	// performing a separate call to avoid allocating another slice
	// for a subsequent loop
	if m, err := s.storage.Get(hash); err != nil {
		errFinal = errors.Join(errFinal, err)
	} else {
		metrics = append(metrics, m)
	}

	for _, h := range hashes {
		m, err := s.storage.Get(h)
		if err != nil {
			errFinal = errors.Join(errFinal, err)
			continue
		}
		metrics = append(metrics, m)
	}

	return metrics, errFinal
}
