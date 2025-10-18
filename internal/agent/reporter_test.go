package agent

import (
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

type mockReporter struct {
	mock.Mock
	metrics []model.Metric
	timeout time.Duration
	wantErr bool
	mu      sync.Mutex
}

func (m *mockReporter) Report(metrics []model.Metric) error {
	m.Called(metrics)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = metrics
	if m.timeout > 0 {
		time.Sleep(m.timeout)
	}
	if m.wantErr {
		return errors.New("report error")
	}
	return nil
}

func Test_reporter_reportSingle(t *testing.T) {
	type fields struct {
		sender   *mockSender
		reported *storagetest.MockStorage
	}
	type args struct {
		metric model.Metric
	}
	type want struct {
		metricSent   model.Metric
		metricStored model.Metric
		checkErr     func(*testing.T, error)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "send counter without value",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			},
			want: want{
				metricSent:   model.Metric{},
				metricStored: model.Metric{},
				checkErr: func(t *testing.T, err error) {
					assert.ErrorIs(t, err, ErrReporterEmptyMetric)
				},
			},
		},
		{
			name: "send counter without value 2",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(model.NewCounterMetric("id1", -5)),
			},
			args: args{
				metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			},
			want: want{
				metricSent:   model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
				metricStored: model.NewCounterMetric("id1", -5),
				checkErr: func(t *testing.T, err error) {
					assert.ErrorIs(t, err, ErrReporterEmptyMetric)
				},
			},
		},
		{
			name: "send gauge without value",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metric: model.Metric{Type: model.MetricTypeGauge, ID: "id1"},
			},
			want: want{
				metricSent:   model.Metric{},
				metricStored: model.Metric{},
				checkErr: func(t *testing.T, err error) {
					assert.ErrorIs(t, err, ErrReporterEmptyMetric)
				},
			},
		},
		{
			name: "send counter",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metric: model.NewCounterMetric("id1", 5),
			},
			want: want{
				metricSent:   model.NewCounterMetric("id1", 5),
				metricStored: model.NewCounterMetric("id1", 5),
				checkErr:     func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "send counter 2",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(model.NewCounterMetric("id1", -5)),
			},
			args: args{
				metric: model.NewCounterMetric("id1", 5),
			},
			want: want{
				metricSent:   model.NewCounterMetric("id1", 10),
				metricStored: model.NewCounterMetric("id1", 5),
				checkErr:     func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "send counter 3",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(model.NewCounterMetric("id1", 5)),
			},
			args: args{
				metric: model.NewCounterMetric("id1", 5),
			},
			want: want{
				metricSent:   model.NewCounterMetric("id1", 0),
				metricStored: model.NewCounterMetric("id1", 5),
				checkErr:     func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "send gauge",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metric: model.NewGaugeMetric("id1", -5.5),
			},
			want: want{
				metricSent:   model.NewGaugeMetric("id1", -5.5),
				metricStored: model.NewGaugeMetric("id1", -5.5),
				checkErr:     func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "send gauge 2",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(model.NewGaugeMetric("id1", 3.8)),
			},
			args: args{
				metric: model.NewGaugeMetric("id1", -5.5),
			},
			want: want{
				metricSent:   model.NewGaugeMetric("id1", -5.5),
				metricStored: model.NewGaugeMetric("id1", -5.5),
				checkErr:     func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "sender error",
			fields: fields{
				sender: &mockSender{
					wantErr: func(m model.Metric) error { return errors.New("something went wrong") },
				},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metric: model.NewCounterMetric("id1", 5),
			},
			want: want{
				metricSent:   model.NewCounterMetric("id1", 5),
				metricStored: model.Metric{},
				checkErr:     func(t *testing.T, err error) { assert.Errorf(t, err, "something went wrong") },
			},
		},
		{
			name: "sender error 2",
			fields: fields{
				sender: &mockSender{
					wantErr: func(m model.Metric) error { return errors.New("something went wrong") },
				},
				reported: storagetest.NewMockStorage(model.NewCounterMetric("id1", 10)),
			},
			args: args{
				metric: model.NewCounterMetric("id1", 5),
			},
			want: want{
				metricSent:   model.NewCounterMetric("id1", -5),
				metricStored: model.NewCounterMetric("id1", 10),
				checkErr:     func(t *testing.T, err error) { assert.Errorf(t, err, "something went wrong") },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reporter{
				sender:   tt.fields.sender,
				reported: tt.fields.reported,
			}
			tt.fields.sender.On("Send", tt.want.metricSent).Return(mock.AnythingOfType("error"))
			metric := tt.args.metric.Copy()

			err := r.reportSingle(metric)

			defer func() {
				assert.Equal(t, tt.args.metric, metric)
			}()
			if tt.want.metricSent.Empty() {
				tt.fields.sender.AssertNotCalled(t, "Send")
			} else {
				tt.fields.sender.AssertExpectations(t)
			}
			tt.want.checkErr(t, err)
			if !tt.want.metricStored.Empty() {
				got, err := tt.fields.reported.Get(tt.want.metricStored.Key())
				require.NoError(t, err)
				assert.Equal(t, tt.want.metricStored, got)
			}
		})
	}
}

func TestNewReporter(t *testing.T) {
	type args struct {
		sender  Sender
		storage repository.Storage
	}
	tests := []struct {
		name      string
		args      args
		assertion func(assert.TestingT, args, *reporter)
	}{
		{
			name: "emtpy",
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
			tt.assertion(t, tt.args, NewReporter(tt.args.sender, tt.args.storage))
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
			assert.Equal(t, tt.want, r.getSendableMetric(tt.args.metric))
		})
	}
}

func Test_reporter_Report(t *testing.T) {
	type fields struct {
		sender   *mockSender
		reported *storagetest.MockStorage
	}
	type args struct {
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
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metrics: []model.Metric{{Type: model.MetricTypeCounter, ID: "id1"}},
			},
			want: want{
				sentMetrics:   []model.Metric{},
				storedMetrics: []model.Metric{},
			},
			assertion: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "send single counter",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
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
			name: "send multiple counters",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
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
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", -10), model.NewCounterMetric("id1", 7)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", -15), model.NewCounterMetric("id1", 17)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 7)},
			},
			assertion: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "send multiple metrics",
			fields: fields{
				sender:   &mockSender{},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
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
					wantErr: func(m model.Metric) error {
						if m.Type == model.MetricTypeGauge && m.ID == "id1" {
							return errors.New("no luck here")
						}
						return nil
					},
				},
				reported: storagetest.NewMockStorage(),
			},
			args: args{
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
			},
			want: want{
				sentMetrics:   []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id2", -3.01)},
				storedMetrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id2", -3.01)},
			},
			assertion: func(t *testing.T, err error) {
				assert.Errorf(t, err, "no luck here")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reporter{
				sender:   tt.fields.sender,
				reported: tt.fields.reported,
			}
			sender := tt.fields.sender.On("Send", mock.AnythingOfType("model.Metric")).Return(mock.AnythingOfType("error"))

			err := r.Report(tt.args.metrics)

			tt.assertion(t, err)
			for _, m := range tt.want.sentMetrics {
				sender.Parent.AssertCalled(t, "Send", m)
			}
			for _, m := range tt.want.storedMetrics {
				got, err := tt.fields.reported.Get(m.Key())
				require.NoError(t, err)
				assert.Equal(t, m, got)
			}
		})
	}
}
