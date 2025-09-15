package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	// ErrPathConversionError is raised when conversion from URL path to Metric struct
	// fails.
	ErrPathConversionError = errors.New("path to metric conversion error")
)

// NewMetricFromURLPath converts an URL path (e.g. /update/TYPE/ID/VALUE)
// into a metric. Returns error if the conversion is not possible.
func NewMetricFromURLPath(path string) (model.Metric, error) {
	var metric model.Metric

	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) != 4 {
		return metric, fmt.Errorf("expected 4 path parts: %w", ErrPathConversionError)
	}

	if parts[0] != "update" {
		return metric, fmt.Errorf("expected update operation: %w", ErrPathConversionError)
	}

	mType := model.MetricType(parts[1])
	mID := strings.TrimSpace(parts[2])
	mValue := strings.TrimSpace(parts[3])

	if mID == "" {
		return metric, fmt.Errorf("empty id: %w", ErrPathConversionError)
	}

	switch mType {
	case model.MetricTypeCounter:
		parsed, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return metric, fmt.Errorf("invalid counter value: %w: %w", err, ErrPathConversionError)
		}
		metric = model.NewCounterMetric(mID, parsed)
	case model.MetricTypeGauge:
		parsed, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return metric, fmt.Errorf("invalid gauge value: %w: %w", err, ErrPathConversionError)
		}
		metric = model.NewGaugeMetric(mID, parsed)
	default:
		return metric, fmt.Errorf("invalid metric type: %w", ErrPathConversionError)
	}

	return metric, nil
}
