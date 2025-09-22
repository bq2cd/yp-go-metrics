package agent

import (
	"context"
	"testing"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	type args struct {
		ctx       context.Context
		cfg       config.Config
		collector Collector
		reporter  Reporter
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default initialisation",
			args: args{
				ctx:       context.Background(),
				cfg:       config.Config{},
				collector: &defaultCollector{},
				reporter:  &defaultReporter{},
			},
		},
		{
			name: "mock initialisation",
			args: args{
				ctx:       context.Background(),
				cfg:       config.Config{},
				collector: &mockCollector{},
				reporter:  &mockReporter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAgent(tt.args.ctx, tt.args.cfg, tt.args.collector, tt.args.reporter)
			assert.Equal(t, tt.args.ctx, got.context)
			assert.Equal(t, tt.args.cfg, got.config)
			assert.Equal(t, tt.args.collector, got.collector)
			assert.Equal(t, tt.args.reporter, got.reporter)
		})
	}
}

func Test_agentWorker_Run(t *testing.T) {
	type fields struct {
		config    config.Config
		collector *mockCollector
		reporter  *mockReporter
	}
	type want struct {
		metrics         []model.Metric
		numCallsCollect int
		numCallsReport  int
	}
	tests := []struct {
		name    string
		timeout time.Duration
		fields  fields
		want    want
	}{
		{
			name:    "normal flow",
			timeout: 23 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 5 * time.Millisecond, ReportInterval: 10 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  2,
			},
		},
		{
			name:    "slow reporter",
			timeout: 27 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 6 * time.Millisecond, ReportInterval: 15 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{timeout: 20 * time.Millisecond},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  1,
			},
		},
		{
			name:    "slow reporter 2",
			timeout: 27 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 6 * time.Millisecond, ReportInterval: 15 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{timeout: 30 * time.Millisecond},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  1,
			},
		},
		{
			name:    "slow reporter 3",
			timeout: 27 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 6 * time.Millisecond, ReportInterval: 8 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{timeout: 30 * time.Millisecond},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  1,
			},
		},
		{
			name:    "poll interval > report interval",
			timeout: 25 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 20 * time.Millisecond, ReportInterval: 10 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 2,
				numCallsReport:  2,
			},
		},
		{
			name:    "poll interval == report interval",
			timeout: 25 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 10 * time.Millisecond, ReportInterval: 10 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 3,
				numCallsReport:  2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			a := &agentWorker{
				context:   ctx,
				config:    tt.fields.config,
				collector: tt.fields.collector,
				reporter:  tt.fields.reporter,
			}

			mCollector := tt.fields.collector.On("Collect").Return(nil).On("Snapshot").Return(tt.want.metrics, nil)
			mReporter := tt.fields.reporter.On("Report", tt.want.metrics).Return(nil)

			err := a.Run()
			assert.NoError(t, err)

			mCollector.Parent.AssertExpectations(t)
			mReporter.Parent.AssertExpectations(t)

			mCollector.Parent.AssertNumberOfCalls(t, "Collect", tt.want.numCallsCollect)
			mCollector.Parent.AssertNumberOfCalls(t, "Snapshot", tt.want.numCallsReport)
			mReporter.Parent.AssertNumberOfCalls(t, "Report", tt.want.numCallsReport)

		})
	}
}
