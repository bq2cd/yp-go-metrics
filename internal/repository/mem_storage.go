package repository

import (
	"context"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type memStorageData map[model.MetricKey]model.Metric

type memStorage struct {
	data memStorageData
	mu   sync.RWMutex
}

// NewMemStorage initialises an empty memory storage
func NewMemStorage() *memStorage {
	return &memStorage{data: make(memStorageData)}
}

// Get retrieves a metric by its hash from in-memory map.
func (s *memStorage) Get(_ context.Context, key model.MetricKey) (model.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var metric model.Metric
	metric, ok := s.data[key]
	if !ok {
		return metric, ErrMetricNotFound
	}
	if metric.Empty() {
		return model.Metric{}, ErrMetricNotFound
	}
	return metric, nil
}

// GetAll returns all metrics currently stored in the in-memory map.
func (s *memStorage) GetAll(_ context.Context) ([]model.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]model.Metric, 0, len(s.data))
	for k := range s.data {
		m := s.data[k]
		if m.Empty() {
			continue
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// Set stores a given metric to in-memory map.
func (s *memStorage) Set(_ context.Context, metric model.Metric) error {
	if metric.Empty() {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[metric.Key()] = metric
	return nil
}
