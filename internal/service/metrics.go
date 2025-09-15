package service

import (
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
	return nil
}
