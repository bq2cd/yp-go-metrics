package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/sourcetest"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/testutil"
)

type mockCollector struct {
	testutil.Mock

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

func (m *mockCollector) Snapshot(ctx context.Context) (<-chan model.Metric, error) {
	m.Called(ctx)
	outCh := make(chan model.Metric)
	if m.wantErr {
		close(outCh)
		return outCh, errors.New("snapshot error")
	}
	go func() {
		defer close(outCh)
		for _, metric := range m.metrics {
			select {
			case <-ctx.Done():
				return
			case outCh <- metric:
			}
		}
	}()
	return outCh, nil
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

func TestNewCollector(t *testing.T) {
	type args struct {
		sources []source.Source
		storage repository.Storage
	}
	type want struct {
		checkFn func(testing.TB, args, *collector)
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"args are properly assigned": {
			args: args{
				sources: source.DefaultSources(),
				storage: storagetest.NewMockStorage(),
			},
			want: want{
				checkFn: func(t testing.TB, want args, got *collector) {
					assert.Equal(t, want.sources, got.sources)
					assert.Equal(t, want.storage, got.collected)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewCollector(tt.args.sources, tt.args.storage)
			tt.want.checkFn(t, tt.args, got)
		})
	}
}

func mockSourceWithMetrics(ctrl *gomock.Controller, metrics ...model.Metric) *sourcetest.MockSource {
	m := sourcetest.NewMockSource(ctrl)
	m.EXPECT().ReadMetrics().Return(metrics, nil)
	return m
}

func mockSourceWithError(ctrl *gomock.Controller, err error) *sourcetest.MockSource {
	m := sourcetest.NewMockSource(ctrl)
	m.EXPECT().ReadMetrics().Return(nil, err)
	return m
}

func mockSourceWithDelay(ctrl *gomock.Controller, delay time.Duration, metrics ...model.Metric) *sourcetest.MockSource {
	m := sourcetest.NewMockSource(ctrl)
	m.EXPECT().ReadMetrics().DoAndReturn(func() ([]model.Metric, error) {
		time.Sleep(delay)
		return metrics, nil
	})
	return m
}

func Test_collector_Collect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		sources   []source.Source
		collected repository.Storage
	}
	type args struct {
		timeout time.Duration
	}
	type want struct {
		wantErr func(testing.TB, error)
		metrics model.MetricSet
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"empty storage, single source successfully produces some metrics": {
			fields: fields{
				sources: []source.Source{
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id1", 5),
						model.NewGaugeMetric("id2", -3.3),
						model.NewCounterMetric("id1", -7),
					),
				},
				collected: storagetest.NewMockStorage(),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", -7),
					model.NewGaugeMetric("id2", -3.3),
				),
			},
		},
		"empty storage, multiple sources successfully produce some metrics": {
			fields: fields{
				sources: []source.Source{
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id1", 5),
						model.NewGaugeMetric("id2", -3.3),
						model.NewCounterMetric("id1", -7),
					),
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id10", 15),
						model.NewGaugeMetric("id20", -33.3),
						model.NewCounterMetric("id10", -70),
					),
				},
				collected: storagetest.NewMockStorage(),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", -7),
					model.NewCounterMetric("id10", -70),
					model.NewGaugeMetric("id2", -3.3),
					model.NewGaugeMetric("id20", -33.3),
				),
			},
		},
		"pre-populated storage, multiple sources successfully produce some metrics": {
			fields: fields{
				sources: []source.Source{
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id1", 5),
						model.NewGaugeMetric("id2", -3.3),
						model.NewCounterMetric("id1", -7),
					),
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id10", 15),
						model.NewGaugeMetric("id20", -33.3),
						model.NewCounterMetric("id10", -70),
					),
				},
				collected: storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 444),
					model.NewCounterMetric("id10", 888),
					model.NewCounterMetric("id001", 11),
					model.NewGaugeMetric("id2", 4.44),
					model.NewGaugeMetric("id20", 8.88),
					model.NewGaugeMetric("id002", -2.22),
				),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", -7),
					model.NewCounterMetric("id10", -70),
					model.NewCounterMetric("id001", 11),
					model.NewGaugeMetric("id2", -3.3),
					model.NewGaugeMetric("id20", -33.3),
					model.NewGaugeMetric("id002", -2.22),
				),
			},
		},
		"empty storage, one source fails to produce metrics": {
			fields: fields{
				sources: []source.Source{
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id1", 5),
						model.NewGaugeMetric("id2", -3.3),
						model.NewCounterMetric("id1", -7),
					),
					mockSourceWithError(ctrl, errors.New("source two failed")),
				},
				collected: storagetest.NewMockStorage(),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "source two failed")
				},
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", -7),
					model.NewGaugeMetric("id2", -3.3),
				),
			},
		},
		"empty storage, multiple sources successfully produce some metrics, storage fails": {
			fields: fields{
				sources: []source.Source{
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id1", 5),
						model.NewGaugeMetric("id2", -3.3),
						model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, -7),
					),
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id10", 15),
						model.NewGaugeMetric(storagetest.FaultyStorageErrorTrigger, -33.3),
						model.NewCounterMetric("id10", -70),
					),
				},
				collected: storagetest.NewMockStorage().MakeFaulty(),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, storagetest.ErrFaultyStorage)
				},
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id10", -70),
					model.NewGaugeMetric("id2", -3.3),
				),
			},
		},
		"empty storage, multiple sources, context cancelled while producing metrics": {
			fields: fields{
				sources: []source.Source{
					mockSourceWithMetrics(ctrl,
						model.NewCounterMetric("id1", 5),
						model.NewGaugeMetric("id2", -3.3),
						model.NewCounterMetric("id1", -7),
					),
					mockSourceWithDelay(ctrl, 500*time.Millisecond,
						model.NewCounterMetric("id10", 15),
						model.NewGaugeMetric("id20", -33.3),
						model.NewCounterMetric("id10", -70),
					),
				},
				collected: storagetest.NewMockStorage(),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "context deadline exceeded")
				},
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", -7),
					model.NewGaugeMetric("id2", -3.3),
				),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			ctx, cancel := context.WithTimeout(t.Context(), tt.args.timeout)
			defer cancel()
			c := &collector{
				sources:   tt.fields.sources,
				collected: tt.fields.collected,
			}

			// Act
			err := c.Collect(ctx)

			// Assert
			tt.want.wantErr(t, err)
			storedMetrics, err := tt.fields.collected.GetAll(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want.metrics, model.NewMetricSet(storedMetrics...))
		})
	}
}

func Test_collector_Snapshot(t *testing.T) {
	type fields struct {
		collected repository.Storage
	}
	type args struct {
		timeout time.Duration
	}
	type want struct {
		wantErr func(testing.TB, error)
		metrics model.MetricSet
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"empty storage produces no metrics": {
			fields: fields{
				collected: storagetest.NewMockStorage(),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
				metrics: model.NewMetricSet(),
			},
		},
		"non-empty storage produces some metrics": {
			fields: fields{
				collected: storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id10", 1.55),
					model.NewGaugeMetric("id20", -1.55),
				),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id10", 1.55),
					model.NewGaugeMetric("id20", -1.55),
				),
			},
		},
		"non-empty storage fails and produces no metrics": {
			fields: fields{
				collected: storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 0),
					model.NewGaugeMetric("id10", 1.55),
					model.NewGaugeMetric("id20", -1.55),
				).MakeFaulty(),
			},
			args: args{timeout: 100 * time.Millisecond},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, storagetest.ErrFaultyStorage)
				},
				metrics: model.NewMetricSet(),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			ctx, cancel := context.WithTimeout(t.Context(), tt.args.timeout)
			defer cancel()
			c := &collector{
				collected: tt.fields.collected,
			}

			// Act
			gotCh, err := c.Snapshot(ctx)

			// Assert
			tt.want.wantErr(t, err)
			gotMetrics := model.NewMetricSet()
			for metric := range gotCh {
				gotMetrics.Upsert(metric)
			}
			assert.Equal(t, tt.want.metrics, gotMetrics)
		})
	}
}
