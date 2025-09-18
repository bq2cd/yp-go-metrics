package memstats

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
			"Alloc":         model.MetricTypeGauge,
			"BuckHashSys":   model.MetricTypeGauge,
			"Frees":         model.MetricTypeGauge,
			"GCCPUFraction": model.MetricTypeGauge,
			"GCSys":         model.MetricTypeGauge,
			"HeapAlloc":     model.MetricTypeGauge,
			"HeapIdle":      model.MetricTypeGauge,
			"HeapInuse":     model.MetricTypeGauge,
			"HeapObjects":   model.MetricTypeGauge,
			"HeapReleased":  model.MetricTypeGauge,
			"HeapSys":       model.MetricTypeGauge,
			"LastGC":        model.MetricTypeGauge,
			"Lookups":       model.MetricTypeGauge,
			"MCacheInuse":   model.MetricTypeGauge,
			"MCacheSys":     model.MetricTypeGauge,
			"MSpanInuse":    model.MetricTypeGauge,
			"MSpanSys":      model.MetricTypeGauge,
			"Mallocs":       model.MetricTypeGauge,
			"NextGC":        model.MetricTypeGauge,
			"NumForcedGC":   model.MetricTypeGauge,
			"NumGC":         model.MetricTypeGauge,
			"OtherSys":      model.MetricTypeGauge,
			"PauseTotalNs":  model.MetricTypeGauge,
			"StackInuse":    model.MetricTypeGauge,
			"StackSys":      model.MetricTypeGauge,
			"Sys":           model.MetricTypeGauge,
			"TotalAlloc":    model.MetricTypeGauge,
		}
	})
	return supportedMetrics
}

type source struct{}

// New creates an instance of runtime.MemStats metrics source.
func New() *source {
	return &source{}
}

// ReadMetrics reads metrics from runtime.MemStats and converts them
// into internal representation.
func (s *source) ReadMetrics() ([]model.Metric, error) {
	return nil, nil
}
