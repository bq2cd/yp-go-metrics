package handler

import (
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	// ErrPathConversionError is raised when conversion from URL path to Metric struct
	// fails.
	ErrPathConversionError = errors.New("failed to convert url path to metric")
)

// NewMetricFromURLPath converts an URL path (e.g. /update/TYPE/ID/VALUE)
// into a metric. Returns error if the conversion is not possible.
func NewMetricFromURLPath(path string) (model.Metric, error) {
	return model.Metric{}, nil
}
