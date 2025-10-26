package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type memStorage struct {
	data model.MetricSet
	mu   sync.RWMutex
}

// NewMemStorage initializes an empty memory storage
func NewMemStorage() *memStorage {
	return &memStorage{data: model.NewMetricSet()}
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

// GetMulti retrieves multiple metrics in a single go (essentially calling [Get] for each metric).
func (s *memStorage) GetMulti(ctx context.Context, keys model.MetricKeySet) ([]model.Metric, error) {
	var errFinal error
	metrics := make([]model.Metric, 0, len(keys))
	for key := range keys {
		m, err := s.Get(ctx, key)
		if err == ErrMetricNotFound {
			continue
		}
		errFinal = errors.Join(errFinal, err)
		if !m.Empty() {
			metrics = append(metrics, m)
		}
	}
	return metrics, errFinal
}

// SetMulti stores multiple metrics in a single go (essentially calling [Set] for each metric).
func (s *memStorage) SetMulti(ctx context.Context, metrics model.MetricSet) error {
	var errFinal error
	for _, m := range metrics {
		errFinal = errors.Join(errFinal, s.Set(ctx, m))
	}
	return errFinal
}
