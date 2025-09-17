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

func (s *faultyStorage) Get(h model.MetricHash) (model.Metric, error) {
	m, err := s.realStorage.Get(h)
	switch m.ID {
	case faultyStorageErrorTrigger:
		return model.Metric{}, errors.New("faulty storage get error")
	default:
		return m, err
	}
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
		name      string
		fields    fields
		args      args
		want      *want
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty storage",
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
			args:   args{metric: model.NewCounterMetric("id1", 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
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
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
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
			want: &want{metric: model.NewCounterMetric("id1", 15)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
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
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
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
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				storage: tt.fields.storage(),
			}
			tt.assertion(t, s.storeSingle(tt.args.metric))
			got, err := s.storage.Get(tt.args.metric.Hash)
			require.NoError(t, err)
			if tt.want != nil {
				assert.Equal(t, tt.want.metric, got)
			} else {
				assert.Equal(t, tt.args.metric, got)
			}
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
			tt.assertion(t, s.Store(tt.args.metrics[0], tt.args.metrics[1:]...))
			for _, m := range tt.want.metrics {
				got, err := s.storage.Get(m.Hash)
				assert.NoError(t, err)
				assert.Equal(t, m, got)
			}
		})
	}
}

func Test_metricService_Retrieve(t *testing.T) {
	type fields struct {
		storage func() repository.Storage
	}
	type args struct {
		hashes []model.MetricHash
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
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
			args: args{hashes: []model.MetricHash{"counter/id1"}},
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
			args: args{hashes: []model.MetricHash{"counter/id1", "gauge/id2"}},
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
			args: args{hashes: []model.MetricHash{"counter/id1", "gauge/id2"}},
			want: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 3.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				storage: tt.fields.storage(),
			}
			got, err := s.Retrieve(tt.args.hashes[0], tt.args.hashes[1:]...)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
