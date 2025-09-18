package agent

import (
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// Collector abstracts a way to collect metrics.
type Collector interface {
	Collect() ([]model.Metric, error)
}

type defaultCollector struct {
	sources []source.Source
}

// NewDefaultCollector creates an instance of the default collector.
func NewDefaultCollector() *defaultCollector {
	return &defaultCollector{sources: source.DefaultSources()}
}

// Collect returns collected metrics.
func (c *defaultCollector) Collect() ([]model.Metric, error) {
	var errFinal error

	metrics := make([]model.Metric, 0)

	for _, src := range c.sources {
		mm, err := src.ReadMetrics()
		metrics = append(metrics, mm...)
		errFinal = errors.Join(errFinal, err)
	}

	return metrics, errFinal
}
