package agent

import (
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

// Collector abstracts a way to collect metrics.
type Collector interface {
	Collect() error
	Snapshot() ([]model.Metric, error)
}

type defaultCollector struct {
	sources   []source.Source
	collected repository.Storage
}

// NewDefaultCollector creates an instance of the default collector
// with default metric sources and in-memory internal storage.
func NewDefaultCollector() *defaultCollector {
	return &defaultCollector{sources: source.DefaultSources(), collected: repository.NewMemStorage()}
}

// NewCollector creates an instance of the default collector with specific
// metric sources and internal storage.
func NewCollector(sources []source.Source, storage repository.Storage) *defaultCollector {
	return &defaultCollector{sources: sources, collected: storage}
}

func (c *defaultCollector) storeMetrics(metrics []model.Metric) error {
	var errFinal error
	for _, m := range metrics {
		errFinal = errors.Join(errFinal, c.collected.Set(m))
	}
	return errFinal
}

// Collect queries metric sources and stores obtained metrics
// into the internal storage.
func (c *defaultCollector) Collect() error {
	var errFinal error
	for _, src := range c.sources {
		metrics, err := src.ReadMetrics()
		errFinal = errors.Join(errFinal, err, c.storeMetrics(metrics))
	}
	return errFinal
}

// Snapshot returns latest values of all metrics from the internal storage.
func (c *defaultCollector) Snapshot() ([]model.Metric, error) {
	return c.collected.GetAll()
}
