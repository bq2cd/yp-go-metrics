package service

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMetricStorer struct {
	mock.Mock
	metrics []model.Metric
	err     error
}

func (m *mockMetricStorer) StoreSingle(metric model.Metric) error {
	m.Called(metric)
	return m.err
}
func (m *mockMetricStorer) StoreBatch(metrics []model.Metric) error {
	m.Called(metrics)
	return m.err
}
func (m *mockMetricStorer) RetrieveSingle(key model.MetricKey) (model.Metric, error) {
	m.Called(key)
	return model.Metric{}, m.err
}
func (m *mockMetricStorer) RetrieveBatch(keys []model.MetricKey) ([]model.Metric, error) {
	m.Called(keys)
	return m.metrics, m.err
}
func (m *mockMetricStorer) RetrieveAll() ([]model.Metric, error) {
	m.Called()
	return m.metrics, m.err
}

func TestNewMetricStorer(t *testing.T) {
	type args struct {
		storage repository.Storage
	}
	tests := []struct {
		name string
		args args
		want *metricStorer
	}{
		{
			name: "new service",
			args: args{storage: storagetest.NewMockStorage()},
			want: &metricStorer{storage: storagetest.NewMockStorage()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetricStorer(tt.args.storage))
		})
	}
}

func Test_metricStorer_StoreSingle(t *testing.T) {
	type fields struct {
		storage repository.Storage
	}
	type args struct {
		metric model.Metric
	}
	type want struct {
		metric model.Metric
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		want               want
		assertion          assert.ErrorAssertionFunc
		assertStoredMetric func(assert.TestingT, repository.Storage, want)
	}{
		{
			name:   "empty storage",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{metric: model.NewCounterMetric("id1", 10)},
			want:   want{metric: model.NewCounterMetric("id1", 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "empty counter",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{metric: model.Metric{ID: "id1", Type: model.MetricTypeCounter}},
			want:   want{metric: model.Metric{}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.ErrorIs(t, err, repository.ErrMetricNotFound)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "empty counter 2",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:   args{metric: model.Metric{ID: "id1", Type: model.MetricTypeCounter}},
			want:   want{metric: model.NewCounterMetric("id1", 5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "empty counter in storage",
			fields: fields{storage: storagetest.NewMockStorage(model.Metric{ID: "id1", Type: model.MetricTypeCounter})},
			args:   args{metric: model.NewCounterMetric("id2", 10)},
			want:   want{metric: model.NewCounterMetric("id2", 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "new counter",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:   args{metric: model.NewCounterMetric("id2", 10)},
			want:   want{metric: model.NewCounterMetric("id2", 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "existing counter",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:   args{metric: model.NewCounterMetric("id1", 10)},
			want:   want{metric: model.NewCounterMetric("id1", 15)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "new gauge",
			fields: fields{storage: storagetest.NewMockStorage(model.NewGaugeMetric("id1", 1.5))},
			args:   args{metric: model.NewGaugeMetric("id2", 5.1)},
			want:   want{metric: model.NewGaugeMetric("id2", 5.1)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "existing gauge",
			fields: fields{storage: storagetest.NewMockStorage(model.NewGaugeMetric("id1", 1.5))},
			args:   args{metric: model.NewGaugeMetric("id1", -1.5)},
			want:   want{metric: model.NewGaugeMetric("id1", -1.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:   "faulty storage",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 10)).MakeFaulty()},
			args:   args{metric: model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 5)},
			want:   want{metric: model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, storagetest.ErrFaultyStorage)
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.(*storagetest.MockStorage).Get(want.metric.Key())
				assert.ErrorIs(t, err, storagetest.ErrFaultyStorage)
				assert.Equal(t, model.Metric{}, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricStorer{
				storage: tt.fields.storage,
			}
			tt.assertion(t, s.StoreSingle(tt.args.metric))
			tt.assertStoredMetric(t, s.storage, tt.want)
		})
	}
}

func Test_metricStorer_StoreBatch(t *testing.T) {
	type fields struct {
		storage repository.Storage
	}
	type args struct {
		metrics []model.Metric
	}
	type want struct {
		metrics []model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty counter metric",
			fields: fields{storage: storagetest.NewMockStorage()},
			args: args{metrics: []model.Metric{
				func() model.Metric {
					return model.Metric{ID: "id1", Type: model.MetricTypeCounter}
				}(),
			}},
			want: want{metrics: []model.Metric{}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "empty gauge metric",
			fields: fields{storage: storagetest.NewMockStorage()},
			args: args{metrics: []model.Metric{
				func() model.Metric {
					var value int64 = 5
					return model.Metric{ID: "id1", Type: model.MetricTypeGauge, Delta: &value}
				}(),
			}},
			want: want{metrics: []model.Metric{}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "single metric",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple metrics",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple metrics, some empty",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5), {}, {ID: "id5", Type: model.MetricTypeCounter}, {ID: "id6", Type: model.MetricTypeGauge}}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple counters with same id",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 10), model.NewCounterMetric("id1", -5)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 10)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "faulty storage",
			fields: fields{storage: storagetest.NewMockStorage().MakeFaulty()},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5), model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 0)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, storagetest.ErrFaultyStorage)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricStorer{
				storage: tt.fields.storage,
			}
			metrics := make([]model.Metric, 0, len(tt.args.metrics))
			for _, m := range tt.args.metrics {
				metrics = append(metrics, m.Copy())
			}
			tt.assertion(t, s.StoreBatch(metrics))
			for _, m := range tt.want.metrics {
				got, err := s.storage.Get(m.Key())
				assert.NoError(t, err)
				assert.Equal(t, m, got)
			}
			assert.Equal(t, tt.args.metrics, metrics)
		})
	}
}

func Test_metricStorer_RetrieveSingle(t *testing.T) {
	type fields struct {
		storage repository.Storage
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
			name:   "missing metric",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				assert.Error(t, err)
				return assert.ErrorIs(t, err, repository.ErrMetricNotFound)
			},
		},
		{
			name:   "single metric",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:   args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want:   model.NewCounterMetric("id1", 5),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple counters with same id",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 15))},
			args:   args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want:   model.NewCounterMetric("id1", 15),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "empty metric in storage",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5), model.Metric{ID: "id2", Type: model.MetricTypeCounter})},
			args:   args{key: model.NewMetricKey(model.MetricTypeCounter, "id2")},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, repository.ErrMetricNotFound)
			},
		},
		{
			name:   "faulty storage",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5), model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 0)).MakeFaulty()},
			args:   args{key: model.NewMetricKey(model.MetricTypeCounter, storagetest.FaultyStorageErrorTrigger)},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, storagetest.ErrFaultyStorage)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricStorer{
				storage: tt.fields.storage,
			}
			got, err := s.RetrieveSingle(tt.args.key)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_metricStorer_RetrieveBatch(t *testing.T) {
	type fields struct {
		storage repository.Storage
	}
	type args struct {
		keys []model.MetricKey
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "missing metric",
			fields: fields{storage: storagetest.NewMockStorage()},
			args:   args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1")}},
			want:   []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "single metric",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:   args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1")}},
			want:   []model.Metric{model.NewCounterMetric("id1", 5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple metrics",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5))},
			args:   args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1"), model.NewMetricKey(model.MetricTypeGauge, "id2")}},
			want:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "faulty storage",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5), model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 0)).MakeFaulty()},
			args:   args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1"), model.NewMetricKey(model.MetricTypeGauge, "id2"), model.NewMetricKey(model.MetricTypeCounter, storagetest.FaultyStorageErrorTrigger)}},
			want:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, storagetest.ErrFaultyStorage)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricStorer{
				storage: tt.fields.storage,
			}
			got, err := s.RetrieveBatch(tt.args.keys)
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_metricStorer_RetrieveAll(t *testing.T) {
	type fields struct {
		storage repository.Storage
	}
	tests := []struct {
		name      string
		fields    fields
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty storage",
			fields: fields{storage: storagetest.NewMockStorage()},
			want:   []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "faulty storage",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5), model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, -2)).MakeFaulty()},
			want:   []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, storagetest.ErrFaultyStorage)
			},
		},
		{
			name:   "multiple metrics",
			fields: fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5))},
			want:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricStorer{
				storage: tt.fields.storage,
			}
			got, err := s.RetrieveAll()
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
