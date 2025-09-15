package repository

import (
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// Storage abstracts an underlying technology for storing metrics.
type Storage interface {
	Get(hash model.MetricHash) (model.Metric, error)
	Set(metric model.Metric) error
}
