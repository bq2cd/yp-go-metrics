package sqlstorage

import (
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type sqlItem interface {
	GetID() string
	GetValue() any
	MetricType() model.MetricType
	ToMetric() model.Metric
	FromMetric(m model.Metric) sqlItem
}

type sqlItemCounter struct {
	ID    string `db:"metric_id"`
	Value int64  `db:"value"`
}

type sqlItemGauge struct {
	ID    string  `db:"metric_id"`
	Value float64 `db:"value"`
}

// Counter

func (it sqlItemCounter) GetID() string                { return it.ID }
func (it sqlItemCounter) GetValue() any                { return it.Value }
func (it sqlItemCounter) MetricType() model.MetricType { return model.MetricTypeCounter }
func (it sqlItemCounter) ToMetric() model.Metric       { return model.NewCounterMetric(it.ID, it.Value) }

func (it sqlItemCounter) FromMetric(m model.Metric) sqlItem {
	it.ID = m.ID
	if m.Delta != nil {
		it.Value = *m.Delta
	}
	return it
}

// Gauge

func (it sqlItemGauge) GetID() string                { return it.ID }
func (it sqlItemGauge) GetValue() any                { return it.Value }
func (it sqlItemGauge) MetricType() model.MetricType { return model.MetricTypeGauge }
func (it sqlItemGauge) ToMetric() model.Metric       { return model.NewGaugeMetric(it.ID, it.Value) }

func (it sqlItemGauge) FromMetric(m model.Metric) sqlItem {
	it.ID = m.ID
	if m.Value != nil {
		it.Value = *m.Value
	}
	return it
}

// Misc

func sqlAnySlice[S ~[]E, E any](s S) []any {
	out := make([]any, len(s))
	for i := range s {
		out[i] = s[i]
	}
	return out
}
