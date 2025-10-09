package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCounterMetric(t *testing.T) {
	type args struct {
		mID   string
		value int64
	}
	tests := []struct {
		name string
		args args
		want func() Metric
	}{
		{
			name: "zero",
			args: args{mID: "id1", value: 0},
			want: func() Metric {
				var value int64 = 0
				return Metric{
					ID:    "id1",
					Type:  MetricTypeCounter,
					Delta: &value,
					Value: nil,
				}
			},
		},
		{
			name: "positive value",
			args: args{mID: "id1", value: 10},
			want: func() Metric {
				var value int64 = 10
				return Metric{
					ID:    "id1",
					Type:  MetricTypeCounter,
					Delta: &value,
					Value: nil,
				}
			},
		},
		{
			name: "negative value",
			args: args{mID: "id1", value: -10},
			want: func() Metric {
				var value int64 = -10
				return Metric{
					ID:    "id1",
					Type:  MetricTypeCounter,
					Delta: &value,
					Value: nil,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want(), NewCounterMetric(tt.args.mID, tt.args.value))
		})
	}
}

func TestNewGaugeMetric(t *testing.T) {
	type args struct {
		mID   string
		value float64
	}
	tests := []struct {
		name string
		args args
		want func() Metric
	}{
		{
			name: "zero",
			args: args{mID: "id1", value: 0.0},
			want: func() Metric {
				var value = 0.0
				return Metric{
					ID:    "id1",
					Type:  MetricTypeGauge,
					Delta: nil,
					Value: &value,
				}
			},
		},
		{
			name: "positive value",
			args: args{mID: "id1", value: 9.99},
			want: func() Metric {
				var value = 9.99
				return Metric{
					ID:    "id1",
					Type:  MetricTypeGauge,
					Delta: nil,
					Value: &value,
				}
			},
		},
		{
			name: "negative value",
			args: args{mID: "id1", value: -9.99},
			want: func() Metric {
				var value = -9.99
				return Metric{
					ID:    "id1",
					Type:  MetricTypeGauge,
					Delta: nil,
					Value: &value,
				}
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want(), NewGaugeMetric(tt.args.mID, tt.args.value))
		})
	}
}

func TestMetric_Key(t *testing.T) {
	type fields struct {
		ID    string
		Type  MetricType
		Delta *int64
		Value *float64
		Hash  MetricHash
	}
	tests := []struct {
		name   string
		fields fields
		want   MetricKey
	}{
		{
			name:   "empty metric",
			fields: fields{},
			want:   MetricKey{},
		},
		{
			name:   "empty ID",
			fields: fields{Type: MetricTypeCounter},
			want:   MetricKey{Type: MetricTypeCounter},
		},
		{
			name:   "empty type",
			fields: fields{ID: "id1"},
			want:   MetricKey{ID: "id1"},
		},
		{
			name:   "normal metric",
			fields: fields{Type: MetricTypeGauge, ID: "id1"},
			want:   MetricKey{Type: MetricTypeGauge, ID: "id1"},
		},
		{
			name:   "custom type",
			fields: fields{Type: MetricType("custom"), ID: "id1"},
			want:   MetricKey{Type: MetricType("custom"), ID: "id1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				ID:    tt.fields.ID,
				Type:  tt.fields.Type,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
				Hash:  tt.fields.Hash,
			}
			assert.Equal(t, tt.want, m.Key())
		})
	}
}

func TestNewMetricKey(t *testing.T) {
	type args struct {
		mType MetricType
		mID   string
	}
	tests := []struct {
		name string
		args args
		want MetricKey
	}{
		{
			name: "empty key",
			args: args{},
			want: MetricKey{},
		},
		{
			name: "empty ID",
			args: args{mType: MetricTypeCounter},
			want: MetricKey{Type: MetricTypeCounter},
		},
		{
			name: "empty type",
			args: args{mID: "id1"},
			want: MetricKey{ID: "id1"},
		},
		{
			name: "normal key",
			args: args{mType: MetricTypeGauge, mID: "id1"},
			want: MetricKey{Type: MetricTypeGauge, ID: "id1"},
		},
		{
			name: "custom key",
			args: args{mType: MetricType("custom"), mID: "id1"},
			want: MetricKey{Type: MetricType("custom"), ID: "id1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetricKey(tt.args.mType, tt.args.mID))
		})
	}
}

func TestMetric_Empty(t *testing.T) {
	type fields struct {
		ID    string
		Type  MetricType
		Delta *int64
		Value *float64
		Hash  MetricHash
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// true.
		{
			name:   "all fields empty",
			fields: fields{},
			want:   true,
		},
		{
			name:   "empty id",
			fields: fields{Type: MetricTypeCounter},
			want:   true,
		},
		{
			name:   "empty type",
			fields: fields{ID: "id1"},
			want:   true,
		},
		{
			name: "empty counter",
			fields: func() fields {
				var value = 1.1
				return fields{ID: "id1", Type: MetricTypeCounter, Value: &value}
			}(),
			want: true,
		},
		{
			name: "empty gauge",
			fields: func() fields {
				var value int64 = 10
				return fields{ID: "id1", Type: MetricTypeGauge, Delta: &value}
			}(),
			want: true,
		},
		{
			name:   "empty custom type",
			fields: fields{ID: "id1", Type: MetricType("custom")},
			want:   true,
		},
		// false
		{
			name: "normal counter",
			fields: func() fields {
				var value int64 = 10
				return fields{ID: "id1", Type: MetricTypeCounter, Delta: &value}
			}(),
			want: false,
		},
		{
			name: "normal gauge",
			fields: func() fields {
				var value = 9.2
				return fields{ID: "id1", Type: MetricTypeGauge, Value: &value}
			}(),
			want: false,
		},
		{
			name: "normal custom type",
			fields: func() fields {
				var value = 9.2
				return fields{ID: "id1", Type: MetricType("custom"), Value: &value}
			}(),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				ID:    tt.fields.ID,
				Type:  tt.fields.Type,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
				Hash:  tt.fields.Hash,
			}
			assert.Equal(t, tt.want, m.Empty())
		})
	}
}

func TestMetric_Copy(t *testing.T) {
	type fields struct {
		ID    string
		Type  MetricType
		Delta *int64
		Value *float64
		Hash  MetricHash
	}
	tests := []struct {
		name      string
		fields    fields
		want      Metric
		assertion func(assert.TestingT, Metric, Metric)
	}{
		{
			name:   "empty",
			fields: fields{},
			want:   Metric{},
			assertion: func(t assert.TestingT, want, got Metric) {
				assert.Equal(t, want, got)
			},
		},
		{
			name: "no delta or value",
			fields: fields{
				ID:   "id1",
				Type: MetricType("custom"),
				Hash: "something",
			},
			want: Metric{
				ID:   "id1",
				Type: MetricType("custom"),
				Hash: "something",
			},
			assertion: func(t assert.TestingT, want, got Metric) {
				assert.Equal(t, want, got)
			},
		},
		{
			name: "counter",
			fields: func() fields {
				var v int64 = 10
				return fields{
					ID:    "id1",
					Type:  MetricTypeCounter,
					Delta: &v,
				}
			}(),
			want: NewCounterMetric("id1", 10),
			assertion: func(t assert.TestingT, want, got Metric) {
				assert.Equal(t, want, got)
				*got.Delta += 15
				assert.NotEqual(t, want, got)
				assert.Equal(t, int64(10), *want.Delta)
				assert.Equal(t, int64(25), *got.Delta)
			},
		},
		{
			name: "gauge",
			fields: func() fields {
				var v = 0.1
				return fields{
					ID:    "id1",
					Type:  MetricTypeGauge,
					Value: &v,
				}
			}(),
			want: NewGaugeMetric("id1", 0.1),
			assertion: func(t assert.TestingT, want, got Metric) {
				assert.Equal(t, want, got)
				*got.Value += 1.5
				assert.NotEqual(t, want, got)
				assert.Equal(t, 0.1, *want.Value)
				assert.Equal(t, 1.6, *got.Value)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				ID:    tt.fields.ID,
				Type:  tt.fields.Type,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
				Hash:  tt.fields.Hash,
			}
			tt.assertion(t, tt.want, m.Copy())
		})
	}
}

func TestMetricKey_Empty(t *testing.T) {
	type fields struct {
		Type MetricType
		ID   string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "no id, no type",
			fields: fields{},
			want:   true,
		},
		{
			name:   "no id",
			fields: fields{Type: MetricTypeCounter},
			want:   true,
		},
		{
			name:   "no type",
			fields: fields{ID: "id1"},
			want:   true,
		},
		{
			name:   "have id and type",
			fields: fields{ID: "id1", Type: MetricTypeCounter},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := MetricKey{
				Type: tt.fields.Type,
				ID:   tt.fields.ID,
			}
			assert.Equal(t, tt.want, k.Empty())
		})
	}
}

func TestMetricSet_UniqueByKey(t *testing.T) {
	tests := []struct {
		name string
		ms   MetricSet
		want map[MetricKey]Metric
	}{
		{
			name: "empty set",
			ms:   MetricSet{},
			want: map[MetricKey]Metric{},
		},
		{
			name: "single metric",
			ms:   MetricSet{NewCounterMetric("id1", 5)},
			want: map[MetricKey]Metric{
				NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
			},
		},
		{
			name: "multiple metrics",
			ms: MetricSet{
				NewCounterMetric("id1", 5),
				NewCounterMetric("id2", -10),
				NewGaugeMetric("id3", 2.4),
				NewGaugeMetric("id4", -1.3),
			},
			want: map[MetricKey]Metric{
				NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
				NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
				NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 2.4),
				NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
			},
		},
		{
			name: "multiple metrics with duplicate ids",
			ms: MetricSet{
				NewCounterMetric("id1", 5),
				NewCounterMetric("id2", -10),
				NewCounterMetric("id1", 33),
				NewGaugeMetric("id3", 2.4),
				NewGaugeMetric("id4", -1.3),
				NewGaugeMetric("id3", 0.85),
			},
			want: map[MetricKey]Metric{
				NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 33),
				NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
				NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 0.85),
				NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ms.UniqueByKey())
		})
	}
}
