package urlpath

import (
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	// errors
	ErrInvalidURLPath     = errors.New("invalid url path")
	ErrMissingOperation   = errors.New("missing operation")
	ErrMissingMetricType  = errors.New("missing metric type")
	ErrMissingMetricID    = errors.New("missing metric id")
	ErrMissingMetricValue = errors.New("missing metric value")
	ErrInvalidOperation   = errors.New("invalid operation")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrInvalidMetricValue = errors.New("invalid metric value")

	// supported operation types
	OperationTypeUpdate = OperationType("update")
	OperationTypeValue  = OperationType("value")
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
	return MetricOperation{}
}

// ToMetric creates current metric operation to a metric object
// or returns an error if such conversion is not possible,
// e.g. due to invalid values
func (mo *MetricOperation) ToMetric() (model.Metric, error) {
	return model.Metric{}, nil
}

// ToURLPath generates an url path from the given metric operation
// or returns an error if such conversion is not possible,
// e.g. due to missing values.
func (mo *MetricOperation) ToURLPath() (string, error) {
	return "", nil
}

// NewOperationFromMetric creates a metric operation object
// from given operation type and metric object.
func NewOperationFromMetric(operation OperationType, metric model.Metric) MetricOperation {
	return MetricOperation{}
}
