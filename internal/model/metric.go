package model

import "slices"

const (
	_metricTypeEmpty = MetricType("")

	// MetricTypeCounter defines a counter type
	MetricTypeCounter = MetricType("counter")

	// MetricTypeGauge defines a gauge type
	MetricTypeGauge = MetricType("gauge")
)

const (
	MetricAggregateStrategyLastValueWins MetricAggregateStrategy = iota
	MetricAggregateStrategyFirstValueWins
	MetricAggregateStrategyCounterValueAccumulates
)

type (
	// MetricType wraps a string into a user type.
	MetricType string

	// MetricHash wraps a string into a user type.
	MetricHash string

	// MetricAggregateStrategy defines how to aggregate a list of metrics by a unique [MetricKey].
	MetricAggregateStrategy int
)

// MetricKey defines a (type, id) pair to be used as a map key.
// This is different from metric hash as the hash can have a different
// logic behind it.
type MetricKey struct {
	Type MetricType `json:"type"`
	ID   string     `json:"id"`
}

// NewMetricKey is a convenience function to create a MetricKey struct instance.
func NewMetricKey(mType MetricType, mID string) MetricKey {
	return MetricKey{Type: mType, ID: mID}
}

// Empty returns true if either type or id or both are missing
func (k MetricKey) Empty() bool {
	if k.Type == _metricTypeEmpty {
		return true
	}
	if k.ID == "" {
		return true
	}
	return false
}

// Compare simulates [Comparable] interface for [MetricKey] struct.
func (k MetricKey) Compare(other MetricKey) int {
	return slices.Compare(
		[]string{string(k.Type), k.ID},
		[]string{string(other.Type), other.ID},
	)
}

// MetricKeySet represents unique metric keys.
type MetricKeySet map[MetricKey]struct{}

// NewMetricKeySet create a new [MetricKeySet] from a list of keys.
// The last seen key wins.
func NewMetricKeySet(keys ...MetricKey) MetricKeySet {
	unique := MetricKeySet{}
	for _, key := range keys {
		unique[key] = struct{}{}
	}
	return unique
}

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
	return metric
}

// NewGaugeMetric creates an instance of a counter type metric.
func NewGaugeMetric(mID string, value float64) Metric {
	metric := Metric{
		ID:    mID,
		Type:  MetricTypeGauge,
		Value: &value,
	}
	return metric
}

// Key returns a MetricKey struct for this metric.
func (m *Metric) Key() MetricKey {
	return MetricKey{Type: m.Type, ID: m.ID}
}

// Empty returns true if either of the following is true:
// 1) Metric id is empty.
// 2) Metric type is empty.
// 3) Metric type is counter and delta is nil.
// 4) Metric type is gauge and value is nil.
// 5) Metric type is custom and both value and delta are nil.
func (m *Metric) Empty() bool {
	if m.ID == "" {
		return true
	}
	if m.Type == MetricType("") {
		return true
	}
	switch m.Type {
	case MetricTypeCounter:
		if m.Delta == nil {
			return true
		}
	case MetricTypeGauge:
		if m.Value == nil {
			return true
		}
	default:
		if m.Delta == nil && m.Value == nil {
			return true
		}
	}
	return false
}

// Copy creates a duplicate of the given metric.
func (m *Metric) Copy() Metric {
	metric := Metric{
		Type: m.Type,
		ID:   m.ID,
		Hash: m.Hash,
	}
	if m.Delta != nil {
		tmp := *m.Delta
		metric.Delta = &tmp
	}
	if m.Value != nil {
		tmp := *m.Value
		metric.Value = &tmp
	}
	return metric
}

// AddDelta adds provided value to the metric [Delta] field but only if metric type is [MetricTypeCounter] and new value is not nil.
func (m *Metric) AddDelta(other *int64) {
	if m.Type != MetricTypeCounter {
		return
	}
	if other == nil {
		return
	}
	if m.Delta == nil {
		m.Delta = other
		return
	}
	*m.Delta += *other
}

// MetricSet represents a set of unique metrics keyed by their [MetricKey].
type MetricSet map[MetricKey]Metric

// NewMetricSet converts a list of metrics into unique set of metrics, keyed by [MetricKey] using [MetricAggregateStrategyLastValueWins].
func NewMetricSet(metrics ...Metric) MetricSet {
	return NewMetricSetWithStrategy(MetricAggregateStrategyLastValueWins, metrics...)
}

// NewMetricSetWithStrategy converts a list of metrics into unique set of metrics, keyed by [MetricKey] using provided [MetricAggregateStrategy].
func NewMetricSetWithStrategy(strategy MetricAggregateStrategy, metrics ...Metric) MetricSet {
	unique := make(MetricSet, len(metrics))
	for _, m := range metrics {
		if m.Empty() {
			continue
		}
		prev, ok := unique[m.Key()]
		switch strategy {
		case MetricAggregateStrategyFirstValueWins:
			if ok {
				continue
			}
		case MetricAggregateStrategyCounterValueAccumulates:
			switch m.Type {
			case MetricTypeCounter:
				if ok {
					m = m.Copy()
					m.AddDelta(prev.Delta)
				}
			}
		}
		unique[m.Key()] = m
	}
	return unique
}

// GroupByType create a separate metric set per metric type and
// maps them together.
func (ms MetricSet) GroupByType() map[MetricType]MetricSet {
	group := make(map[MetricType]MetricSet)
	for _, m := range ms {
		_, ok := group[m.Type]
		if !ok {
			group[m.Type] = make(MetricSet)
		}
		group[m.Type][m.Key()] = m
	}
	return group
}

// Keys return only metric keys in a [MetricKeySet] form.
func (ms MetricSet) Keys() MetricKeySet {
	keys := make(MetricKeySet, len(ms))
	for key := range ms {
		keys[key] = struct{}{}
	}
	return keys
}
