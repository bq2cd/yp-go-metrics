package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMetricStorer struct {
	mock.Mock
	metrics []model.Metric
	err     error
}

func (m *mockMetricStorer) StoreSingle(ctx context.Context, metric model.Metric) error {
	m.Called(ctx, metric)
	return m.err
}
func (m *mockMetricStorer) StoreBatch(ctx context.Context, metrics []model.Metric) error {
	m.Called(ctx, metrics)
	return m.err
}
func (m *mockMetricStorer) RetrieveSingle(ctx context.Context, key model.MetricKey) (model.Metric, error) {
	m.Called(ctx, key)
	return model.Metric{}, m.err
}
func (m *mockMetricStorer) RetrieveBatch(ctx context.Context, keys []model.MetricKey) ([]model.Metric, error) {
	m.Called(ctx, keys)
	return m.metrics, m.err
}
func (m *mockMetricStorer) RetrieveAll(ctx context.Context) ([]model.Metric, error) {
	m.Called(ctx)
	return m.metrics, m.err
}

func TestNewMetricStorer(t *testing.T) {
	type args struct {
		reader repository.StorageMultiReader
		writer StorageBatchWriter
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "with mock storage",
			args: args{
				reader: storagetest.NewMockStorage(),
				writer: NewStorageBatchWriter(storagetest.NewMockStorage()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetricStorer(tt.args.reader, tt.args.writer)
			assert.Equal(t, tt.args.reader, got.reader)
			assert.Equal(t, tt.args.writer, got.writer)
		})
	}
}

func Test_metricStorer_StoreSingle(t *testing.T) {
	type fields struct {
		storage repository.StorageMulti
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
		assertStoredMetric func(*testing.T, repository.StorageMulti, want)
	}{
		{
			name:      "empty storage",
			fields:    fields{storage: storagetest.NewMockStorage()},
			args:      args{metric: model.NewCounterMetric("id1", 10)},
			want:      want{metric: model.NewCounterMetric("id1", 10)},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:      "empty storage, empty counter ignored",
			fields:    fields{storage: storagetest.NewMockStorage()},
			args:      args{metric: model.Metric{ID: "id1", Type: model.MetricTypeCounter}},
			want:      want{metric: model.Metric{}},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.ErrorIs(t, err, repository.ErrMetricNotFound)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:      "non-empty storage, empty counter ignored",
			fields:    fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:      args{metric: model.Metric{ID: "id1", Type: model.MetricTypeCounter}},
			want:      want{metric: model.NewCounterMetric("id1", 5)},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:      "empty counter in storage",
			fields:    fields{storage: storagetest.NewMockStorage(model.Metric{ID: "id1", Type: model.MetricTypeCounter})},
			args:      args{metric: model.NewCounterMetric("id2", 10)},
			want:      want{metric: model.NewCounterMetric("id2", 10)},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:      "new counter",
			fields:    fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:      args{metric: model.NewCounterMetric("id2", 10)},
			want:      want{metric: model.NewCounterMetric("id2", 10)},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:      "existing counter",
			fields:    fields{storage: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:      args{metric: model.NewCounterMetric("id1", 10)},
			want:      want{metric: model.NewCounterMetric("id1", 15)},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:      "new gauge",
			fields:    fields{storage: storagetest.NewMockStorage(model.NewGaugeMetric("id1", 1.5))},
			args:      args{metric: model.NewGaugeMetric("id2", 5.1)},
			want:      want{metric: model.NewGaugeMetric("id2", 5.1)},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.NoError(t, err)
				assert.Equal(t, want.metric, got)
			},
		},
		{
			name:      "existing gauge",
			fields:    fields{storage: storagetest.NewMockStorage(model.NewGaugeMetric("id1", 1.5))},
			args:      args{metric: model.NewGaugeMetric("id1", -1.5)},
			want:      want{metric: model.NewGaugeMetric("id1", -1.5)},
			assertion: assert.NoError,
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.Get(t.Context(), want.metric.Key())
				require.NoError(t, err)
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
			assertStoredMetric: func(t *testing.T, s repository.StorageMulti, want want) {
				got, err := s.(*storagetest.MockStorage).Get(t.Context(), want.metric.Key())
				require.ErrorIs(t, err, storagetest.ErrFaultyStorage)
				assert.Equal(t, model.Metric{}, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewStorageBatchWriter(tt.fields.storage)
			go w.StartProcessing(t.Context())
			s := &metricStorer{
				reader: tt.fields.storage,
				writer: w,
			}
			tt.assertion(t, s.StoreSingle(t.Context(), tt.args.metric))
			tt.assertStoredMetric(t, tt.fields.storage, tt.want)
		})
	}
}

func Test_metricStorer_StoreBatch(t *testing.T) {
	type fields struct {
		storage repository.StorageMulti
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
			w := NewStorageBatchWriter(tt.fields.storage)
			go w.StartProcessing(t.Context())
			s := &metricStorer{
				reader: tt.fields.storage,
				writer: w,
			}
			metrics := make([]model.Metric, 0, len(tt.args.metrics))
			for _, m := range tt.args.metrics {
				metrics = append(metrics, m.Copy())
			}
			tt.assertion(t, s.StoreBatch(t.Context(), metrics))
			for _, m := range tt.want.metrics {
				got, err := tt.fields.storage.Get(t.Context(), m.Key())
				require.NoError(t, err)
				assert.Equal(t, m, got)
			}
			assert.Equal(t, tt.args.metrics, metrics)
		})
	}
}

func Test_metricStorer_RetrieveSingle(t *testing.T) {
	type fields struct {
		storage repository.StorageMulti
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
				return assert.ErrorIs(t, err, ErrMetricNotFound)
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
				return assert.ErrorIs(t, err, ErrMetricNotFound)
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
			w := NewStorageBatchWriter(tt.fields.storage)
			go w.StartProcessing(t.Context())
			s := &metricStorer{
				reader: tt.fields.storage,
				writer: w,
			}
			got, err := s.RetrieveSingle(t.Context(), tt.args.key)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_metricStorer_RetrieveBatch(t *testing.T) {
	type fields struct {
		storage repository.StorageMulti
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
			w := NewStorageBatchWriter(tt.fields.storage)
			go w.StartProcessing(t.Context())
			s := &metricStorer{
				reader: tt.fields.storage,
				writer: w,
			}
			got, err := s.RetrieveBatch(t.Context(), tt.args.keys)
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_metricStorer_RetrieveAll(t *testing.T) {
	type fields struct {
		storage repository.StorageMulti
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
			w := NewStorageBatchWriter(tt.fields.storage)
			go w.StartProcessing(t.Context())
			s := &metricStorer{
				reader: tt.fields.storage,
				writer: w,
			}
			got, err := s.RetrieveAll(t.Context())
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_metricStorer_StoreBatch_Concurrent(t *testing.T) {
	type batch []model.Metric
	type args struct {
		batches []batch
	}
	type want struct {
		metrics model.MetricSet
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]struct {
		storage        func() repository.StorageMulti
		storagePerCase bool
		cases          map[string]testcase
	}{
		"initially empty storage, isolated per case": {
			storage:        func() repository.StorageMulti { return storagetest.NewMockStorage() },
			storagePerCase: true,
			cases: map[string]testcase{
				"single gauge in different batches with same id but different values": {
					args: args{batches: []batch{
						{model.NewGaugeMetric("id1", 1.3)},
						{model.NewGaugeMetric("id1", -0.3)},
						{model.NewGaugeMetric("id1", 4.6)},
						{model.NewGaugeMetric("id1", -5.01)},
						{model.NewGaugeMetric("id1", 99.1)},
						{model.NewGaugeMetric("id1", 0.01)},
						{model.NewGaugeMetric("id1", 17)},
						{model.NewGaugeMetric("id1", -323.11)},
						{model.NewGaugeMetric("id1", 3.1415)},
						{model.NewGaugeMetric("id1", -0.0025)},
						{model.NewGaugeMetric("id1", 1024.1)},
					}},
					want: want{
						metrics: model.NewMetricSet(
							model.NewGaugeMetric("id1", 1024.1),
						),
						wantErr: func(t testing.TB, err error) {
							require.NoError(t, err)
						},
					},
				},
				"single counter in different batches with same id but different values": {
					args: args{batches: []batch{
						{model.NewCounterMetric("id1", 13)},
						{model.NewCounterMetric("id1", 3)},
						{model.NewCounterMetric("id1", -46)},
						{model.NewCounterMetric("id1", 5)},
						{model.NewCounterMetric("id1", 9)},
						{model.NewCounterMetric("id1", -7)},
						{model.NewCounterMetric("id1", 17)},
						{model.NewCounterMetric("id1", -11)},
						{model.NewCounterMetric("id1", 55)},
						{model.NewCounterMetric("id1", -40)},
						{model.NewCounterMetric("id1", 8)},
					}},
					want: want{
						metrics: model.NewMetricSet(
							model.NewCounterMetric("id1", 6),
						),
						wantErr: func(t testing.TB, err error) {
							require.NoError(t, err)
						},
					},
				},
			},
		},
		"initially empty storage, shared between cases": {
			storage:        func() repository.StorageMulti { return storagetest.NewMockStorage() },
			storagePerCase: false,
			cases: map[string]testcase{
				"single gauge in different batches with same id but different values": {
					args: args{batches: []batch{
						{model.NewGaugeMetric("id1", 1.3)},
						{model.NewGaugeMetric("id1", -0.3)},
						{model.NewGaugeMetric("id1", 4.6)},
						{model.NewGaugeMetric("id1", -5.01)},
						{model.NewGaugeMetric("id1", 99.1)},
						{model.NewGaugeMetric("id1", 0.01)},
						{model.NewGaugeMetric("id1", 17)},
						{model.NewGaugeMetric("id1", -323.11)},
						{model.NewGaugeMetric("id1", 3.1415)},
						{model.NewGaugeMetric("id1", -0.0025)},
						{model.NewGaugeMetric("id1", 1024.1)},
					}},
					want: want{
						metrics: model.NewMetricSet(
							model.NewGaugeMetric("id1", 1024.1),
							model.NewCounterMetric("id1", 6),
						),
						wantErr: func(t testing.TB, err error) {
							require.NoError(t, err)
						},
					},
				},
				"single counter in different batches with same id but different values": {
					args: args{batches: []batch{
						{model.NewCounterMetric("id1", 13)},
						{model.NewCounterMetric("id1", 3)},
						{model.NewCounterMetric("id1", -46)},
						{model.NewCounterMetric("id1", 5)},
						{model.NewCounterMetric("id1", 9)},
						{model.NewCounterMetric("id1", -7)},
						{model.NewCounterMetric("id1", 17)},
						{model.NewCounterMetric("id1", -11)},
						{model.NewCounterMetric("id1", 55)},
						{model.NewCounterMetric("id1", -40)},
						{model.NewCounterMetric("id1", 8)},
					}},
					want: want{
						metrics: model.NewMetricSet(
							model.NewGaugeMetric("id1", 1024.1),
							model.NewCounterMetric("id1", 6),
						),
						wantErr: func(t testing.TB, err error) {
							require.NoError(t, err)
						},
					},
				},
			},
		},
	}
	var (
		storageMu      sync.RWMutex
		storageFactory = make(map[string]repository.StorageMulti, len(tests))
	)
	for gname, group := range tests {
		t.Run(gname, func(t *testing.T) {
			t.Parallel()
			storageMu.Lock()
			storageFactory[gname] = group.storage()
			storageMu.Unlock()
			for name, tt := range group.cases {
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					// Arrange
					ctx := t.Context()
					var storage repository.StorageMulti
					if group.storagePerCase {
						storage = group.storage()
					} else {
						storageMu.RLock()
						storage = storageFactory[gname]
						storageMu.RUnlock()
					}
					w := NewStorageBatchWriter(storage)
					go w.StartProcessing(t.Context())
					s := &metricStorer{
						reader: storage,
						writer: w,
					}

					// Act
					var wg sync.WaitGroup
					errCh := make(chan error, len(tt.args.batches))
					sendCh := make(chan batch)
					for range tt.args.batches {
						wg.Add(1)
						go func() {
							defer wg.Done()
							batch := <-sendCh
							errCh <- s.StoreBatch(ctx, batch)
						}()
					}
					go func() {
						wg.Wait()
						close(errCh)
					}()
					for i := range tt.args.batches {
						sendCh <- tt.args.batches[i]
						// it seems fair to assume that batches would not be arriving faster than every 2 milliseconds,
						// especially in networked environments.
						time.Sleep(2 * time.Millisecond)
					}
					close(sendCh)
					var errFinal error
					for err := range errCh {
						errFinal = errors.Join(errFinal, err)
					}

					// Assert
					tt.want.wantErr(t, errFinal)
					stored, err := storage.GetAll(ctx)
					require.NoError(t, err)
					assert.Equal(t, tt.want.metrics, model.NewMetricSet(stored...))
				})

			}
		})
	}
}
