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
		hash model.MetricHash
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
			args:   args{hash: model.MetricHash("counter/id1")},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Equal(t, ErrMetricNotFound, err)
			},
		},
		{
			name: "metric exists",
			fields: fields{data: memStorageData{
				"counter/id1": model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter/id1",
				},
				"gauge/id1": model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge/id1",
				}}},
			args: args{hash: model.MetricHash("counter/id1")},
			want: model.Metric{
				ID:   "id1",
				Type: model.MetricTypeCounter,
				Hash: "counter/id1",
			},
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
			got, err := s.Get(tt.args.hash)
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
					Hash:  "counter/id1",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "metric added",
			fields: fields{data: memStorageData{
				"counter/id1": model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter/id1",
				},
				"gauge/id1": model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge/id1",
				},
			}},
			args: args{metric: func() model.Metric {
				var value int64 = 5
				return model.Metric{
					ID:    "id3",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter/id3",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "metric replaced",
			fields: fields{data: memStorageData{
				"counter/id1": model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter/id1",
				},
				"gauge/id1": model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge/id1",
				},
			}},
			args: args{metric: func() model.Metric {
				var value int64 = 15
				return model.Metric{
					ID:    "id1",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter/id1",
				}
			}},
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
			metric := tt.args.metric()
			tt.assertion(t, s.Set(metric))
			assert.Contains(t, tt.fields.data, metric.Hash)
			assert.Equal(t, metric, tt.fields.data[metric.Hash])
		})
	}
}
