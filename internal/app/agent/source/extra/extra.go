package extra

import (
	"math/rand/v2"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	once             sync.Once
	supportedMetrics map[string]model.MetricType
)

// GetSupportedMetrics returns a map of metric ID to metric type,
// that can be produced by this source.
func GetSupportedMetrics() map[string]model.MetricType {
	once.Do(func() {
		supportedMetrics = map[string]model.MetricType{
			"PollCount":   model.MetricTypeCounter,
			"RandomValue": model.MetricTypeGauge,
		}
	})
	return supportedMetrics
}

type source struct {
	pollCounter int
}

// New creates an instance of extra metrics source.
func New() *source {
	return &source{}
}

// AvailableMetricNames returns a map of unique metric names to their types, that can be collected
// through this source.
func (s *source) AvailableMetricNames() map[string]model.MetricType {
	return GetSupportedMetrics()
}

// ReadMetrics reads custom metrics and converts them
// into internal representation.
func (s *source) ReadMetrics() ([]model.Metric, error) {
	s.pollCounter++

	metrics := []model.Metric{
		model.NewCounterMetric("PollCount", int64(s.pollCounter)),
		model.NewGaugeMetric("RandomValue", rand.Float64()),
	}

	return metrics, nil
}
