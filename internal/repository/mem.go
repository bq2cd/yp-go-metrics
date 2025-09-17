package repository

import (
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type memStorageData map[model.MetricKey]model.Metric

type memStorage struct {
	data memStorageData
}

// NewMemStorage initialises an empty memory storage
func NewMemStorage() *memStorage {
	return &memStorage{data: make(memStorageData)}
}

// Get retrieves a metric by its hash from in-memory map.
func (s *memStorage) Get(key model.MetricKey) (model.Metric, error) {
	var metric model.Metric
	metric, ok := s.data[key]
	if !ok {
		return metric, ErrMetricNotFound
	}
	return metric, nil
}

// Set stores a given metric to in-memory map.
func (s *memStorage) Set(metric model.Metric) error {
	s.data[metric.Key()] = metric
	return nil
}
