package storagetest

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockStorage(t *testing.T) {
	type args struct {
		metrics []model.Metric
	}
	type want struct {
		isFaulty bool
		metrics  map[model.MetricKey]model.Metric
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "default",
			args: args{},
			want: want{metrics: map[model.MetricKey]model.Metric{}},
		},
		{
			name: "some metrics",
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewGaugeMetric("id2", -3.3),
				},
			},
			want: want{
				metrics: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
		},
		{
			name: "some metrics with duplicate ids",
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewGaugeMetric("id2", -3.3),
					model.NewCounterMetric("id1", -5),
					model.NewGaugeMetric("id2", 4),
					model.NewCounterMetric("id3", 15),
				},
			},
			want: want{
				metrics: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", -5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", 4),
					model.NewMetricKey(model.MetricTypeCounter, "id3"): model.NewCounterMetric("id3", 15),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMockStorage(tt.args.metrics...)
			assert.Equal(t, tt.want.isFaulty, s.isFaulty)
			assert.Equal(t, tt.want.metrics, s.data)
		})
	}
}

func TestMockStorage_Get(t *testing.T) {
	type fields struct {
		data      map[model.MetricKey]model.Metric
		isFaulty  bool
		triggerID string
	}
	type args struct {
		k model.MetricKey
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty storage",
			fields: fields{
				data: map[model.MetricKey]model.Metric{},
			},
			args: args{k: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, repository.ErrMetricNotFound)
			},
		},
		{
			name: "normal get",
			fields: fields{
				data: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{k: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: model.NewCounterMetric("id1", 5),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "faulty get",
			fields: fields{
				isFaulty:  true,
				triggerID: "zztop",
				data: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id1"):   model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeCounter, "zztop"): model.NewCounterMetric(FaultyStorageErrorTrigger, 17),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):     model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{k: model.NewMetricKey(model.MetricTypeCounter, "zztop")},
			want: model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrFaultyStorage)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			got, err := s.Get(t.Context(), tt.args.k)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMockStorage_GetAll(t *testing.T) {
	type fields struct {
		data      map[model.MetricKey]model.Metric
		isFaulty  bool
		triggerID string
	}
	type want struct {
		metrics []model.Metric
		err     error
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "empty storage",
			fields: fields{
				data: map[model.MetricKey]model.Metric{},
			},
			want: want{metrics: []model.Metric{}},
		},
		{
			name: "some metrics",
			fields: fields{
				data: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id3"): model.NewCounterMetric("id3", -8),
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			want: want{metrics: []model.Metric{
				model.NewCounterMetric("id1", 5),
				model.NewGaugeMetric("id2", -3.3),
				model.NewCounterMetric("id3", -8),
			}},
		},
		{
			name: "faulty metric",
			fields: fields{
				isFaulty:  true,
				triggerID: "tada",
				data: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id3"): model.NewCounterMetric("id3", -8),
					model.NewMetricKey(model.MetricTypeGauge, "tada"):  model.NewGaugeMetric("tada", 1.8),
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			want: want{metrics: []model.Metric{}, err: ErrFaultyStorage},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			got, err := s.GetAll(t.Context())
			if tt.want.err != nil {
				assert.ErrorIs(t, err, tt.want.err)
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want.metrics, got)
		})
	}
}

func TestMockStorage_Set(t *testing.T) {
	type fields struct {
		data      map[model.MetricKey]model.Metric
		isFaulty  bool
		triggerID string
	}
	type args struct {
		m model.Metric
	}
	type want struct {
		m   model.Metric
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "empty storage",
			fields: fields{
				data: map[model.MetricKey]model.Metric{},
			},
			args: args{m: model.NewCounterMetric("id1", 5)},
			want: want{m: model.NewCounterMetric("id1", 5)},
		},
		{
			name: "add new",
			fields: fields{
				data: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{m: model.NewCounterMetric("id3", -5)},
			want: want{m: model.NewCounterMetric("id3", -5)},
		},
		{
			name: "overwrite existing",
			fields: fields{
				data: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{m: model.NewCounterMetric("id1", -5)},
			want: want{m: model.NewCounterMetric("id1", -5)},
		},
		{
			name: "faulty metric",
			fields: fields{
				isFaulty:  true,
				triggerID: "oops",
				data: map[model.MetricKey]model.Metric{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{m: model.NewCounterMetric("oops", -5)},
			want: want{m: model.Metric{}, err: ErrFaultyStorage},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}

			err := s.Set(t.Context(), tt.args.m)

			if tt.want.err != nil {
				require.ErrorIs(t, err, tt.want.err)
				return
			}

			require.NoError(t, err)
			require.Contains(t, s.data, tt.want.m.Key())
			assert.Equal(t, tt.want.m, s.data[tt.want.m.Key()])
		})
	}
}

func TestMockStorage_MakeFaulty(t *testing.T) {
	type fields struct {
		data      map[model.MetricKey]model.Metric
		isFaulty  bool
		triggerID string
	}
	type want struct {
		isFaulty bool
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name:   "normal -> faulty",
			fields: fields{data: map[model.MetricKey]model.Metric{}, isFaulty: false, triggerID: "something"},
			want:   want{isFaulty: true},
		},
		{
			name:   "faulty -> faulty",
			fields: fields{data: map[model.MetricKey]model.Metric{}, isFaulty: true, triggerID: "something"},
			want:   want{isFaulty: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			assert.Equal(t, tt.want.isFaulty, s.MakeFaulty().isFaulty)
		})
	}
}

func TestMockStorage_MakeNormal(t *testing.T) {
	type fields struct {
		data      map[model.MetricKey]model.Metric
		isFaulty  bool
		triggerID string
	}
	type want struct {
		isFaulty bool
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name:   "normal -> normal",
			fields: fields{data: map[model.MetricKey]model.Metric{}, isFaulty: false, triggerID: "something"},
			want:   want{isFaulty: false},
		},
		{
			name:   "faulty -> normal",
			fields: fields{data: map[model.MetricKey]model.Metric{}, isFaulty: true, triggerID: "something"},
			want:   want{isFaulty: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			assert.Equal(t, tt.want.isFaulty, s.MakeNormal().isFaulty)
		})
	}
}
