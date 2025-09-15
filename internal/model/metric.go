package model

import "fmt"

const (
	// MetricTypeCounter defines a counter type
	MetricTypeCounter = MetricType("counter")

	// MetricTypeGauge defines a gauge type
	MetricTypeGauge = MetricType("gauge")
)

type (
	// MetricType wraps a string into a user type.
	MetricType string

	// MetricHash wraps a string into a user type.
	MetricHash string
)

// Metric defines a basic data structure to store metric values and metadata.
// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metric struct {
	ID    string     `json:"id"`
	Type  MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
	Hash  MetricHash `json:"hash,omitempty"`
}

// NewCounterMetric creates an instance of a counter type metric.
func NewCounterMetric(mID string, value int64) Metric {
	metric := Metric{
		ID:    mID,
		Type:  MetricTypeCounter,
		Delta: &value,
	}
	metric.updateHash()
	return metric
}

// NewGaugeMetric creates an instance of a counter type metric.
func NewGaugeMetric(mID string, value float64) Metric {
	metric := Metric{
		ID:    mID,
		Type:  MetricTypeGauge,
		Value: &value,
	}
	metric.updateHash()
	return metric
}

// It is very easy to mess up with the Metric struct because the fields are
// exported.
// TODO: find better solution.
func (m *Metric) updateHash() {
	m.Hash = MetricHash(fmt.Sprintf("%s/%s", m.Type, m.ID))
}
