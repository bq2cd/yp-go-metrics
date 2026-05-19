package psutil

import (
	"sync"

	pscpu "github.com/shirou/gopsutil/v4/cpu"
	psmem "github.com/shirou/gopsutil/v4/mem"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

const (
	nameTotalMemory    = "TotalMemory"
	nameFreeMemory     = "FreeMemory"
	nameCPUutilisation = "CPUutilization1"
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
			nameTotalMemory:    model.MetricTypeGauge,
			nameFreeMemory:     model.MetricTypeGauge,
			nameCPUutilisation: model.MetricTypeGauge,
		}
	})
	return supportedMetrics
}

type readMetricsFn func() ([]model.Metric, error)

type source struct{}

// New creates an instance of gopsutil metrics source.
func New() *source {
	return &source{}
}

func (s *source) readMemoryMetrics() ([]model.Metric, error) {
	vmstat, err := psmem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return []model.Metric{
		model.NewGaugeMetric(nameTotalMemory, float64(vmstat.Total)),
		model.NewGaugeMetric(nameFreeMemory, float64(vmstat.Free)),
	}, nil
}

func (s *source) readCPUMetrics() ([]model.Metric, error) {
	data, err := pscpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	return []model.Metric{model.NewGaugeMetric(nameCPUutilisation, data[0])}, nil
}

// AvailableMetricNames returns a map of unique metric names to their types, that can be collected
// through this source.
func (s *source) AvailableMetricNames() map[string]model.MetricType {
	return GetSupportedMetrics()
}

// ReadMetrics reads supported metrics from gopsutil and converts them
// into internal representation.
func (s *source) ReadMetrics() ([]model.Metric, error) {
	allMetrics := make([]model.Metric, 0, len(supportedMetrics))

	for _, fn := range []readMetricsFn{
		s.readMemoryMetrics,
		s.readCPUMetrics,
	} {
		metrics, err := fn()
		if err != nil {
			return nil, err
		}
		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, nil
}
