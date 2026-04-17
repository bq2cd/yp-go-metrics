package service

import (
	"io"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// MetricEncoder encodes a slice of metrics and writes the result into the provided writer.
type MetricEncoder interface {
	EncodeBatch(w io.Writer, metrics []model.Metric) error
}

type metricJSONEncoder struct{}

// NewMetricJSONEncoder creates an instance of a JSON encoder of metrics.
func NewMetricJSONEncoder() *metricJSONEncoder {
	return &metricJSONEncoder{}
}

// EncodeBatch encodes the provided metrics into JSON and writes the result into the provided writer.
func (d *metricJSONEncoder) EncodeBatch(w io.Writer, metrics []model.Metric) error {
	if metrics == nil {
		metrics = []model.Metric{}
	}
	return json.NewEncoder(w).Encode(metrics)
}
