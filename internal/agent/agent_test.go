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
		storer    Storer
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
				storer:    &defaultStorer{},
				reporter:  &defaultReporter{},
			},
		},
		{
			name: "mock initialisation",
			args: args{
				ctx:       context.Background(),
				cfg:       config.Config{},
				collector: &mockCollector{},
				storer:    &mockStorer{},
				reporter:  &mockReporter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAgent(tt.args.ctx, tt.args.cfg, tt.args.collector, tt.args.storer, tt.args.reporter)
			assert.Equal(t, tt.args.ctx, got.context)
			assert.Equal(t, tt.args.cfg, got.config)
			assert.Equal(t, tt.args.collector, got.collector)
			assert.Equal(t, tt.args.storer, got.storer)
			assert.Equal(t, tt.args.reporter, got.reporter)
		})
	}
}

func Test_agentWorker_Run(t *testing.T) {
	type fields struct {
		config    config.Config
		collector *mockCollector
		storer    *mockStorer
		reporter  *mockReporter
	}
	type want struct {
		metrics        []model.Metric
		numCallsReport int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "normal flow",
			fields: fields{
				config: config.Config{PollInterval: 2 * time.Millisecond, ReportInterval: 10 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				storer:   &mockStorer{},
				reporter: &mockReporter{},
			},
			want: want{
				metrics:        []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsReport: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := time.Duration(tt.want.numCallsReport)*tt.fields.config.ReportInterval + tt.fields.config.PollInterval/2
			numCallsCollect := tt.want.numCallsReport * int(tt.fields.config.ReportInterval/tt.fields.config.PollInterval)

			ctx, cancel := context.WithTimeout(t.Context(), timeout)
			defer cancel()

			a := &agentWorker{
				context:   ctx,
				config:    tt.fields.config,
				collector: tt.fields.collector,
				storer:    tt.fields.storer,
				reporter:  tt.fields.reporter,
			}

			mCollector := tt.fields.collector.On("Collect").Return(tt.want.metrics, nil)
			mStorer := tt.fields.storer.On("Store", tt.want.metrics).Return(nil).On("Retrieve").Return(tt.want.metrics, nil)
			mReporter := tt.fields.reporter.On("Report", tt.want.metrics).Return(nil)

			err := a.Run()
			assert.NoError(t, err)

			mCollector.Parent.AssertExpectations(t)
			mStorer.Parent.AssertExpectations(t)
			mReporter.Parent.AssertExpectations(t)

			mCollector.Parent.AssertNumberOfCalls(t, "Collect", numCallsCollect)
			mStorer.Parent.AssertNumberOfCalls(t, "Store", numCallsCollect)
			mStorer.Parent.AssertNumberOfCalls(t, "Retrieve", tt.want.numCallsReport)
			mReporter.Parent.AssertNumberOfCalls(t, "Report", tt.want.numCallsReport)

		})
	}
}
