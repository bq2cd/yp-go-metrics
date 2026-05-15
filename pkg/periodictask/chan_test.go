package periodictask

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockChanTask[T any] struct {
	mock.Mock

	mu           sync.Mutex
	workDuration func() time.Duration
	wantErr      func() bool
	elapsed      time.Duration
}

func (m *mockChanTask[T]) numCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Calls)
}

func (m *mockChanTask[T]) doWork(ctx context.Context, v T) error {
	m.mu.Lock()
	m.Called(ctx, v)
	m.mu.Unlock()

	work := m.workDuration()
	time.Sleep(work)
	m.elapsed += work

	if m.wantErr() {
		return fmt.Errorf("work error")
	}
	return nil
}

func TestNewChanTask(t *testing.T) {
	type args struct {
		incoming <-chan any
		taskFn   TaskChanFn[any]
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, *chanTask[any])
	}{
		{
			name: "any",
			args: args{
				incoming: make(<-chan any),
				taskFn:   func(ctx context.Context, v any) error { return nil },
			},
			assertion: func(t *testing.T, args args, got *chanTask[any]) {
				assert.Equal(t, reflect.ValueOf(args.incoming), reflect.ValueOf(got.incoming))
				assert.Equal(t, reflect.ValueOf(args.taskFn), reflect.ValueOf(got.taskFn))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewChanTask(tt.args.incoming, tt.args.taskFn))
		})
	}
}

func Test_chanTask_Run(t *testing.T) {
	type args struct {
		incoming chan any
		mockTask *mockChanTask[any]
	}
	type want struct {
		calls int
	}
	tests := []struct {
		name      string
		timeout   time.Duration
		args      args
		want      want
		assertion func(*testing.T, *mockChanTask[any], error)
	}{
		{
			name:    "fast task",
			timeout: 99 * time.Millisecond,
			args: args{
				incoming: make(chan any),
				mockTask: &mockChanTask[any]{
					workDuration: func() time.Duration { return 14 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
			},
			want: want{
				calls: 7,
			},
			assertion: func(t *testing.T, m *mockChanTask[any], err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "slow task",
			timeout: 50 * time.Millisecond,
			args: args{
				incoming: make(chan any),
				mockTask: &mockChanTask[any]{
					workDuration: func() time.Duration { return 30 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
			},
			want: want{
				calls: 2,
			},
			assertion: func(t *testing.T, m *mockChanTask[any], err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "very slow task",
			timeout: 50 * time.Millisecond,
			args: args{
				incoming: make(chan any),
				mockTask: &mockChanTask[any]{
					workDuration: func() time.Duration { return 100 * time.Millisecond },
					wantErr:      func() bool { return false },
				},
			},
			want: want{
				calls: 1,
			},
			assertion: func(t *testing.T, m *mockChanTask[any], err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "always faulty task",
			timeout: 55 * time.Millisecond,
			args: args{
				incoming: make(chan any),
				mockTask: &mockChanTask[any]{
					workDuration: func() time.Duration { return 10 * time.Millisecond },
					wantErr:      func() bool { return true },
				},
			},
			want: want{
				calls: 6,
			},
			assertion: func(t *testing.T, m *mockChanTask[any], err error) {
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			tt.args.mockTask.On("doWork", mock.Anything, mock.Anything).Return(mock.AnythingOfType("error"))
			tr := &chanTask[any]{
				incoming: tt.args.incoming,
				taskFn:   tt.args.mockTask.doWork,
			}

			errCh := make(chan error, 1)
			go func() {
				errCh <- tr.Run(ctx)
			}()
			go func() {
				for i := range math.MaxUint16 {
					tt.args.incoming <- i
				}
			}()
			err := <-errCh

			tt.args.mockTask.AssertExpectations(t)
			assert.GreaterOrEqual(t, tt.args.mockTask.numCalls(), tt.want.calls-1)
			assert.LessOrEqual(t, tt.args.mockTask.numCalls(), tt.want.calls+1)
			tt.assertion(t, tt.args.mockTask, err)
		})
	}
}
