package periodictask

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTimerTask struct {
	mock.Mock
	workDuration func() time.Duration
	wantErr      func() bool
}

func (m *mockTimerTask) doWork(ctx context.Context) error {
	m.Called(ctx)

	time.Sleep(m.workDuration())

	if m.wantErr() {
		return fmt.Errorf("work error")
	}
	return nil
}

func TestNewTimerTask(t *testing.T) {
	type args struct {
		ctx          context.Context
		interval     time.Duration
		taskFn       TaskTimerFn
		initialDelay time.Duration
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, *timerTask)
	}{
		{
			name: "default",
			args: args{
				ctx:          context.Background(),
				interval:     1 * time.Millisecond,
				taskFn:       func(ctx context.Context) error { return nil },
				initialDelay: 0,
			},
			assertion: func(t *testing.T, args args, got *timerTask) {
				assert.Equal(t, args.ctx, got.context)
				assert.Equal(t, args.interval, got.interval)
				assert.Equal(t, reflect.ValueOf(args.taskFn), reflect.ValueOf(got.taskFn))
				assert.Equal(t, args.initialDelay, got.initialDelay)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewTimerTask(tt.args.ctx, tt.args.interval, tt.args.taskFn, tt.args.initialDelay))
		})
	}
}

func Test_timerTask_Run(t *testing.T) {
	type args struct {
		interval     time.Duration
		mockTask     *mockTimerTask
		initialDelay time.Duration
	}
	type want struct {
		calls int
	}
	tests := []struct {
		name      string
		timeout   time.Duration
		args      args
		want      want
		assertion func(*testing.T, *mockTimerTask, error)
	}{
		{
			name:    "fast task without initial delay",
			timeout: 100 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockTimerTask{
					workDuration: func() time.Duration { return 5 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 0,
			},
			want: want{
				calls: 7,
			},
			assertion: func(t *testing.T, m *mockTimerTask, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "fast task with initial delay",
			timeout: 100 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockTimerTask{
					workDuration: func() time.Duration { return 5 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 30 * time.Millisecond,
			},
			want: want{
				calls: 5,
			},
			assertion: func(t *testing.T, m *mockTimerTask, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "very slow task without initial delay",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockTimerTask{
					workDuration: func() time.Duration { return 100 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 0,
			},
			want: want{
				calls: 1,
			},
			assertion: func(t *testing.T, m *mockTimerTask, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "slow task without initial delay",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockTimerTask{
					workDuration: func() time.Duration { return 30 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 0,
			},
			want: want{
				calls: 2,
			},
			assertion: func(t *testing.T, m *mockTimerTask, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "slow task with initial delay",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockTimerTask{
					workDuration: func() time.Duration { return 30 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 30 * time.Millisecond,
			},
			want: want{
				calls: 1,
			},
			assertion: func(t *testing.T, m *mockTimerTask, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "slow task with smaller initial delay",
			timeout: 50 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockTimerTask{
					workDuration: func() time.Duration { return 20 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
				initialDelay: 10 * time.Millisecond,
			},
			want: want{
				calls: 2,
			},
			assertion: func(t *testing.T, m *mockTimerTask, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "always faulty task",
			timeout: 55 * time.Millisecond,
			args: args{
				interval: 15 * time.Millisecond,
				mockTask: &mockTimerTask{
					workDuration: func() time.Duration { return 10 * time.Millisecond },
					wantErr:      func() bool { return true },
				},
				initialDelay: 5 * time.Millisecond,
			},
			want: want{
				calls: 3,
			},
			assertion: func(t *testing.T, m *mockTimerTask, err error) {
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			tt.args.mockTask.On("doWork", mock.Anything).Return(mock.AnythingOfType("error")).Times(tt.want.calls)
			tr := &timerTask{
				context:      ctx,
				taskFn:       tt.args.mockTask.doWork,
				interval:     tt.args.interval,
				initialDelay: tt.args.initialDelay,
			}

			err := tr.Run()

			tt.args.mockTask.AssertExpectations(t)
			tt.assertion(t, tt.args.mockTask, err)
		})
	}
}
