package model

const (
	_metricTypeEmpty = MetricType("")

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
// - metric id is empty
// - metric type is empty
// - metric type is counter and delta is nil
// - metric type is gauge and value is nil
// - metric type is custom and both value and delta are nil
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

// MetricSet represents a slice of metrics.
type MetricSet []Metric

// UniqueByKey converts a slice of metrics into a map with
// `MetricKey` as a key and `Metric` as a value.
// During conversion, duplicates are eliminated; last seen
// metric in the slice wins.
func (ms MetricSet) UniqueByKey() map[MetricKey]Metric {
	unique := make(map[MetricKey]Metric, len(ms))
	for _, m := range ms {
		unique[m.Key()] = m
	}
	return unique
}
