package service

import (
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

// Metrics defines an interface to work with metrics.
// E.g. update, retrieve, delete, etc.
type Metrics interface {
	Update(metric model.Metric) error
}

type metricService struct {
	storage repository.Storage
}

// NewMetrics creates an instance of the metrics service.
func NewMetrics(storage repository.Storage) *metricService {
	return &metricService{storage: storage}
}

// Update implements a mechanism to update or replace a given metric.
func (s *metricService) Update(metric model.Metric) error {
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
