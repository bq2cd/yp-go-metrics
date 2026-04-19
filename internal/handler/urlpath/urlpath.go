package urlpath

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	// ErrInvalidURLPath is returned when HTTP URL does not match certain pattern.
	ErrInvalidURLPath = errors.New("invalid url path")
	// ErrMissingOperation is returned when operation on metric is missing from the URL.
	ErrMissingOperation = errors.New("missing operation")
	// ErrMissingMetricType is returned when metric type is missing from the URL.
	ErrMissingMetricType = errors.New("missing metric type")
	// ErrMissingMetricID is returned when metric ID is missing from the URL.
	ErrMissingMetricID = errors.New("missing metric id")
	// ErrMissingMetricValue is returned when metric value is missing from the URL.
	ErrMissingMetricValue = errors.New("missing metric value")
	// ErrInvalidOperation is returned when operation on metric is not supported.
	ErrInvalidOperation = errors.New("invalid operation")
	// ErrInvalidMetricType is returned when metric type is not supported.
	ErrInvalidMetricType = errors.New("invalid metric type")
	// ErrInvalidMetricValue is returned when metric value is not a valid number.
	ErrInvalidMetricValue = errors.New("invalid metric value")

	// OperationTypeUpdate represents metric update operation.
	OperationTypeUpdate = OperationType("update")
	// OperationTypeValue represents metric value read operation.
	OperationTypeValue = OperationType("value")

	reMultipleSlashes = regexp.MustCompile("/+")
)

// OperationType denotes an operation that we need to apply to a metric.
type OperationType string

// MetricOperation encapsulates details about metric and operation to be applied to it.
type MetricOperation struct {
	Type        OperationType
	MetricType  model.MetricType
	MetricID    string
	MetricValue string
}

// NewOperationFromURLPath parses an url path and returns a new metric operation object.
// No validation of the url path is performed during parsing.
// Use `ToMetric()` method on the resulting metric operation
// to perform validation.
func NewOperationFromURLPath(urlPath string) MetricOperation {
	var metricOp MetricOperation

	sanitizedPath := strings.TrimSpace(reMultipleSlashes.ReplaceAllString(urlPath, "/"))

	// strings.SplitN will produce a slice with at least a single element,
	// even if the initial string is empty.
	parts := strings.SplitN(strings.Trim(sanitizedPath, "/"), "/", 4)

	metricOp.Type = OperationType(strings.TrimSpace(parts[0]))

	if len(parts) > 1 {
		metricOp.MetricType = model.MetricType(strings.TrimSpace(parts[1]))
	}

	if len(parts) > 2 {
		metricOp.MetricID = strings.TrimSpace(parts[2])
	}

	if len(parts) > 3 {
		metricOp.MetricValue = strings.TrimSpace(parts[3])
	}

	return metricOp
}

// ToMetric creates current metric operation to a metric object
// or returns an error if such conversion is not possible,
// e.g. due to invalid values
func (mo MetricOperation) ToMetric() (model.Metric, error) {
	var metric model.Metric

	if mo.MetricType == model.MetricType("") {
		return metric, ErrMissingMetricType
	} else {
		metric.Type = mo.MetricType
	}

	if mo.MetricID == "" {
		return metric, ErrMissingMetricID
	} else {
		metric.ID = mo.MetricID
	}

	if mo.MetricValue == "" {
		return metric, ErrMissingMetricValue
	}

	switch metric.Type {
	case model.MetricTypeCounter:
		if v, err := strconv.ParseInt(mo.MetricValue, 10, 64); err != nil {
			return metric, fmt.Errorf("%w: %w", err, ErrInvalidMetricValue)
		} else {
			metric.Delta = &v
		}
	default:
		if v, err := strconv.ParseFloat(mo.MetricValue, 64); err != nil {
			return metric, fmt.Errorf("%w: %w", err, ErrInvalidMetricValue)
		} else {
			metric.Value = &v
		}
	}

	return metric, nil
}

// ToURLPath generates an url path from the given metric operation
// or returns an error if such conversion is not possible,
// e.g. due to missing values.
func (mo MetricOperation) ToURLPath() (string, error) {
	out := make([]string, 1, 5)

	switch mo.Type {
	case OperationType(""):
		return "", ErrMissingOperation
	default:
		out = append(out, string(mo.Type))
	}

	switch mo.MetricType {
	case model.MetricType(""):
		return "", ErrMissingMetricType
	default:
		out = append(out, string(mo.MetricType))
	}

	switch mo.MetricID {
	case "":
		return "", ErrMissingMetricID
	default:
		out = append(out, mo.MetricID)
	}

	switch mo.Type {
	case OperationTypeValue:
		// we're good here
	case OperationTypeUpdate:
		switch mo.MetricValue {
		case "":
			return "", ErrMissingMetricValue
		default:
			out = append(out, mo.MetricValue)
		}
	default:
		return "", ErrInvalidOperation
	}

	return strings.Join(out, "/"), nil
}

// NewOperationFromMetric creates a metric operation object
// from given operation type and metric object.
func NewOperationFromMetric(operation OperationType, metric model.Metric) MetricOperation {
	metricOp := MetricOperation{
		Type:       operation,
		MetricType: metric.Type,
		MetricID:   metric.ID,
	}

	switch metric.Type {
	case model.MetricTypeCounter:
		if metric.Delta != nil {
			metricOp.MetricValue = strconv.FormatInt(*metric.Delta, 10)
		}
	default:
		if metric.Value != nil {
			metricOp.MetricValue = strconv.FormatFloat(*metric.Value, 'g', 10, 64)
		}
	}

	return metricOp
}
