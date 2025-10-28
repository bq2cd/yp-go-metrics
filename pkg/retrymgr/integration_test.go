package retrymgr_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testLogger interface {
	Logf(format string, args ...any)
}

type taskInput struct {
	logger      testLogger
	callCounter int
	data        []string
	workDelay   time.Duration
	maxErrs     int
}

type taskOutput struct {
	output     []string
	totalCalls int
}

func testTask(ctx context.Context, input *taskInput) (output *taskOutput, err error) {
	input.callCounter++
	input.logger.Logf("[testTask] called with input %v", input)

	output = &taskOutput{
		output:     input.data,
		totalCalls: input.callCounter,
	}

	if input.callCounter <= input.maxErrs {
		input.logger.Logf("[testTask] simulating error for call %d", input.callCounter)
		err = fmt.Errorf("task error %d", input.callCounter)
		return
	}

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-time.After(input.workDelay):
	}

	return
}

func testTaskExecutor(t *testing.T, ctx context.Context, factory retrymgr.RetrierFactory, maxErrs int) (*taskOutput, error) {
	input := &taskInput{
		logger:  t,
		data:    []string{"item1", "item2"},
		maxErrs: maxErrs,
	}
	res, err := retrymgr.NewRetrier[*taskOutput](factory).Do(
		ctx, "task1",
		func(ctx context.Context) (*taskOutput, error) {
			res, err := testTask(ctx, input)
			if err != nil {
				t.Logf("[retry] task1 returned error: %v", err)
			}
			return res, err
		},
		func(err error) bool {
			return true
		},
	)
	t.Logf("task1 completed: res=%v, err=%v", res, err)
	return res, err
}

func runTestTaskExecutor(t *testing.T, ctrl *gomock.Controller, maxErrs int, timeout time.Duration) (*taskOutput, error) {
	t.Helper()

	logger := log.NewTestLogger()

	newStrategyFn := func() retrymgr.Strategy {
		s := retrymgrtest.NewMockStrategy(ctrl)
		s.EXPECT().Name().Return("mock_strategy").Times(1)
		calls := []any{}
		for i, d := range []int{10, 30, 50, 0} {
			c := s.EXPECT().NextDelay().Return(time.Duration(d)*time.Millisecond, d > 0)
			if i+1 <= maxErrs {
				c.Times(1)
			} else {
				c.Times(0)
			}
			calls = append(calls, c)
		}
		gomock.InOrder(calls...)
		return s
	}

	ctx := t.Context()
	if timeout > 0 {
		ctxT, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = ctxT
	}

	factory := retrymgr.NewRetrierFactory(logger, retrymgr.NewSleeper(), newStrategyFn)

	return testTaskExecutor(t, ctx, factory, maxErrs)
}

func TestExampleUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type testcase struct {
		maxErrs   int
		timeout   time.Duration
		wantCalls int
		wantErr   func(testing.TB, error)
	}

	tests := map[string][]testcase{
		"no deadline imposed": {
			{maxErrs: 0, timeout: 0, wantCalls: 1, wantErr: func(t testing.TB, err error) { require.NoError(t, err) }},
			{maxErrs: 1, timeout: 0, wantCalls: 2, wantErr: func(t testing.TB, err error) { require.NoError(t, err) }},
			{maxErrs: 2, timeout: 0, wantCalls: 3, wantErr: func(t testing.TB, err error) { require.NoError(t, err) }},
			{maxErrs: 3, timeout: 0, wantCalls: 4, wantErr: func(t testing.TB, err error) { require.NoError(t, err) }},
			{maxErrs: 4, timeout: 0, wantCalls: 4, wantErr: func(t testing.TB, err error) { require.Errorf(t, err, "task error 4") }},
			{maxErrs: 5, timeout: 0, wantCalls: 4, wantErr: func(t testing.TB, err error) { require.Errorf(t, err, "task error 4") }},
		},
		"with deadline": {
			{maxErrs: 2, timeout: 20 * time.Millisecond, wantCalls: 2, wantErr: func(t testing.TB, err error) { require.Errorf(t, err, "context deadline exceeded") }},
			{maxErrs: 3, timeout: 50 * time.Millisecond, wantCalls: 3, wantErr: func(t testing.TB, err error) { require.Errorf(t, err, "context deadline exceeded") }},
		},
	}

	for gname, group := range tests {
		t.Run(gname, func(t *testing.T) {
			for _, tt := range group {
				t.Run(fmt.Sprintf("maxErrs=%d,timeout=%v", tt.maxErrs, tt.timeout), func(t *testing.T) {
					got, err := runTestTaskExecutor(t, ctrl, tt.maxErrs, tt.timeout)
					assert.Equal(t, tt.wantCalls, got.totalCalls)
					tt.wantErr(t, err)
				})
			}
		})
	}
}
