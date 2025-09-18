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

	// strings.Split will produce a slice with at least a single element,
	// even if the initial string is empty.
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// validate operation
	operation := parts[0]

	switch operation {
	case "update":
		// ok
	default:
		return metric, fmt.Errorf("invalid operation: %w", ErrInvalidURLPath)
	}

	// validate metric type
	if len(parts) < 2 {
		return metric, fmt.Errorf("missing metric type: %w", ErrInvalidURLPath)
	}

	mType := model.MetricType(parts[1])

	switch mType {
	case model.MetricTypeCounter:
	case model.MetricTypeGauge:
	default:
		return metric, fmt.Errorf("invalid metric type: %w", ErrInvalidURLPath)
	}

	// validate metric ID
	if len(parts) < 3 {
		return metric, ErrEmptyMetricID
	}

	mID := strings.TrimSpace(parts[2])

	if mID == "" {
		return metric, ErrEmptyMetricID
	}

	// validate metric value
	if len(parts) < 4 {
		return metric, fmt.Errorf("missing metric value: %w", ErrInvalidURLPath)
	}

	mValue := strings.TrimSpace(parts[3])

	// fail if path contains extra elements
	if len(parts) > 4 {
		return metric, fmt.Errorf("extra path elements: %w", ErrInvalidURLPath)
	}

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
		// we already check for invalid metric type above, so this
		// branch is unreachable
	}

	return metric, nil
}
