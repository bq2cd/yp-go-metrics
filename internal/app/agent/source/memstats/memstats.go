package memstats

import (
	"reflect"
	"runtime"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	once             sync.Once
	supportedMetrics map[string]model.MetricType
)

// GetSupportedMetrics returns a map of metric ID to metric type for all possible
// metrics that can be returned by this source.
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

type memStatsReader interface {
	ReadStats() reflect.Value
}

type memStats struct{}

// ReadStats invokes [runtime.ReadMemStats] and returns reflected value of [runtime.MemStats].
func (m *memStats) ReadStats() reflect.Value {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return reflect.ValueOf(stats)
}

type source struct {
	supportedMetrics map[string]model.MetricType
	reader           memStatsReader
}

// New creates an instance of [runtime.MemStats] metrics source.
func New() *source {
	return &source{supportedMetrics: GetSupportedMetrics(), reader: &memStats{}}
}

// ReadMetrics reads metrics from [runtime.MemStats] and converts them
// into internal representation.
func (s *source) ReadMetrics() ([]model.Metric, error) {
	metrics := make([]model.Metric, 0, len(s.supportedMetrics))

	stats := s.reader.ReadStats()

	for mID, mType := range s.supportedMetrics {
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

func castToInt64(v reflect.Value) (int64, bool) {
	if v.CanInt() {
		return v.Int(), true
	}
	if v.CanUint() {
		return int64(v.Uint()), true
	}
	return 0, false
}

func castToFloat64(v reflect.Value) (float64, bool) {
	if value, ok := castToInt64(v); ok {
		return float64(value), true
	}
	if v.CanFloat() {
		return v.Float(), true
	}
	return 0, false
}
