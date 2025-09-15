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
	ErrInvalidURLPath = errors.New("invalid url path")
	ErrEmptyMetricID  = errors.New("empty metric id")
)

// NewMetricFromURLPath converts an URL path (e.g. /update/TYPE/ID/VALUE)
// into a metric. Returns error if the conversion is not possible.
func NewMetricFromURLPath(path string) (model.Metric, error) {
	var metric model.Metric

	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 2 {
		return metric, fmt.Errorf("expected at least 2 path parts: %w", ErrInvalidURLPath)
	}

	if parts[0] != "update" {
		return metric, fmt.Errorf("expected update operation: %w", ErrInvalidURLPath)
	}

	mType := model.MetricType(parts[1])
	switch mType {
	case model.MetricTypeCounter:
	case model.MetricTypeGauge:
	default:
		return metric, fmt.Errorf("invalid metric type: %w", ErrInvalidURLPath)
	}

	if len(parts) < 3 {
		return metric, ErrEmptyMetricID
	}

	mID := strings.TrimSpace(parts[2])

	if mID == "" {
		return metric, ErrEmptyMetricID
	}

	if len(parts) != 4 {
		return metric, fmt.Errorf("expected exactly 4 path parts: %w", ErrInvalidURLPath)
	}

	mValue := strings.TrimSpace(parts[3])

	switch mType {
	case model.MetricTypeCounter:
		parsed, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return metric, fmt.Errorf("invalid counter value: %w: %w", err, ErrInvalidURLPath)
		}
		metric = model.NewCounterMetric(mID, parsed)
	case model.MetricTypeGauge:
		parsed, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return metric, fmt.Errorf("invalid gauge value: %w: %w", err, ErrInvalidURLPath)
		}
		metric = model.NewGaugeMetric(mID, parsed)
	default:
		return metric, fmt.Errorf("invalid metric type: %w", ErrInvalidURLPath)
	}

	return metric, nil
}
