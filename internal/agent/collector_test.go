package agent

import (
	"errors"
	"maps"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/agent/source/extra"
	"github.com/bq2cd/yp-go-metrics/internal/agent/source/memstats"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCollector struct {
	mock.Mock
	metrics []model.Metric
	wantErr bool
}

func (m *mockCollector) Collect() error {
	m.Called()
	if m.wantErr {
		return errors.New("collect error")
	}
	return nil
}

func (m *mockCollector) Snapshot() ([]model.Metric, error) {
	m.Called()
	if m.wantErr {
		return nil, errors.New("snapshot error")
	}
	return m.metrics, nil
}

type faultyStorage struct{}

func (s *faultyStorage) Get(key model.MetricKey) (model.Metric, error) {
	return model.Metric{}, errors.New("faulty storage get error")
}

func (s *faultyStorage) Set(metric model.Metric) error {
	return errors.New("faulty storage set error")
}

func (s *faultyStorage) GetAll() ([]model.Metric, error) {
	return nil, errors.New("faulty storage getAll error")
}

func Test_collector_Collect(t *testing.T) {
	type args struct {
		sources []source.Source
		storage repository.Storage
	}
	type want struct {
		metricIDToType map[string]model.MetricType
	}
	tests := []struct {
		name      string
		args      args
		want      want
		assertion func(assert.TestingT, want, []model.Metric)
	}{
		{
			name: "default metrics",
			args: args{sources: []source.Source{memstats.New(), extra.New()}, storage: repository.NewMemStorage()},
			want: want{
				metricIDToType: func() map[string]model.MetricType {
					m := make(map[string]model.MetricType)
					maps.Copy(m, memstats.GetSupportedMetrics())
					maps.Copy(m, extra.GetSupportedMetrics())
					return m
				}(),
			},
			assertion: func(t assert.TestingT, want want, got []model.Metric) {
				assert.Len(t, got, len(want.metricIDToType))
				metricIDToType := make(map[string]model.MetricType, len(want.metricIDToType))
				for i := range got {
					m := got[i]
					metricIDToType[m.ID] = m.Type
				}
				for mID, mType := range want.metricIDToType {
					assert.Contains(t, metricIDToType, mID)
					assert.Equal(t, mType, metricIDToType[mID])
				}
			},
		},
		{
			name: "extra metrics",
			args: args{
				sources: []source.Source{extra.New()},
				storage: func() repository.Storage {
					s := repository.NewMemStorage()
					err := s.Set(model.NewCounterMetric("PollCount", 10))
					assert.NoError(t, err)
					return s
				}(),
			},
			want: want{
				metricIDToType: func() map[string]model.MetricType {
					m := make(map[string]model.MetricType)
					maps.Copy(m, extra.GetSupportedMetrics())
					return m
				}(),
			},
			assertion: func(t assert.TestingT, want want, got []model.Metric) {
				assert.Len(t, got, len(want.metricIDToType))
				metricIDToType := make(map[string]model.MetricType, len(want.metricIDToType))
				metricsByKey := make(map[model.MetricKey]model.Metric)
				for i := range got {
					m := got[i]
					metricIDToType[m.ID] = m.Type
					metricsByKey[m.Key()] = m
				}
				for mID, mType := range want.metricIDToType {
					assert.Contains(t, metricIDToType, mID)
					assert.Equal(t, mType, metricIDToType[mID])
				}
				test := model.NewCounterMetric("PollCount", 1)
				assert.Equal(t, test, metricsByKey[test.Key()])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &collector{sources: tt.args.sources, collected: tt.args.storage}
			err := c.Collect()
			assert.NoError(t, err)
			got, err := tt.args.storage.GetAll()
			assert.NoError(t, err)
			tt.assertion(t, tt.want, got)
		})
	}
}

func TestNewCollector(t *testing.T) {
	type args struct {
		sources []source.Source
		storage repository.Storage
	}
	tests := []struct {
		name string
		args args
		want *collector
	}{
		{
			name: "empty",
			args: args{},
			want: &collector{},
		},
		{
			name: "empty sources",
			args: args{storage: repository.NewMemStorage()},
			want: &collector{collected: repository.NewMemStorage()},
		},
		{
			name: "some sources",
			args: args{sources: []source.Source{extra.New()}, storage: repository.NewMemStorage()},
			want: &collector{sources: []source.Source{extra.New()}, collected: repository.NewMemStorage()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewCollector(tt.args.sources, tt.args.storage))
		})
	}
}

func Test_collector_storeMetrics(t *testing.T) {
	type fields struct {
		sources   []source.Source
		collected repository.Storage
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
		assertion func(assert.TestingT, repository.Storage, want, error)
	}{
		{
			name: "no metrics",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
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
			name: "single metric",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
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
			name: "multiple metrics",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
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
			name: "multiple metrics 2",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
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
			name: "multiple metrics 3",
			fields: fields{
				collected: func() repository.Storage {
					s := repository.NewMemStorage()
					for _, m := range []model.Metric{
						model.NewCounterMetric("id5", 7),
					} {
						err := s.Set(m)
						assert.NoError(t, err)
					}
					return s
				}(),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5), model.NewCounterMetric("id5", 7)}},
			assertion: func(t assert.TestingT, s repository.Storage, want want, err error) {
				assert.NoError(t, err)
				metrics, err := s.GetAll()
				assert.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple counters with the same id",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
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
			name: "multiple counters with the same id 2",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
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
			name: "multiple counters with the same id 3",
			fields: fields{
				collected: func() repository.Storage {
					s := repository.NewMemStorage()
					for _, m := range []model.Metric{
						model.NewCounterMetric("id1", 7),
					} {
						err := s.Set(m)
						assert.NoError(t, err)
					}
					return s
				}(),
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
			name: "multiple gauges with the same id",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
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
			name: "multiple gauges with the same id 2",
			fields: fields{
				collected: func() repository.Storage {
					s := repository.NewMemStorage()
					for _, m := range []model.Metric{
						model.NewGaugeMetric("id1", 7.7),
					} {
						err := s.Set(m)
						assert.NoError(t, err)
					}
					return s
				}(),
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
			name: "storage error",
			fields: fields{
				collected: func() repository.Storage {
					return &faultyStorage{}
				}(),
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
			c := &collector{
				sources:   tt.fields.sources,
				collected: tt.fields.collected,
			}
			tt.assertion(t, tt.fields.collected, tt.want, c.storeMetrics(tt.args.metrics))
		})
	}
}

func Test_collector_Snapshot(t *testing.T) {
	type fields struct {
		sources   []source.Source
		collected repository.Storage
	}
	tests := []struct {
		name      string
		fields    fields
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty storage",
			fields: fields{
				collected: func() repository.Storage {
					return repository.NewMemStorage()
				}(),
			},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "single metric",
			fields: fields{
				collected: func() repository.Storage {
					s := repository.NewMemStorage()
					for _, m := range []model.Metric{
						model.NewCounterMetric("id1", 5),
					} {
						err := s.Set(m)
						assert.NoError(t, err)
					}
					return s
				}(),
			},
			want: []model.Metric{model.NewCounterMetric("id1", 5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{
				collected: func() repository.Storage {
					s := repository.NewMemStorage()
					for _, m := range []model.Metric{
						model.NewCounterMetric("id1", 5),
						model.NewGaugeMetric("id2", -4.1),
					} {
						err := s.Set(m)
						assert.NoError(t, err)
					}
					return s
				}(),
			},
			want: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -4.1)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "retrieval error",
			fields: fields{
				collected: func() repository.Storage {
					return &faultyStorage{}
				}(),
			},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &collector{
				sources:   tt.fields.sources,
				collected: tt.fields.collected,
			}
			got, err := c.Snapshot()
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
