package source

import (
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/extra"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats"
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// Source abstracts a source of metrics and its underlying implementation.
type Source interface {
	ReadMetrics() ([]model.Metric, error)
}

// DefaultSources creates a slice of preconfigured metric sources.
func DefaultSources() []Source {
	return []Source{memstats.New(), extra.New()}
}
