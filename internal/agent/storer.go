package agent

import (
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

// Storer abstracts a way to store metrics locally before reporting to an upstream.
// Typically uses an in-memory storage.
type Storer interface {
	Store(metrics []model.Metric) error
	Retrieve() ([]model.Metric, error)
}

type defaultStorer struct {
	storage repository.Storage
}

// NewDefaultStorer creates an instance of the default storer
// backed by in-memory storage.
func NewDefaultStorer(storage repository.Storage) *defaultStorer {
	return &defaultStorer{storage: storage}
}

// Store receives incoming metrics and stores them in the underlying storage.
func (s *defaultStorer) Store(metrics []model.Metric) error {
	var errFinal error
	for _, m := range metrics {
		errFinal = errors.Join(errFinal, s.storage.Set(m))
	}
	return errFinal
}

// Retrieve outputs all metrics available in the underlying storage
func (s *defaultStorer) Retrieve() ([]model.Metric, error) {
	return s.storage.GetAll()
}
