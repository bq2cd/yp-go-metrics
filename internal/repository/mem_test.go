package repository

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *memStorage
	}{
		{
			name: "empty storage",
			want: &memStorage{data: make(memStorageData)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMemStorage())
		})
	}
}

func Test_memStorage_Get(t *testing.T) {
	type fields struct {
		data memStorageData
	}
	type args struct {
		key model.MetricKey
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty storage",
			fields: fields{data: memStorageData{}},
			args:   args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMetricNotFound)
			},
		},
		{
			name: "metric exists",
			fields: fields{data: memStorageData{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): func() model.Metric {
					var value int64 = 9
					return model.Metric{
						ID:    "id1",
						Type:  model.MetricTypeCounter,
						Delta: &value,
						Hash:  "counter1",
					}
				}(),
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				}}},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: func() model.Metric {
				var value int64 = 9
				return model.Metric{
					ID:    "id1",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter1",
				}
			}(),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "empty metric requested",
			fields: fields{data: memStorageData{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter1",
				},
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				}}},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMetricNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			got, err := s.Get(tt.args.key)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_memStorage_Set(t *testing.T) {
	type fields struct {
		data memStorageData
	}
	type args struct {
		metric func() model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
		contains  func(assert.TestingT, memStorageData, model.Metric)
	}{
		{
			name:   "empty storage",
			fields: fields{data: make(memStorageData)},
			args: args{metric: func() model.Metric {
				var value int64 = 10
				return model.Metric{
					ID:    "id1",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter1",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data memStorageData, metric model.Metric) {
				assert.Contains(t, data, metric.Key())
				assert.Equal(t, metric, data[metric.Key()])
			},
		},
		{
			name:   "empty metric not added",
			fields: fields{data: make(memStorageData)},
			args: args{metric: func() model.Metric {
				var value int64 = 5
				return model.Metric{
					ID:    "id3",
					Type:  model.MetricTypeGauge,
					Delta: &value,
					Hash:  "gauge3",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data memStorageData, metric model.Metric) {
				assert.NotContains(t, data, metric.Key())
			},
		},
		{
			name: "metric added",
			fields: fields{data: memStorageData{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter1",
				},
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				},
			}},
			args: args{metric: func() model.Metric {
				var value int64 = 5
				return model.Metric{
					ID:    "id3",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter3",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data memStorageData, metric model.Metric) {
				assert.Contains(t, data, metric.Key())
				assert.Equal(t, metric, data[metric.Key()])
			},
		},
		{
			name: "metric replaced",
			fields: fields{data: memStorageData{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter1",
				},
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				},
			}},
			args: args{metric: func() model.Metric {
				var value int64 = 15
				return model.Metric{
					ID:    "id1",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter1",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data memStorageData, metric model.Metric) {
				assert.Contains(t, data, metric.Key())
				assert.Equal(t, metric, data[metric.Key()])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			metric := tt.args.metric()
			tt.assertion(t, s.Set(metric))
			tt.contains(t, tt.fields.data, metric)
		})
	}
}

func Test_memStorage_GetAll(t *testing.T) {
	type fields struct {
		data memStorageData
	}
	tests := []struct {
		name      string
		fields    fields
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty storage",
			fields: fields{data: make(memStorageData)},
			want:   []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{data: memStorageData{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 10),
				model.NewMetricKey(model.MetricTypeGauge, "id1"):   model.NewGaugeMetric("id2", 1.5),
			}},
			want: []model.Metric{model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id2", 1.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "storage with empty metrics",
			fields: fields{data: memStorageData{
				model.NewMetricKey(model.MetricTypeCounter, "id1"):   model.NewCounterMetric("id1", 10),
				model.NewMetricKey(model.MetricTypeGauge, "id1"):     model.NewGaugeMetric("id2", 1.5),
				model.NewMetricKey(model.MetricTypeCounter, ""):      model.Metric{Type: model.MetricTypeCounter},
				model.NewMetricKey("", ""):                           model.Metric{},
				model.NewMetricKey(model.MetricTypeGauge, "emptyMe"): model.Metric{ID: "emptyMe", Type: model.MetricTypeGauge},
			}},
			want: []model.Metric{model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id2", 1.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			got, err := s.GetAll()
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
