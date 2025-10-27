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
				assert.InEpsilon(t, 0.1, *want.Value, 1e-10)
				assert.InEpsilon(t, 1.6, *got.Value, 1e-10)
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

func TestNewMetricSetWithStrategy(t *testing.T) {
	type testcase struct {
		metrics []Metric
		want    MetricSet
	}
	tests := map[string]struct {
		strategy MetricAggregateStrategy
		cases    map[string]testcase
	}{
		"last value wins": {
			strategy: MetricAggregateStrategyLastValueWins,
			cases: map[string]testcase{
				"empty list": {
					metrics: []Metric{},
					want:    MetricSet{},
				},
				"single metric": {
					metrics: []Metric{NewCounterMetric("id1", 5)},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
					},
				},
				"multiple metrics without duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 2.4),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
				"multiple metrics with duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewCounterMetric("id1", 33),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
						NewGaugeMetric("id3", 0.85),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 33),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 0.85),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
				"empty metrics are skipped": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						{},
						NewCounterMetric("id2", -10),
						{Type: MetricTypeGauge},
						NewCounterMetric("id1", 33),
						{Type: MetricTypeGauge, ID: "empty1"},
						NewGaugeMetric("id3", 2.4),
						func() Metric {
							var delta int64 = 3
							return Metric{Type: MetricTypeGauge, ID: "empty2", Delta: &delta}
						}(),
						NewGaugeMetric("id4", -1.3),
						func() Metric {
							var value = 3.1
							return Metric{Type: MetricTypeCounter, ID: "empty3", Value: &value}
						}(),
						NewGaugeMetric("id3", 0.85),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 33),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 0.85),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
			},
		},
		"first value wins": {
			strategy: MetricAggregateStrategyFirstValueWins,
			cases: map[string]testcase{
				"empty list": {
					metrics: []Metric{},
					want:    MetricSet{},
				},
				"single metric": {
					metrics: []Metric{NewCounterMetric("id1", 5)},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
					},
				},
				"multiple metrics without duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 2.4),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
				"multiple metrics with duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewCounterMetric("id1", 33),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
						NewGaugeMetric("id3", 0.85),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 2.4),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
				"empty metrics are skipped": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						{},
						NewCounterMetric("id2", -10),
						{Type: MetricTypeGauge},
						NewCounterMetric("id1", 33),
						{Type: MetricTypeGauge, ID: "empty1"},
						NewGaugeMetric("id3", 2.4),
						func() Metric {
							var delta int64 = 3
							return Metric{Type: MetricTypeGauge, ID: "empty2", Delta: &delta}
						}(),
						NewGaugeMetric("id4", -1.3),
						func() Metric {
							var value = 3.1
							return Metric{Type: MetricTypeCounter, ID: "empty3", Value: &value}
						}(),
						NewGaugeMetric("id3", 0.85),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 2.4),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
			},
		},
		"counter value accumulates": {
			strategy: MetricAggregateStrategyCounterValueAccumulates,
			cases: map[string]testcase{
				"empty list": {
					metrics: []Metric{},
					want:    MetricSet{},
				},
				"single metric": {
					metrics: []Metric{NewCounterMetric("id1", 5)},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
					},
				},
				"multiple metrics without duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 2.4),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
				"multiple metrics with duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewCounterMetric("id1", 33),
						NewCounterMetric("id2", 7),
						NewCounterMetric("id1", 2),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
						NewGaugeMetric("id3", 0.85),
						NewGaugeMetric("id4", -4.5),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 40),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -3),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 0.85),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -4.5),
					},
				},
				"empty metrics are skipped": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						{},
						NewCounterMetric("id2", -10),
						{Type: MetricTypeGauge},
						NewCounterMetric("id1", 33),
						{Type: MetricTypeGauge, ID: "empty1"},
						NewGaugeMetric("id3", 2.4),
						func() Metric {
							var delta int64 = 3
							return Metric{Type: MetricTypeGauge, ID: "empty2", Delta: &delta}
						}(),
						NewGaugeMetric("id4", -1.3),
						func() Metric {
							var value = 3.1
							return Metric{Type: MetricTypeCounter, ID: "empty3", Value: &value}
						}(),
						NewGaugeMetric("id3", 0.85),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 38),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 0.85),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
			},
		},
		"unknown strategy behaves like last value wins": {
			strategy: MetricAggregateStrategy(-1),
			cases: map[string]testcase{
				"empty list": {
					metrics: []Metric{},
					want:    MetricSet{},
				},
				"single metric": {
					metrics: []Metric{NewCounterMetric("id1", 5)},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
					},
				},
				"multiple metrics without duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 5),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 2.4),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
				"multiple metrics with duplicates": {
					metrics: []Metric{
						NewCounterMetric("id1", 5),
						NewCounterMetric("id2", -10),
						NewCounterMetric("id1", 33),
						NewGaugeMetric("id3", 2.4),
						NewGaugeMetric("id4", -1.3),
						NewGaugeMetric("id3", 0.85),
					},
					want: MetricSet{
						NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 33),
						NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", -10),
						NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", 0.85),
						NewMetricKey(MetricTypeGauge, "id4"):   NewGaugeMetric("id4", -1.3),
					},
				},
			},
		},
	}
	for gname, group := range tests {
		t.Run(gname, func(t *testing.T) {
			for tname, tc := range group.cases {
				t.Run(tname, func(t *testing.T) {
					orig := tc.metrics
					got := NewMetricSetWithStrategy(group.strategy, tc.metrics...)
					assert.Equal(t, tc.want, got)
					assert.Equal(t, orig, tc.metrics)
				})
			}
		})
	}
}

func TestMetric_AddDelta(t *testing.T) {
	getDelta := func(v int64) *int64 { return &v }
	type fields struct {
		metric Metric
	}
	type args struct {
		other *int64
	}
	type want struct {
		metric Metric
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"nil delta changes nothing": {
			fields: fields{metric: NewCounterMetric("id1", 15)},
			args:   args{other: nil},
			want:   want{metric: NewCounterMetric("id1", 15)},
		},
		"some delta changes nothing for gauge": {
			fields: fields{metric: NewGaugeMetric("id1", -3.4)},
			args:   args{other: getDelta(8)},
			want:   want{metric: NewGaugeMetric("id1", -3.4)},
		},
		"some delta added to previous value for counter": {
			fields: fields{metric: NewCounterMetric("id1", -15)},
			args:   args{other: getDelta(8)},
			want:   want{metric: NewCounterMetric("id1", -7)},
		},
		"some delta replaces empty value for counter": {
			fields: fields{metric: Metric{Type: MetricTypeCounter, ID: "id1"}},
			args:   args{other: getDelta(8)},
			want:   want{metric: NewCounterMetric("id1", 8)},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := tt.fields.metric
			m.AddDelta(tt.args.other)
			assert.Equal(t, tt.want.metric, m)
		})
	}
}

func TestNewMetricKeySet(t *testing.T) {
	type args struct {
		keys []MetricKey
	}
	type want struct {
		got MetricKeySet
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty list -> empty set": {
			args: args{keys: []MetricKey{}},
			want: want{got: MetricKeySet{}},
		},
		"single key": {
			args: args{keys: []MetricKey{NewMetricKey(MetricTypeCounter, "id1")}},
			want: want{got: MetricKeySet{NewMetricKey(MetricTypeCounter, "id1"): struct{}{}}},
		},
		"multiple keys without duplicates": {
			args: args{keys: []MetricKey{
				NewMetricKey(MetricTypeCounter, "id1"),
				NewMetricKey(MetricTypeCounter, "id2"),
				NewMetricKey(MetricTypeGauge, "id3"),
				NewMetricKey(MetricTypeGauge, "id4"),
			}},
			want: want{got: MetricKeySet{
				NewMetricKey(MetricTypeCounter, "id1"): struct{}{},
				NewMetricKey(MetricTypeCounter, "id2"): struct{}{},
				NewMetricKey(MetricTypeGauge, "id3"):   struct{}{},
				NewMetricKey(MetricTypeGauge, "id4"):   struct{}{},
			}},
		},
		"multiple keys with duplicates": {
			args: args{keys: []MetricKey{
				NewMetricKey(MetricTypeCounter, "id1"),
				NewMetricKey(MetricTypeCounter, "id2"),
				NewMetricKey(MetricTypeGauge, "id3"),
				NewMetricKey(MetricTypeGauge, "id4"),
				NewMetricKey(MetricTypeCounter, "id1"),
				NewMetricKey(MetricTypeCounter, "id2"),
			}},
			want: want{got: MetricKeySet{
				NewMetricKey(MetricTypeCounter, "id1"): struct{}{},
				NewMetricKey(MetricTypeCounter, "id2"): struct{}{},
				NewMetricKey(MetricTypeGauge, "id3"):   struct{}{},
				NewMetricKey(MetricTypeGauge, "id4"):   struct{}{},
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewMetricKeySet(tt.args.keys...)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestNewMetricSet(t *testing.T) {
	// covered by [TestNewMetricSetWithStrategy]
	t.SkipNow()
}

func TestMetricSet_GroupByType(t *testing.T) {
	type want struct {
		got map[MetricType]MetricSet
	}
	type testcase struct {
		ms   MetricSet
		want want
	}
	tests := map[string]testcase{
		"empty set": {
			ms:   MetricSet{},
			want: want{got: map[MetricType]MetricSet{}},
		},
		"single metric": {
			ms: MetricSet{
				NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 15),
			},
			want: want{got: map[MetricType]MetricSet{
				MetricTypeCounter: {
					NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 15),
				},
			}},
		},
		"multiple metrics with single type": {
			ms: MetricSet{
				NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 15),
				NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", 25),
				NewMetricKey(MetricTypeCounter, "id3"): NewCounterMetric("id3", -5),
			},
			want: want{got: map[MetricType]MetricSet{
				MetricTypeCounter: {
					NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 15),
					NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", 25),
					NewMetricKey(MetricTypeCounter, "id3"): NewCounterMetric("id3", -5),
				},
			}},
		},
		"multiple metrics with various types": {
			ms: MetricSet{
				NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 15),
				NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", 25),
				NewMetricKey(MetricTypeCounter, "id3"): NewCounterMetric("id3", -5),
				NewMetricKey(MetricTypeGauge, "id1"):   NewGaugeMetric("id1", 1.5),
				NewMetricKey(MetricTypeGauge, "id2"):   NewGaugeMetric("id2", 2.5),
				NewMetricKey(MetricTypeGauge, "id3"):   NewGaugeMetric("id3", -0.5),
				NewMetricKey(MetricTypeCounter, "id7"): NewCounterMetric("id7", 150),
				NewMetricKey(MetricTypeCounter, "id8"): NewCounterMetric("id8", 250),
				NewMetricKey(MetricTypeCounter, "id9"): NewCounterMetric("id9", -50),
				NewMetricKey(MetricTypeGauge, "id10"):  NewGaugeMetric("id10", 1.05),
				NewMetricKey(MetricTypeGauge, "id20"):  NewGaugeMetric("id20", 2.05),
				NewMetricKey(MetricTypeGauge, "id30"):  NewGaugeMetric("id30", -0.05),
			},
			want: want{got: map[MetricType]MetricSet{
				MetricTypeCounter: {
					NewMetricKey(MetricTypeCounter, "id1"): NewCounterMetric("id1", 15),
					NewMetricKey(MetricTypeCounter, "id2"): NewCounterMetric("id2", 25),
					NewMetricKey(MetricTypeCounter, "id3"): NewCounterMetric("id3", -5),
					NewMetricKey(MetricTypeCounter, "id7"): NewCounterMetric("id7", 150),
					NewMetricKey(MetricTypeCounter, "id8"): NewCounterMetric("id8", 250),
					NewMetricKey(MetricTypeCounter, "id9"): NewCounterMetric("id9", -50),
				},
				MetricTypeGauge: {
					NewMetricKey(MetricTypeGauge, "id1"):  NewGaugeMetric("id1", 1.5),
					NewMetricKey(MetricTypeGauge, "id2"):  NewGaugeMetric("id2", 2.5),
					NewMetricKey(MetricTypeGauge, "id3"):  NewGaugeMetric("id3", -0.5),
					NewMetricKey(MetricTypeGauge, "id10"): NewGaugeMetric("id10", 1.05),
					NewMetricKey(MetricTypeGauge, "id20"): NewGaugeMetric("id20", 2.05),
					NewMetricKey(MetricTypeGauge, "id30"): NewGaugeMetric("id30", -0.05),
				},
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.ms.GroupByType()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestMetricKey_Compare(t *testing.T) {
	type fields struct {
		Type MetricType
		ID   string
	}
	type args struct {
		other MetricKey
	}
	type want struct {
		got int
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"same type, same id -> 0": {
			fields: fields{Type: MetricTypeCounter, ID: "id1"},
			args:   args{other: NewMetricKey(MetricTypeCounter, "id1")},
			want:   want{got: 0},
		},
		"same type, id1 < id2 -> -1": {
			fields: fields{Type: MetricTypeCounter, ID: "id1"},
			args:   args{other: NewMetricKey(MetricTypeCounter, "id2")},
			want:   want{got: -1},
		},
		"same type, id3 > id2 -> +1": {
			fields: fields{Type: MetricTypeCounter, ID: "id3"},
			args:   args{other: NewMetricKey(MetricTypeCounter, "id2")},
			want:   want{got: +1},
		},
		"counter < gauge -> -1": {
			fields: fields{Type: MetricTypeCounter, ID: "id3"},
			args:   args{other: NewMetricKey(MetricTypeGauge, "id2")},
			want:   want{got: -1},
		},
		"ztype > gauge -> +1": {
			fields: fields{Type: MetricType("ztype"), ID: "id3"},
			args:   args{other: NewMetricKey(MetricTypeGauge, "id2")},
			want:   want{got: +1},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			k := MetricKey{
				Type: tt.fields.Type,
				ID:   tt.fields.ID,
			}
			got := k.Compare(tt.args.other)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestMetricSet_Keys(t *testing.T) {
	type want struct {
		got MetricKeySet
	}
	type testcase struct {
		ms   MetricSet
		want want
	}
	tests := map[string]testcase{
		"empty set -> empty keys": {
			ms:   NewMetricSet(),
			want: want{got: NewMetricKeySet()},
		},
		"single metric -> single key": {
			ms:   NewMetricSet(NewCounterMetric("id1", 5)),
			want: want{got: NewMetricKeySet(NewMetricKey(MetricTypeCounter, "id1"))},
		},
		"multiple metrics -> multiple keys": {
			ms: NewMetricSet(
				NewCounterMetric("id1", 5),
				NewGaugeMetric("id3", 5.5),
				NewCounterMetric("id5", -5),
				NewGaugeMetric("id7", -5.5),
			),
			want: want{got: NewMetricKeySet(
				NewMetricKey(MetricTypeCounter, "id1"),
				NewMetricKey(MetricTypeGauge, "id3"),
				NewMetricKey(MetricTypeCounter, "id5"),
				NewMetricKey(MetricTypeGauge, "id7"),
			)},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.ms.Keys()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestMetricSet_Empty(t *testing.T) {
	type want struct {
		got bool
	}
	type testcase struct {
		ms   MetricSet
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.ms.Empty()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestMetricSet_Upsert(t *testing.T) {
	type args struct {
		m Metric
	}
	type want struct {
	}
	type testcase struct {
		ms   MetricSet
		args args
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.ms.Upsert(tt.args.m)
		})
	}
}

func TestMetricSet_Merge(t *testing.T) {
	type args struct {
		other MetricSet
	}
	type want struct {
	}
	type testcase struct {
		ms   MetricSet
		args args
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.ms.Merge(tt.args.other)
		})
	}
}

func TestMetricSet_Values(t *testing.T) {
	type want struct {
		got []Metric
	}
	type testcase struct {
		ms   MetricSet
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.ms.Values()
			assert.Equal(t, tt.want.got, got)
		})
	}
}
