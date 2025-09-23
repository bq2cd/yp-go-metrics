package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPeriodicTask struct {
	mock.Mock
	workDuration func() time.Duration
	wantErr      func() bool
}

func (m *mockPeriodicTask) doWork() error {
	m.Called()
	time.Sleep(m.workDuration())

	if m.wantErr() {
		return fmt.Errorf("work error")
	}
	return nil
}

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

func Test_agent_Run(t *testing.T) {
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
			timeout: 28 * time.Millisecond,
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
			timeout: 28 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 6 * time.Millisecond, ReportInterval: 15 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{timeout: 35 * time.Millisecond},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  1,
			},
		},
		{
			name:    "slow reporter 3",
			timeout: 28 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 6 * time.Millisecond, ReportInterval: 8 * time.Millisecond},
				collector: &mockCollector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: &mockReporter{timeout: 35 * time.Millisecond},
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

			a := &agent{
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

func Test_agent_doReport(t *testing.T) {
	type fields struct {
		collector *mockCollector
		reporter  *mockReporter
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "no errors",
			fields: fields{
				collector: &mockCollector{},
				reporter:  &mockReporter{},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "faulty collector",
			fields: fields{
				collector: &mockCollector{wantErr: true},
				reporter:  &mockReporter{},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Errorf(t, err, "snapshot error")
			},
		},
		{
			name: "faulty reporter",
			fields: fields{
				collector: &mockCollector{},
				reporter:  &mockReporter{wantErr: true},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Errorf(t, err, "report error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &agent{
				context:   t.Context(),
				config:    config.Config{},
				collector: tt.fields.collector,
				reporter:  tt.fields.reporter,
			}
			tt.fields.collector.On("Snapshot").Return(mock.Anything, mock.Anything)
			tt.fields.reporter.On("Report", mock.Anything).Return(mock.Anything)

			tt.assertion(t, a.doReport())

			tt.fields.collector.AssertExpectations(t)
			tt.fields.reporter.AssertExpectations(t)
		})
	}
}

func TestRunPeriodicTask(t *testing.T) {
	type args struct {
		interval     time.Duration
		mockTask     *mockPeriodicTask
		initialDelay time.Duration
	}
	tests := []struct {
		name      string
		timeout   time.Duration
		args      args
		assertion func(*testing.T, *mockPeriodicTask, error)
	}{
		// TODO: Add test cases.
		{
			name:    "fast task without initial delay",
			timeout: 100 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockPeriodicTask{
					workDuration: func() time.Duration { return 5 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 0,
			},
			assertion: func(t *testing.T, m *mockPeriodicTask, err error) {
				assert.NoError(t, err)
				m.AssertExpectations(t)
				m.AssertNumberOfCalls(t, "doWork", 7)
			},
		},
		{
			name:    "fast task with initial delay",
			timeout: 100 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockPeriodicTask{
					workDuration: func() time.Duration { return 5 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 30 * time.Millisecond,
			},
			assertion: func(t *testing.T, m *mockPeriodicTask, err error) {
				assert.NoError(t, err)
				m.AssertExpectations(t)
				m.AssertNumberOfCalls(t, "doWork", 5)
			},
		},
		{
			name:    "slow task without initial delay",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockPeriodicTask{
					workDuration: func() time.Duration { return 100 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 0,
			},
			assertion: func(t *testing.T, m *mockPeriodicTask, err error) {
				assert.NoError(t, err)
				m.AssertExpectations(t)
				m.AssertNumberOfCalls(t, "doWork", 1)
			},
		},
		{
			name:    "slow task without initial delay 2",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockPeriodicTask{
					workDuration: func() time.Duration { return 30 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 0,
			},
			assertion: func(t *testing.T, m *mockPeriodicTask, err error) {
				assert.NoError(t, err)
				m.AssertExpectations(t)
				m.AssertNumberOfCalls(t, "doWork", 2)
			},
		},
		{
			name:    "slow task with initial delay",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockPeriodicTask{
					workDuration: func() time.Duration { return 30 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 30 * time.Millisecond,
			},
			assertion: func(t *testing.T, m *mockPeriodicTask, err error) {
				assert.NoError(t, err)
				m.AssertExpectations(t)
				m.AssertNumberOfCalls(t, "doWork", 1)
			},
		},
		{
			name:    "slow task with initial delay 2",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockPeriodicTask{
					workDuration: func() time.Duration { return 20 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 10 * time.Millisecond,
			},
			assertion: func(t *testing.T, m *mockPeriodicTask, err error) {
				assert.NoError(t, err)
				m.AssertExpectations(t)
				m.AssertNumberOfCalls(t, "doWork", 2)
			},
		},
		{
			name:    "always faulty task",
			timeout: 55 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockPeriodicTask{
					workDuration: func() time.Duration { return 10 * time.Millisecond },
					wantErr:      func() bool { return true },
				},
				initialDelay: 5 * time.Millisecond,
			},
			assertion: func(t *testing.T, m *mockPeriodicTask, err error) {
				assert.Error(t, err)
				m.AssertExpectations(t)
				m.AssertNumberOfCalls(t, "doWork", 3)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			tt.args.mockTask.On("doWork").Return(mock.Anything)

			tt.assertion(t, tt.args.mockTask, RunPeriodicTask(ctx, tt.args.interval, tt.args.mockTask.doWork, tt.args.initialDelay))
		})
	}
}
