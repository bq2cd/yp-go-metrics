package agent

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/testutil"
)

type mockReporter struct {
	testutil.Mock

	metrics []model.Metric
	timeout time.Duration
	wantErr bool
	mu      sync.Mutex
}

func (m *mockReporter) Report(ctx context.Context, inCh <-chan model.Metric) error {
	m.Called(ctx, inCh)
	m.mu.Lock()
	defer m.mu.Unlock()
	for metric := range inCh {
		m.metrics = append(m.metrics, metric)
	}
	if m.timeout > 0 {
		time.Sleep(m.timeout)
	}
	if m.wantErr {
		return errors.New("report error")
	}
	return nil
}

func TestNewReporter(t *testing.T) {
	type args struct {
		sender  *mockSender
		storage *storagetest.MockStorage
	}
	tests := []struct {
		name      string
		args      args
		assertion func(assert.TestingT, args, *reporter)
	}{
		{
			name: "empty",
			args: args{},
			assertion: func(t assert.TestingT, args args, r *reporter) {
				assert.Nil(t, r.sender)
				assert.Nil(t, r.reported)
			},
		},
		{
			name: "sender only",
			args: args{sender: &mockSender{}},
			assertion: func(t assert.TestingT, args args, r *reporter) {
				assert.Equal(t, args.sender, r.sender)
				assert.Nil(t, r.reported)
			},
		},
		{
			name: "sender + storage",
			args: args{sender: &mockSender{}, storage: storagetest.NewMockStorage()},
			assertion: func(t assert.TestingT, args args, r *reporter) {
				assert.Equal(t, args.sender, r.sender)
				assert.Equal(t, args.storage, r.reported)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numWorkers := 3
			got := NewReporter(tt.args.sender, tt.args.storage, uint(numWorkers))
			tt.assertion(t, tt.args, got)
			assert.Equal(t, uint(numWorkers), got.senderPoolSize)
			assert.Equal(t, uint(defaultSenderBatchSize), got.senderBatchSize)
		})
	}
}

func Test_reporter_getSendableMetric(t *testing.T) {
	type fields struct {
		reported *storagetest.MockStorage
	}
	type args struct {
		metric model.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   model.Metric
	}{
		{
			name:   "empty metric, empty storage",
			fields: fields{reported: storagetest.NewMockStorage()},
			args:   args{metric: model.Metric{}},
			want:   model.Metric{},
		},
		{
			name:   "empty counter, empty storage",
			fields: fields{reported: storagetest.NewMockStorage()},
			args:   args{metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"}},
			want:   model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
		},
		{
			name:   "some counter, empty storage",
			fields: fields{reported: storagetest.NewMockStorage()},
			args:   args{metric: model.NewCounterMetric("id1", 5)},
			want:   model.NewCounterMetric("id1", 5),
		},
		{
			name:   "empty counter, non-empty storage",
			fields: fields{reported: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:   args{metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"}},
			want:   model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
		},
		{
			name:   "some counter, non-empty storage",
			fields: fields{reported: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5))},
			args:   args{metric: model.NewCounterMetric("id1", 10)},
			want:   model.NewCounterMetric("id1", 5),
		},
		{
			name:   "some counter, non-empty storage with nil delta",
			fields: fields{reported: storagetest.NewMockStorage(model.Metric{Type: model.MetricTypeCounter, ID: "id1"})},
			args:   args{metric: model.NewCounterMetric("id1", 10)},
			want:   model.NewCounterMetric("id1", 10),
		},
		{
			name:   "some counter, non-empty faulty storage",
			fields: fields{reported: storagetest.NewMockStorage(model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 5)).MakeFaulty()},
			args:   args{metric: model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 10)},
			want:   model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 10),
		},
		{
			name:   "some gauge, non-empty storage",
			fields: fields{reported: storagetest.NewMockStorage(model.NewGaugeMetric("id2", 5.3))},
			args:   args{metric: model.NewGaugeMetric("id2", 1.1)},
			want:   model.NewGaugeMetric("id2", 1.1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reporter{
				reported: tt.fields.reported,
			}
			assert.Equal(t, tt.want, r.getSendableMetric(t.Context(), tt.args.metric))
		})
	}
}

func Test_reporter_Report(t *testing.T) {
	type fields struct {
		sender     *mockSender
		reported   *storagetest.MockStorage
		numWorkers uint
		batchSize  uint
	}
	type args struct {
		timeout time.Duration
		metrics []model.Metric
	}
	type want struct {
		sentMetrics   []model.Metric
		storedMetrics []model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion func(*testing.T, error)
	}{
		{
			name: "invalid metric",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{{Type: model.MetricTypeCounter, ID: "id1"}},
			},
			want: want{
				sentMetrics:   []model.Metric{},
				storedMetrics: []model.Metric{},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send single counter, newly reported",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send single counter, previously reported",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(model.NewCounterMetric("id1", -5)),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 10)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send single counter, previously reported with the same value",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(model.NewCounterMetric("id1", 5)),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 0)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send single gauge, newly reported",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewGaugeMetric("id1", -5.5)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewGaugeMetric("id1", -5.5)},
				storedMetrics: []model.Metric{model.NewGaugeMetric("id1", -5.5)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send single gauge, previously reported",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(model.NewGaugeMetric("id1", 3.8)),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewGaugeMetric("id1", -5.5)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewGaugeMetric("id1", -5.5)},
				storedMetrics: []model.Metric{model.NewGaugeMetric("id1", -5.5)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "sender error, single counter, newly reported",
			fields: fields{
				sender: &mockSender{
					wantBatchErr: func(metrics model.MetricSet) (model.MetricSet, error) {
						return nil, errors.New("something went wrong")
					},
				},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5)},
				storedMetrics: []model.Metric{},
			},
			assertion: func(t *testing.T, err error) {
				assert.Errorf(t, err, "something went wrong")
			},
		},
		{
			name: "sender error, single counter, previously reported",
			fields: fields{
				sender: &mockSender{
					wantBatchErr: func(metrics model.MetricSet) (model.MetricSet, error) {
						return nil, errors.New("something went wrong")
					},
				},
				reported:  storagetest.NewMockStorage(model.NewCounterMetric("id1", 10)),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", -5)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 10)},
			},
			assertion: func(t *testing.T, err error) {
				assert.Errorf(t, err, "something went wrong")
			},
		},
		{
			name: "send multiple counters",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send multiple counters with the same id",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", -10), model.NewCounterMetric("id1", 7)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 7)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 7)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send multiple counters with the same id (storage already contains previous value)",
			fields: fields{
				sender: &mockSender{},
				reported: storagetest.NewMockStorage(
					model.NewCounterMetric("id1", -81),
				),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", -10), model.NewCounterMetric("id1", 7)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 88)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 7)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send multiple metrics",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "sender error on single metric",
			fields: fields{
				sender: &mockSender{
					wantBatchErr: func(metrics model.MetricSet) (model.MetricSet, error) {
						sent := model.NewMetricSet()
						var err error
						for _, m := range metrics {
							if m.Type == model.MetricTypeGauge && m.ID == "id1" {
								err = errors.New("no luck here")
								continue
							}
							sent.Upsert(m)
						}
						return sent, err
					},
				},
				reported:  storagetest.NewMockStorage(),
				batchSize: defaultSenderBatchSize,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id2", -3.01)},
			},
			assertion: func(t *testing.T, err error) {
				assert.Errorf(t, err, "no luck here")
			},
		},
		{
			name: "send multiple metrics in multiple batches",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: 2,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send multiple metrics in multiple batches (non-aligned)",
			fields: fields{
				sender:    &mockSender{},
				reported:  storagetest.NewMockStorage(),
				batchSize: 3,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
			},
			want: want{
				sentMetrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
				storedMetrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send multiple metrics in multiple batches in parallel",
			fields: fields{
				sender:     &mockSender{delay: 75 * time.Millisecond},
				reported:   storagetest.NewMockStorage(),
				batchSize:  2,
				numWorkers: 3,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
			},
			want: want{
				sentMetrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
				storedMetrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send multiple metrics in multiple batches in parallel with slow sender",
			fields: fields{
				sender:     &mockSender{delay: 500 * time.Millisecond},
				reported:   storagetest.NewMockStorage(),
				batchSize:  2,
				numWorkers: 3,
			},
			args: args{
				timeout: 100 * time.Millisecond,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
			},
			want: want{
				sentMetrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -7),
					model.NewGaugeMetric("id1", -5),
					model.NewGaugeMetric("id2", -3.01),
				},
				storedMetrics: []model.Metric{},
			},
			assertion: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "context deadline exceeded")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := context.WithTimeout(t.Context(), tt.args.timeout)
			defer cancel()
			r := &reporter{
				sender:          tt.fields.sender,
				reported:        tt.fields.reported,
				senderPoolSize:  tt.fields.numWorkers,
				senderBatchSize: tt.fields.batchSize,
			}
			if len(tt.want.sentMetrics) > 0 {
				n, r := len(tt.want.sentMetrics)/int(tt.fields.batchSize), len(tt.want.sentMetrics)%int(tt.fields.batchSize)
				if r != 0 {
					n++
				}
				for i := 0; i < n; i++ {
					start := i * int(tt.fields.batchSize)
					end := min(len(tt.want.sentMetrics), (i+1)*int(tt.fields.batchSize))
					sentMetrics := tt.want.sentMetrics[start:end]
					tt.fields.sender.On("SendBatch", ctx, model.NewMetricSet(sentMetrics...)).Return(mock.Anything, mock.AnythingOfType("error"))
				}
			}

			// Act
			inCh := make(chan model.Metric)
			go func() {
				defer close(inCh)
				for _, metric := range tt.args.metrics {
					inCh <- metric
				}
			}()

			err := r.Report(ctx, inCh)

			// Assert
			tt.assertion(t, err)
			tt.fields.sender.AssertExpectations(t)
			for _, m := range tt.want.storedMetrics {
				got, err := tt.fields.reported.Get(t.Context(), m.Key())
				require.NoErrorf(t, err, "expected metric %v", m)
				assert.Equal(t, m, got)
			}
			if len(tt.want.storedMetrics) == 0 {
				metrics, err := tt.fields.reported.GetAll(t.Context())
				require.NoError(t, err)
				assert.Empty(t, metrics)
			}
		})
	}
}
