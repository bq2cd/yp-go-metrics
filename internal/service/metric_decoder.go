package service

import (
	"io"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// MetricDecoder reads from the provided reader and attempts to decode a slice of metrics from it.
type MetricDecoder interface {
	DecodeBatch(r io.Reader) ([]model.Metric, error)
}

type metricJSONDecoder struct{}

// NewMetricJSONDecoder creates an instance of a JSON decoder of metrics.
func NewMetricJSONDecoder() *metricJSONDecoder {
	return &metricJSONDecoder{}
}

// DecodeBatch reads JSON from the provided reader and attempts to decode it to the slice of metrics.
func (d *metricJSONDecoder) DecodeBatch(r io.Reader) ([]model.Metric, error) {
	metrics := []model.Metric{}
	err := json.NewDecoder(r).Decode(&metrics)
	return metrics, err
}
