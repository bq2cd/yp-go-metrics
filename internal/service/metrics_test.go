package service

import (
	"errors"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	faultyStorageErrorTrigger = "faultyStorageErrorTrigger"
)

type faultyStorage struct {
	realStorage repository.Storage
}

func (s *faultyStorage) Get(k model.MetricKey) (model.Metric, error) {
	m, err := s.realStorage.Get(k)
	switch m.ID {
	case faultyStorageErrorTrigger:
		return model.Metric{}, errors.New("faulty storage get error")
	default:
		return m, err
	}
}

func (s *faultyStorage) GetAll() ([]model.Metric, error) {
	return nil, errors.New("faulty storage internal error")
}

func (s *faultyStorage) Set(m model.Metric) error {
	switch m.ID {
	case faultyStorageErrorTrigger:
		return errors.New("faulty storage set error")
	default:
		return s.realStorage.Set(m)
	}
}

func TestNewMetrics(t *testing.T) {
	type args struct {
		storage repository.Storage
	}
	tests := []struct {
		name string
		args args
		want *metricService
	}{
		{
			name: "new service",
			args: args{storage: repository.NewMemStorage()},
			want: &metricService{storage: repository.NewMemStorage()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetrics(tt.args.storage))
		})
	}
}

func Test_metricService_storeSingle(t *testing.T) {
	type fields struct {
		storage func() repository.Storage
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
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
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
			name: "new counter",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewCounterMetric("id1", 5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewCounterMetric("id2", 10)},
			want: want{metric: model.NewCounterMetric("id2", 10)},
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
			name: "existing counter",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewCounterMetric("id1", 5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewCounterMetric("id1", 10)},
			want: want{metric: model.NewCounterMetric("id1", 15)},
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
			name: "new gauge",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewGaugeMetric("id1", 1.5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewGaugeMetric("id2", 5.1)},
			want: want{metric: model.NewGaugeMetric("id2", 5.1)},
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
			name: "existing gauge",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewGaugeMetric("id1", 1.5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewGaugeMetric("id1", -1.5)},
			want: want{metric: model.NewGaugeMetric("id1", -1.5)},
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
			name: "faulty storage",
			fields: fields{storage: func() repository.Storage {
				s := &faultyStorage{realStorage: repository.NewMemStorage()}
				err := s.realStorage.Set(model.NewCounterMetric(faultyStorageErrorTrigger, 10))
				require.NoError(t, err)
				return s
			}},
			args: args{metric: model.NewCounterMetric(faultyStorageErrorTrigger, 5)},
			want: want{metric: model.NewCounterMetric(faultyStorageErrorTrigger, 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Errorf(t, err, "faulty storage get error")
			},
			assertStoredMetric: func(t assert.TestingT, s repository.Storage, want want) {
				got, err := s.(*faultyStorage).realStorage.Get(want.metric.Key())
				assert.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				storage: tt.fields.storage(),
			}
			tt.assertion(t, s.storeSingle(tt.args.metric))
			tt.assertStoredMetric(t, s.storage, tt.want)
		})
	}
}

func Test_metricService_Store(t *testing.T) {
	type fields struct {
		storage func() repository.Storage
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
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
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
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
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
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple metrics",
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple metrics, some empty",
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5), {}, {ID: "id5", Type: model.MetricTypeCounter}, {ID: "id6", Type: model.MetricTypeGauge}}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "multiple counters with same id",
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 10), model.NewCounterMetric("id1", -5)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 10)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "faulty storage",
			fields: fields{storage: func() repository.Storage { return &faultyStorage{repository.NewMemStorage()} }},
			args:   args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5), model.NewCounterMetric(faultyStorageErrorTrigger, 0)}},
			want:   want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				storage: tt.fields.storage(),
			}
			metrics := make([]model.Metric, 0, len(tt.args.metrics))
			for _, m := range tt.args.metrics {
				metrics = append(metrics, m.Copy())
			}
			tt.assertion(t, s.Store(metrics[0], metrics[1:]...))
			for _, m := range tt.want.metrics {
				got, err := s.storage.Get(m.Key())
				assert.NoError(t, err)
				assert.Equal(t, m, got)
			}
			assert.Equal(t, tt.args.metrics, metrics)
		})
	}
}

func Test_metricService_Retrieve(t *testing.T) {
	type fields struct {
		storage func() repository.Storage
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
			name: "missing metric",
			fields: fields{storage: func() repository.Storage {
				s := repository.NewMemStorage()
				return s
			}},
			args: args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1")}},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				assert.Error(t, err)
				return assert.ErrorIs(t, err, repository.ErrMetricNotFound)
			},
		},
		{
			name: "single metric",
			fields: fields{storage: func() repository.Storage {
				s := repository.NewMemStorage()
				for _, m := range []model.Metric{model.NewCounterMetric("id1", 5)} {
					err := s.Set(m)
					assert.NoError(t, err)
				}
				return s
			}},
			args: args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1")}},
			want: []model.Metric{model.NewCounterMetric("id1", 5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{storage: func() repository.Storage {
				s := repository.NewMemStorage()
				for _, m := range []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)} {
					err := s.Set(m)
					assert.NoError(t, err)
				}
				return s
			}},
			args: args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1"), model.NewMetricKey(model.MetricTypeGauge, "id2")}},
			want: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "faulty storage",
			fields: fields{storage: func() repository.Storage {
				s := &faultyStorage{realStorage: repository.NewMemStorage()}
				for _, m := range []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5), model.NewCounterMetric(faultyStorageErrorTrigger, 0)} {
					err := s.realStorage.Set(m)
					assert.NoError(t, err)
				}
				return s
			}},
			args: args{keys: []model.MetricKey{model.NewMetricKey(model.MetricTypeCounter, "id1"), model.NewMetricKey(model.MetricTypeGauge, "id2"), model.NewMetricKey(model.MetricTypeCounter, faultyStorageErrorTrigger)}},
			want: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Errorf(t, err, "faulty storage get error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				storage: tt.fields.storage(),
			}
			got, err := s.Retrieve(tt.args.keys[0], tt.args.keys[1:]...)
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_metricService_RetrieveAll(t *testing.T) {
	type fields struct {
		storage repository.Storage
	}
	tests := []struct {
		name      string
		fields    fields
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{
			name: "empty storage",
			fields: fields{
				storage: func() repository.Storage {
					s := repository.NewMemStorage()
					return s
				}(),
			},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "faulty storage",
			fields: fields{
				storage: func() repository.Storage {
					s := &faultyStorage{realStorage: repository.NewMemStorage()}
					return s
				}(),
			},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{
				storage: func() repository.Storage {
					s := repository.NewMemStorage()
					for _, m := range []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)} {
						err := s.Set(m)
						assert.NoError(t, err)
					}
					return s
				}(),
			},
			want: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				storage: tt.fields.storage,
			}
			got, err := s.RetrieveAll()
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
