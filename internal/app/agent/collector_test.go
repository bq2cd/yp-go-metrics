package agent

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/extra"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockCollector struct {
	mock.Mock
	metrics []model.Metric
	wantErr bool
}

func (m *mockCollector) Collect(ctx context.Context) error {
	m.Called(ctx)
	if m.wantErr {
		return errors.New("collect error")
	}
	return nil
}

func (m *mockCollector) Snapshot(ctx context.Context) ([]model.Metric, error) {
	m.Called(ctx)
	if m.wantErr {
		return nil, errors.New("snapshot error")
	}
	return m.metrics, nil
}

type faultyStorage struct{}

func (s *faultyStorage) Get(ctx context.Context, key model.MetricKey) (model.Metric, error) {
	return model.Metric{}, errors.New("faulty storage get error")
}

func (s *faultyStorage) Set(ctx context.Context, metric model.Metric) error {
	return errors.New("faulty storage set error")
}

func (s *faultyStorage) GetAll(ctx context.Context) ([]model.Metric, error) {
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
			args: args{sources: []source.Source{memstats.New(), extra.New()}, storage: storagetest.NewMockStorage()},
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
				storage: storagetest.NewMockStorage(model.NewCounterMetric("PollCount", 10)),
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
			err := c.Collect(t.Context())
			require.NoError(t, err)
			got, err := tt.args.storage.GetAll(t.Context())
			require.NoError(t, err)
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
			args: args{storage: storagetest.NewMockStorage()},
			want: &collector{collected: storagetest.NewMockStorage()},
		},
		{
			name: "some sources",
			args: args{sources: []source.Source{extra.New()}, storage: storagetest.NewMockStorage()},
			want: &collector{sources: []source.Source{extra.New()}, collected: storagetest.NewMockStorage()},
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
		assertion func(*testing.T, repository.Storage, want, error)
	}{
		{
			name: "no metrics",
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{metrics: []model.Metric{}},
			want: want{metrics: []model.Metric{}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "single metric",
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -2.5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -2.5)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple metrics 2",
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple metrics 3",
			fields: fields{
				collected: storagetest.NewMockStorage(model.NewCounterMetric("id5", 7)),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id1", -2.5), model.NewCounterMetric("id5", 7)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple counters with the same id",
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id1", 8.3)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id1", 8.3)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple counters with the same id 2",
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id1", 8.3), model.NewCounterMetric("id1", -5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", -5), model.NewGaugeMetric("id1", 8.3)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple counters with the same id 3",
			fields: fields{
				collected: storagetest.NewMockStorage(model.NewCounterMetric("id1", 7)),
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id1", 8.3), model.NewCounterMetric("id1", -5)}},
			want: want{metrics: []model.Metric{model.NewCounterMetric("id1", -5), model.NewGaugeMetric("id1", 8.3)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple gauges with the same id",
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{metrics: []model.Metric{model.NewGaugeMetric("id1", 0.5), model.NewGaugeMetric("id1", -0.5), model.NewCounterMetric("id1", -3)}},
			want: want{metrics: []model.Metric{model.NewGaugeMetric("id1", -0.5), model.NewCounterMetric("id1", -3)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "multiple gauges with the same id 2",
			fields: fields{
				collected: storagetest.NewMockStorage(model.NewGaugeMetric("id1", 7.7)),
			},
			args: args{metrics: []model.Metric{model.NewGaugeMetric("id1", 0.5), model.NewGaugeMetric("id1", -0.5), model.NewCounterMetric("id1", -3)}},
			want: want{metrics: []model.Metric{model.NewGaugeMetric("id1", -0.5), model.NewCounterMetric("id1", -3)}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.NoError(t, err)
				metrics, err := s.GetAll(t.Context())
				require.NoError(t, err)
				assert.ElementsMatch(t, want.metrics, metrics)
			},
		},
		{
			name: "storage error",
			fields: fields{
				collected: &faultyStorage{},
			},
			args: args{metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -2.5)}},
			want: want{metrics: []model.Metric{}},
			assertion: func(t *testing.T, s repository.Storage, want want, err error) {
				require.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &collector{
				sources:   tt.fields.sources,
				collected: tt.fields.collected,
			}
			tt.assertion(t, tt.fields.collected, tt.want, c.storeMetrics(t.Context(), tt.args.metrics))
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
				collected: storagetest.NewMockStorage(),
			},
			want: []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "single metric",
			fields: fields{
				collected: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5)),
			},
			want: []model.Metric{model.NewCounterMetric("id1", 5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{
				collected: storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 5),
					model.NewGaugeMetric("id2", -4.1),
				),
			},
			want: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", -4.1)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "retrieval error",
			fields: fields{
				collected: &faultyStorage{},
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
			got, err := c.Snapshot(t.Context())
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
