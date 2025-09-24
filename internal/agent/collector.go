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

type collector struct {
	sources   []source.Source
	collected repository.Storage
}

// NewCollector creates an instance of the default collector with specific
// metric sources and internal storage.
func NewCollector(sources []source.Source, storage repository.Storage) *collector {
	return &collector{sources: sources, collected: storage}
}

func (c *collector) storeMetrics(metrics []model.Metric) error {
	var errFinal error
	for _, m := range metrics {
		errFinal = errors.Join(errFinal, c.collected.Set(m))
	}
	return errFinal
}

// Collect queries metric sources and stores obtained metrics
// into the internal storage.
func (c *collector) Collect() error {
	var errFinal error
	for _, src := range c.sources {
		metrics, err := src.ReadMetrics()
		errFinal = errors.Join(errFinal, err, c.storeMetrics(metrics))
	}
	return errFinal
}

// Snapshot returns latest values of all metrics from the internal storage.
func (c *collector) Snapshot() ([]model.Metric, error) {
	return c.collected.GetAll()
}
