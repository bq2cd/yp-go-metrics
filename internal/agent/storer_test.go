package agent

import (
	"errors"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStorer struct {
	mock.Mock
	metrics []model.Metric
}

func (m *mockStorer) Store(metrics []model.Metric) error {
	m.Called(metrics)
	m.metrics = metrics
	return nil
}

func (m *mockStorer) Retrieve() ([]model.Metric, error) {
	m.Called()
	return m.metrics, nil
}

type faultyStorer struct{}

func (s *faultyStorer) Store(metrics []model.Metric) error {
	return errors.New("faulty storer store error")
}

func (s *faultyStorer) Retrieve() ([]model.Metric, error) {
	return nil, errors.New("faulty storer retrieve error")
}

func Test_defaultStorer_Store(t *testing.T) {
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
		init      func(repository.Storage) Storer
		args      args
		want      want
		assertion func(assert.TestingT, repository.Storage, want, error)
	}{
		{
			name:   "no metrics",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			args: args{metrics: []model.Metric{}},
			want: want{metrics: []model.Metric{}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name:   "single metric",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name:   "multiple metrics",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -2.5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -2.5)}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name:   "multiple metrics 2",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5)}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name:   "multiple counters with the same id",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id1", 8.3)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id1", 8.3)}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name:   "multiple counters with the same id 2",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id1", 8.3), model.NewCounterMetric("id1", -5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", -5), model.NewGaugeMetric("id1", 8.3)}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name:   "multiple gauges with the same id",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			args: args{metrics: []model.Metric{model.NewGaugeMetric("id1", 0.5), model.NewGaugeMetric("id1", -0.5), model.NewCounterMetric("id1", -3)}},
			want: want{metrics: []model.Metric{model.NewGaugeMetric("id1", -0.5), model.NewCounterMetric("id1", -3)}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name:   "storage error",
			fields: fields{storage: repository.NewMemStorage()},
			init: func(s repository.Storage) Storer {
				return &faultyStorer{}
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -2.5)}},
			want: want{metrics: []model.Metric{}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.init(tt.fields.storage)
			tt.assertion(t, tt.fields.storage, tt.want, s.Store(tt.args.metrics))
		})
	}
}

func Test_defaultStorer_Retrieve(t *testing.T) {
	type fields struct {
		storage repository.Storage
	}
	tests := []struct {
		name      string
		fields    fields
		init      func(repository.Storage) Storer
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty storage",
			fields: fields{storage: func() repository.Storage {
				s := repository.NewMemStorage()
				return s
			}()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
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
			}()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			want: []model.Metric{model.NewCounterMetric("id1", 5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{storage: func() repository.Storage {
				s := repository.NewMemStorage()
				for _, m := range []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -4.1)} {
					err := s.Set(m)
					assert.NoError(t, err)
				}
				return s
			}()},
			init: func(s repository.Storage) Storer {
				return &defaultStorer{storage: s}
			},
			want: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -4.1)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "retrieval error",
			fields: fields{storage: func() repository.Storage {
				s := repository.NewMemStorage()
				for _, m := range []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -4.1)} {
					err := s.Set(m)
					assert.NoError(t, err)
				}
				return s
			}()},
			init: func(s repository.Storage) Storer {
				return &faultyStorer{}
			},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.init(tt.fields.storage)
			got, err := s.Retrieve()
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestNewDefaultStorer(t *testing.T) {
	type args struct {
		storage repository.Storage
	}
	tests := []struct {
		name      string
		args      args
		assertion func(assert.TestingT, args, *defaultStorer)
	}{
		{
			name: "default initialisation",
			args: args{storage: repository.NewMemStorage()},
			assertion: func(t assert.TestingT, want args, got *defaultStorer) {
				assert.Equal(t, want.storage, got.storage)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewDefaultStorer(tt.args.storage))
		})
	}
}
