package extra

import (
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	once             sync.Once
	supportedMetrics map[string]model.MetricType
)

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

// ReadMetrics reads metrics from runtime.MemStats and converts them
// into internal representation.
func (s *source) ReadMetrics() ([]model.Metric, error) {
	s.pollCounter++
	return nil, nil
}
