package repository

import (
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	// ErrMetricNotFound is raised when metric does not exist in the storage.
	ErrMetricNotFound = errors.New("metric not found")
)

// Storage abstracts an underlying technology for storing metrics.
type Storage interface {
	Get(hash model.MetricHash) (model.Metric, error)
	Set(metric model.Metric) error
}
