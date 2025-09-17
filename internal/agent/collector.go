package agent

import "github.com/bq2cd/yp-go-metrics/internal/model"

var (
	defaultRuntimeMetrics = map[string]model.MetricType{
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
	defaultExtraMetrics = map[string]model.MetricType{
		"PollCount":   model.MetricTypeCounter,
		"RandomValue": model.MetricTypeGauge,
	}
)

// Collector abstracts a way to collect metrics.
type Collector interface {
	Collect() ([]model.Metric, error)
}

type defaultCollector struct {
	pollCounter int
}

// NewDefaultCollector creates an instance of the default collector.
func NewDefaultCollector() *defaultCollector {
	return &defaultCollector{}
}

// Collect returns collected metrics.
func (c *defaultCollector) Collect() ([]model.Metric, error) {
	return nil, nil
}
