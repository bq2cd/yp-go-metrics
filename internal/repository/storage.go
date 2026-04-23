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

// StorageReader provides methods for reading either a single metric or all metrics.
type StorageReader interface {
	Get(ctx context.Context, key model.MetricKey) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)
}

// StorageWriter provides a method to update single metric.
type StorageWriter interface {
	Set(ctx context.Context, metric model.Metric) error
}

// Storage provides methods for reading and updating metrics.
// It combines [StorageReader] and [StorageWriter] interfaces.
type Storage interface {
	StorageReader
	StorageWriter
}

// StorageMultiReader enhances [StorageReader] with methods for reading multiple metrics at a time.
type StorageMultiReader interface {
	StorageReader
	GetMulti(ctx context.Context, keys model.MetricKeySet) ([]model.Metric, error)
}

// StorageMultiWriter enhances [StorageWriter] with methods for updating multiple metrics at a time.
type StorageMultiWriter interface {
	StorageWriter
	SetMulti(ctx context.Context, metrics model.MetricSet) error
}

// StorageMulti enhances [Storage] with batch capabilities: it provides methods for reading or updating
// multiple metrics at a time.
// It combines [StorageMultiReader] and [StorageMultiWriter] interfaces.
type StorageMulti interface {
	StorageMultiReader
	StorageMultiWriter
}
