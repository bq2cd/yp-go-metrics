package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/testutil"
)

type mockPeriodicTask struct {
	testutil.Mock

	workDuration func() time.Duration
	wantErr      func() bool
}

func (m *mockPeriodicTask) doWork(ctx context.Context) error {
	m.Called(ctx)
	time.Sleep(m.workDuration())

	if m.wantErr() {
		return fmt.Errorf("work error")
	}
	return nil
}

func Test_agent_Run(t *testing.T) {
	type collector struct {
		metrics []model.Metric
	}
	type reporter struct {
		timeout time.Duration
	}
	type fields struct {
		config    config.Config
		collector collector
		reporter  reporter
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
			timeout: 50 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 12 * time.Millisecond, ReportInterval: 23 * time.Millisecond},
				collector: collector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: reporter{},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  2,
			},
		},
		{
			name:    "slow reporter",
			timeout: 50 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 12 * time.Millisecond, ReportInterval: 23 * time.Millisecond},
				collector: collector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: reporter{timeout: 40 * time.Millisecond},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  1,
			},
		},
		{
			name:    "slow reporter 2",
			timeout: 50 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 12 * time.Millisecond, ReportInterval: 12 * time.Millisecond},
				collector: collector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: reporter{timeout: 30 * time.Millisecond},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 5,
				numCallsReport:  2,
			},
		},
		{
			name:    "slow reporter 3",
			timeout: 80 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 12 * time.Millisecond, ReportInterval: 12 * time.Millisecond},
				collector: collector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: reporter{timeout: 20 * time.Millisecond},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 7,
				numCallsReport:  3,
			},
		},
		{
			name:    "poll interval > report interval",
			timeout: 50 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 20 * time.Millisecond, ReportInterval: 15 * time.Millisecond},
				collector: collector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: reporter{},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 3,
				numCallsReport:  3,
			},
		},
		{
			name:    "poll interval == report interval",
			timeout: 50 * time.Millisecond,
			fields: fields{
				config: config.Config{PollInterval: 15 * time.Millisecond, ReportInterval: 15 * time.Millisecond},
				collector: collector{
					metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				},
				reporter: reporter{},
			},
			want: want{
				metrics:         []model.Metric{model.NewCounterMetric("id1", 5), model.NewGaugeMetric("id2", 0.3)},
				numCallsCollect: 4,
				numCallsReport:  3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			mCollector := &mockCollector{metrics: tt.fields.collector.metrics}
			mReporter := &mockReporter{timeout: tt.fields.reporter.timeout}

			a := New(nil, tt.fields.config, mCollector, mReporter)

			mCollector.
				On("Collect", mock.Anything).Return(nil).
				On("Snapshot", mock.Anything).Return(mock.Anything, nil)
			mReporter.On("Report", mock.Anything, mock.Anything).Return(nil)

			err := a.Run(ctx)
			require.NoError(t, err)

			mCollector.AssertExpectations(t)
			mReporter.AssertExpectations(t)

			testutil.AssertNumberInBetween(t, tt.want.numCallsCollect-1, tt.want.numCallsCollect+1, mCollector.GetNumCalls("Collect"))
			testutil.AssertNumberInBetween(t, tt.want.numCallsReport-1, tt.want.numCallsReport+1, mCollector.GetNumCalls("Snapshot"))
			testutil.AssertNumberInBetween(t, tt.want.numCallsReport-1, tt.want.numCallsReport+1, mReporter.GetNumCalls("Report"))
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
				config:    config.Config{},
				collector: tt.fields.collector,
				reporter:  tt.fields.reporter,
			}
			tt.fields.collector.On("Snapshot", mock.Anything).Return(mock.Anything, mock.Anything)
			if !tt.fields.collector.wantErr {
				tt.fields.reporter.On("Report", mock.Anything, mock.Anything).Return(mock.Anything)
			}

			tt.assertion(t, a.doReport(t.Context()))

			tt.fields.collector.AssertExpectations(t)
			tt.fields.reporter.AssertExpectations(t)
		})
	}
}

func Test_runPeriodicTask(t *testing.T) {
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
				require.NoError(t, err)
				m.AssertExpectations(t)
				testutil.AssertNumberInBetween(t, 6, 8, m.GetNumCalls("doWork"))
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
				require.NoError(t, err)
				m.AssertExpectations(t)
				testutil.AssertNumberInBetween(t, 4, 6, m.GetNumCalls("doWork"))
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
				require.NoError(t, err)
				m.AssertExpectations(t)
				testutil.AssertNumberInBetween(t, 1, 2, m.GetNumCalls("doWork"))
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
				require.NoError(t, err)
				m.AssertExpectations(t)
				testutil.AssertNumberInBetween(t, 1, 3, m.GetNumCalls("doWork"))
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
				require.NoError(t, err)
				m.AssertExpectations(t)
				testutil.AssertNumberInBetween(t, 1, 2, m.GetNumCalls("doWork"))
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
				require.NoError(t, err)
				m.AssertExpectations(t)
				testutil.AssertNumberInBetween(t, 1, 3, m.GetNumCalls("doWork"))
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
				require.Error(t, err)
				m.AssertExpectations(t)
				testutil.AssertNumberInBetween(t, 2, 4, m.GetNumCalls("doWork"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			tt.args.mockTask.On("doWork", mock.Anything).Return(mock.Anything)

			tt.assertion(t, tt.args.mockTask, runPeriodicTask(ctx, tt.args.interval, tt.args.mockTask.doWork, tt.args.initialDelay))
		})
	}
}
