package agent

import (
	"math/rand/v2"
	"reflect"
	"runtime"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

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
	c.pollCounter++
	randomValue := rand.Float64()

	metrics := make([]model.Metric, 0, len(defaultRuntimeMetrics)+len(defaultExtraMetrics))

	metrics = append(metrics, model.NewCounterMetric("PollCount", int64(c.pollCounter)))
	metrics = append(metrics, model.NewGaugeMetric("RandomValue", randomValue))

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := reflect.ValueOf(memStats)

	castToInt64 := func(v reflect.Value) (int64, bool) {
		if v.CanInt() {
			return v.Int(), true
		}
		if v.CanUint() {
			return int64(v.Uint()), true
		}
		return 0, false
	}

	castToFloat64 := func(v reflect.Value) (float64, bool) {
		if value, ok := castToInt64(v); ok {
			return float64(value), true
		}
		if v.CanFloat() {
			return v.Float(), true
		}
		return 0, false
	}

	for mID, mType := range defaultRuntimeMetrics {
		value := stats.FieldByName(mID)
		var metric model.Metric
		switch mType {
		case model.MetricTypeCounter:
			mValue, ok := castToInt64(value)
			if !ok {
				continue
			}
			metric = model.NewCounterMetric(mID, mValue)
		case model.MetricTypeGauge:
			mValue, ok := castToFloat64(value)
			if !ok {
				continue
			}
			metric = model.NewGaugeMetric(mID, mValue)
		default:
			continue
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
