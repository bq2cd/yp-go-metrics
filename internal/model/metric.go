package model

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

// NewMetricFromURLPath converts an URL path (e.g. /update/TYPE/ID/VALUE)
// into a metric. Returns error if the conversion is not possible.
func NewMetricFromURLPath(path string) (Metric, error) {
	return Metric{}, nil
}
