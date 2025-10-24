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

// Storage abstracts an underlying technology for storing metrics.
type Storage interface {
	Get(ctx context.Context, key model.MetricKey) (model.Metric, error)
	GetAll(ctx context.Context) ([]model.Metric, error)
	Set(ctx context.Context, metric model.Metric) error
}
