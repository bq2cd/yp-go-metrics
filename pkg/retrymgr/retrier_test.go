package retrymgr

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewRetrier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewNoopLogger()

	for _, l := range []log.Logger{logger, nil} {
		t.Run(fmt.Sprintf("logger %v", l), func(t *testing.T) {
			// Arrange
			sleeper := retrymgrtest.NewMockSleeper(ctrl)
			strategy := retrymgrtest.NewMockStrategy(ctrl)
			strategy.EXPECT().Name().Return("mock_strategy")

			factory := NewRetrierFactory(l, sleeper, func() Strategy { return strategy })

			// Act
			r := NewRetrier[any](factory)

			// Assert
			assert.Equal(t, r.logger, logger)
			assert.Equal(t, r.strategy, strategy)
			assert.Equal(t, r.sleeper, sleeper)
		})

	}
}

func Test_retrier_Do(t *testing.T) {
	type fields struct {
		expectStrategy func(*retrymgrtest.MockStrategy)
		expectSleeper  func(*retrymgrtest.MockSleeper)
	}
	type args struct {
		timeout       time.Duration
		taskName      string
		taskFn        RetryableFn[any]
		shouldRetryFn ShouldRetryFn
	}
	type want struct {
		got           any
		wantErr       func(testing.TB, error)
		wantLogEvents func(testing.TB, log.TestLogEventSet)
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"task succeeds, no retry needed": {
			fields: fields{
				expectStrategy: func(m *retrymgrtest.MockStrategy) {
					m.EXPECT().NextDelay().Times(0)
				},
				expectSleeper: func(m *retrymgrtest.MockSleeper) {
					m.EXPECT().Sleep(gomock.Any(), gomock.Any()).Times(0)
				},
			},
			args: args{
				timeout:  100 * time.Millisecond,
				taskName: "task1",
				taskFn: func(_ context.Context) (any, error) {
					return 1, nil
				},
				shouldRetryFn: func(_ error) bool {
					return true
				},
			},
			want: want{
				got:     1,
				wantErr: func(t testing.TB, err error) { require.NoError(t, err) },
				wantLogEvents: func(t testing.TB, events log.TestLogEventSet) {
					assert.Empty(t, events)
				},
			},
		},
		"task fails first time, one retry happens": {
			fields: fields{
				expectStrategy: func(m *retrymgrtest.MockStrategy) {
					m.EXPECT().NextDelay().Return(10*time.Millisecond, true).Times(1)
				},
				expectSleeper: func(m *retrymgrtest.MockSleeper) {
					m.EXPECT().Sleep(gomock.Any(), 10*time.Millisecond).Return(nil).Times(1)
				},
			},
			args: args{
				timeout:  100 * time.Millisecond,
				taskName: "task1",
				taskFn: func() RetryableFn[any] {
					calls := 0
					return func(_ context.Context) (any, error) {
						calls++
						if calls < 2 {
							return 0, fmt.Errorf("task error %d", calls)
						}
						return 1, nil
					}
				}(),
				shouldRetryFn: func(_ error) bool {
					return true
				},
			},
			want: want{
				got:     1,
				wantErr: func(t testing.TB, err error) { require.NoError(t, err) },
				wantLogEvents: func(t testing.TB, events log.TestLogEventSet) {
					assert.Len(t, events, 1)
				},
			},
		},
		"task fails two times, two retries happen": {
			fields: fields{
				expectStrategy: func(m *retrymgrtest.MockStrategy) {
					m.EXPECT().NextDelay().Return(10*time.Millisecond, true).Times(2)
				},
				expectSleeper: func(m *retrymgrtest.MockSleeper) {
					m.EXPECT().Sleep(gomock.Any(), 10*time.Millisecond).Return(nil).Times(2)
				},
			},
			args: args{
				timeout:  100 * time.Millisecond,
				taskName: "task1",
				taskFn: func() RetryableFn[any] {
					calls := 0
					return func(_ context.Context) (any, error) {
						calls++
						if calls < 3 {
							return 0, fmt.Errorf("task error %d", calls)
						}
						return 1, nil
					}
				}(),
				shouldRetryFn: func(_ error) bool {
					return true
				},
			},
			want: want{
				got:     1,
				wantErr: func(t testing.TB, err error) { require.NoError(t, err) },
				wantLogEvents: func(t testing.TB, events log.TestLogEventSet) {
					assert.Len(t, events, 2)
				},
			},
		},
		"task fails three times, final error returned": {
			fields: fields{
				expectStrategy: func(m *retrymgrtest.MockStrategy) {
					gomock.InOrder(
						m.EXPECT().NextDelay().Return(10*time.Millisecond, true).Times(3),
						m.EXPECT().NextDelay().Return(time.Duration(0), false).Times(1),
					)
				},
				expectSleeper: func(m *retrymgrtest.MockSleeper) {
					m.EXPECT().Sleep(gomock.Any(), 10*time.Millisecond).Return(nil).Times(3)
				},
			},
			args: args{
				timeout:  100 * time.Millisecond,
				taskName: "task1",
				taskFn: func() RetryableFn[any] {
					calls := 0
					return func(_ context.Context) (any, error) {
						calls++
						if calls < 10 {
							return 0, fmt.Errorf("task error %d", calls)
						}
						return 1, nil
					}
				}(),
				shouldRetryFn: func(_ error) bool {
					return true
				},
			},
			want: want{
				got:     0,
				wantErr: func(t testing.TB, err error) { require.Errorf(t, err, "task error 3") },
				wantLogEvents: func(t testing.TB, events log.TestLogEventSet) {
					assert.Len(t, events, 3)
				},
			},
		},
		"task is not retriable after two times, final error returned": {
			fields: fields{
				expectStrategy: func(m *retrymgrtest.MockStrategy) {
					m.EXPECT().NextDelay().Return(10*time.Millisecond, true).Times(2)
				},
				expectSleeper: func(m *retrymgrtest.MockSleeper) {
					m.EXPECT().Sleep(gomock.Any(), 10*time.Millisecond).Return(nil).Times(2)
				},
			},
			args: args{
				timeout:  100 * time.Millisecond,
				taskName: "task1",
				taskFn: func() RetryableFn[any] {
					calls := 0
					return func(_ context.Context) (any, error) {
						calls++
						if calls < 10 {
							return 0, fmt.Errorf("task error %d", calls)
						}
						return 1, nil
					}
				}(),
				shouldRetryFn: func() ShouldRetryFn {
					calls := 0
					return func(_ error) bool {
						calls++
						return calls < 3
					}
				}(),
			},
			want: want{
				got:     0,
				wantErr: func(t testing.TB, err error) { require.Errorf(t, err, "task error 2") },
				wantLogEvents: func(t testing.TB, events log.TestLogEventSet) {
					assert.Len(t, events, 2)
				},
			},
		},
		"delay is too long after two retries, context canceled": {
			fields: fields{
				expectStrategy: func(m *retrymgrtest.MockStrategy) {
					gomock.InOrder(
						m.EXPECT().NextDelay().Return(10*time.Millisecond, true).Times(1),
						m.EXPECT().NextDelay().Return(30*time.Millisecond, true).Times(1),
						m.EXPECT().NextDelay().Return(50*time.Millisecond, true).Times(1),
					)
				},
				expectSleeper: func(m *retrymgrtest.MockSleeper) {
					gomock.InOrder(
						m.EXPECT().Sleep(gomock.Any(), 10*time.Millisecond).Return(nil).Times(1),
						m.EXPECT().Sleep(gomock.Any(), 30*time.Millisecond).Return(nil).Times(1),
						m.EXPECT().Sleep(gomock.Any(), 50*time.Millisecond).Return(fmt.Errorf("context canceled")).Times(1),
					)
				},
			},
			args: args{
				timeout:  100 * time.Millisecond,
				taskName: "task1",
				taskFn: func() RetryableFn[any] {
					calls := 0
					return func(_ context.Context) (any, error) {
						calls++
						if calls < 10 {
							return 0, fmt.Errorf("task error %d", calls)
						}
						return 1, nil
					}
				}(),
				shouldRetryFn: func() ShouldRetryFn {
					calls := 0
					return func(_ error) bool {
						calls++
						return calls < 10
					}
				}(),
			},
			want: want{
				got: 0,
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "task error 2")
					require.Errorf(t, err, "context canceled")
				},
				wantLogEvents: func(t testing.TB, events log.TestLogEventSet) {
					assert.Len(t, events, 4)
					assert.NotEmpty(t, events.FindMatchingEvents(log.LevelInfo, "aborting execution due to sleeper error"))
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := log.NewTestLogger()
			strategy := retrymgrtest.NewMockStrategy(ctrl)
			strategy.EXPECT().Name().Return("mock_strategy")
			tt.fields.expectStrategy(strategy)
			sleeper := retrymgrtest.NewMockSleeper(ctrl)
			tt.fields.expectSleeper(sleeper)

			f := NewRetrierFactory(logger, sleeper, func() Strategy { return strategy })
			r := NewRetrier[any](f)

			ctx, cancel := context.WithTimeout(t.Context(), tt.args.timeout)
			defer cancel()

			// Act
			got, err := r.Do(ctx, tt.args.taskName, tt.args.taskFn, tt.args.shouldRetryFn)

			// Assert
			tt.want.wantErr(t, err)
			tt.want.wantLogEvents(t, logger.RecordedEvents())
			assert.Equal(t, tt.want.got, got)
		})
	}
}
