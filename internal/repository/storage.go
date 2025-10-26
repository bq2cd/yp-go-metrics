package repository

import (
	"context"
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	// ErrMetricNotFound is raised when metric does not exist in the storage.
	ErrMetricNotFound = errors.New("metric not found")
)

type StorageReader interface {
	Get(ctx context.Context, key model.MetricKey) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)
}

type StorageWriter interface {
	Set(ctx context.Context, metric model.Metric) error
}

// Storage abstracts an underlying technology for storing metrics.
type Storage interface {
	StorageReader
	StorageWriter
}

type StorageMultiReader interface {
	StorageReader
	GetMulti(ctx context.Context, keys model.MetricKeySet) ([]model.Metric, error)
}

type StorageMultiWriter interface {
	StorageWriter
	SetMulti(ctx context.Context, metrics model.MetricSet) error
}

// StorageMulti enhances [Storage] with batch capabilities.
type StorageMulti interface {
	StorageMultiReader
	StorageMultiWriter
}
