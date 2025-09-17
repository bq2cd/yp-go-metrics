package agent

import (
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

// Storer abstracts a way to store metrics locally before reporting to an upstream.
// Typically uses an in-memory storage.
type Storer interface {
	Store(metrics []model.Metric) error
	Retrieve() ([]model.Metric, error)
}

type defaultStorer struct {
	storage service.Metrics
}

// NewDefaultStorer creates an instance of the default storer
// backed by in-memory storage.
func NewDefaultStorer(storage service.Metrics) *defaultStorer {
	return &defaultStorer{storage: storage}
}

// Store receives incoming metrics and stores them in the underlying storage.
func (s *defaultStorer) Store(metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	return s.storage.Store(metrics[0], metrics[1:]...)
}

// Retrieve outputs all metrics available in the underlying storage
func (s *defaultStorer) Retrieve() ([]model.Metric, error) {
	return s.storage.RetrieveAll()
}
