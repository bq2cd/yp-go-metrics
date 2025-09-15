package repository

import (
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// MemStorage stores a map of metric hash to metric data in memory.
type memStorage struct {
	values map[model.MetricHash]model.Metric
}

// NewMemStorage initialises an empty memory storage
func NewMemStorage() *memStorage {
	return &memStorage{values: make(map[model.MetricHash]model.Metric)}
}

// Get retrieves a metric by its hash from in-memory map.
func (s *memStorage) Get(hash model.MetricHash) (model.Metric, error) {
	return model.Metric{}, nil
}

// Set stores a given metric to in-memory map.
func (s *memStorage) Set(metric model.Metric) error {
	return nil
}
