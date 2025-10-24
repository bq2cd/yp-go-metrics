package storagetest

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

const (
	FaultyStorageErrorTrigger = "faultyStorageErrorTrigger"
)

var (
	ErrFaultyStorage = errors.New("faulty storage")
)

type MockStorage struct {
	mu        sync.RWMutex
	data      map[model.MetricKey]model.Metric
	isFaulty  bool
	triggerID string
}

// NewMockStorage creates an instance of an in-memory storage operating in normal mode.
// To switch storage into faulty mode, see [MakeFaulty] method.
func NewMockStorage(metrics ...model.Metric) *MockStorage {
	data := make(map[model.MetricKey]model.Metric, len(metrics))
	for _, m := range metrics {
		data[m.Key()] = m
	}
	return &MockStorage{
		data:      data,
		isFaulty:  false,
		triggerID: FaultyStorageErrorTrigger,
	}
}

// MakeFaulty activates faulty mode of operation for the storage instance.
func (s *MockStorage) MakeFaulty() *MockStorage {
	s.isFaulty = true
	return s
}

// MakeNormal deactivates faulty mode of operation for the storage instance.
func (s *MockStorage) MakeNormal() *MockStorage {
	s.isFaulty = false
	return s
}

// Get retrieves a metric by a given key from the storage.
// If metric does not exist, [repository.ErrMetricNotFound] error is returned.
// If storage is faulty and fault is triggered by a special metric ID,
// then [ErrFaultyStorage] error is returned.
func (s *MockStorage) Get(_ context.Context, k model.MetricKey) (model.Metric, error) {
	if s.isFaulty && k.ID == s.triggerID {
		return model.Metric{}, fmt.Errorf("get error: %w", ErrFaultyStorage)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.data[k]
	if !ok {
		return model.Metric{}, repository.ErrMetricNotFound
	}
	return m, nil
}

// GetAll retrieves all metrics currently stored in the storage.
// It relies on [Get] method to retrieve a single metric,
// so returned error depends on the result of that method.
// If [repository.ErrMetricNotFound] was returned by [Get],
// then this metric is not returned.
// If [ErrFaultyStorage] was returned by [Get], it is immediately
// returned along with `nil` (empty slice).
func (s *MockStorage) GetAll(ctx context.Context) ([]model.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Metric, 0, len(s.data))
	for k := range s.data {
		m, err := s.Get(ctx, k)
		switch {
		case err == nil:
			out = append(out, m)
		case errors.Is(err, ErrFaultyStorage):
			return nil, err
		default:
			continue
		}
	}
	return out, nil
}

// Set stores given metric in the underlying storage.
// If faulty mode is activated and metric's ID matches a trigger ID,
// then [ErrFaultyStorage] is immediately returned.
func (s *MockStorage) Set(_ context.Context, m model.Metric) error {
	if s.isFaulty && m.ID == s.triggerID {
		return fmt.Errorf("set error: %w", ErrFaultyStorage)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[m.Key()] = m
	return nil
}
