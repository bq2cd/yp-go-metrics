package source

import (
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/extra"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil"
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

//go:generate go tool mockgen -destination=sourcetest/mock_source.go -package sourcetest github.com/bq2cd/yp-go-metrics/internal/app/agent/source Source

// Source abstracts a source of metrics and its underlying implementation.
type Source interface {
	AvailableMetricNames() map[string]model.MetricType
	ReadMetrics() ([]model.Metric, error)
}

// DefaultSources creates a slice of preconfigured metric sources.
func DefaultSources() []Source {
	return []Source{memstats.New(), extra.New(), psutil.New()}
}
