package rungroup_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/cmd/reset/rungroup"
)

type mockRunnable struct {
	started, finished atomic.Int64
	delay             time.Duration
	returnErr         error
}

func (m *mockRunnable) Run(ctx context.Context) error {
	m.started.Add(1)

	defer m.finished.Add(1)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.delay):
		return m.returnErr
	}
}

func TestGoAndWait(t *testing.T) {
	type testcase struct {
		runnables []*mockRunnable
		timeout   time.Duration
		wantErr   error
	}

	tests := map[string]testcase{
		"no runnables -> immediate completion": {
			runnables: []*mockRunnable{},
			timeout:   10 * time.Millisecond,
			wantErr:   nil,
		},
		"single runnable, no delay -> no error": {
			runnables: []*mockRunnable{
				{},
			},
			timeout: 10 * time.Millisecond,
			wantErr: nil,
		},
		"single runnable, fails -> error": func() testcase {
			err := errors.New("oops")

			return testcase{
				runnables: []*mockRunnable{
					{returnErr: err},
				},
				timeout: 10 * time.Millisecond,
				wantErr: err,
			}
		}(),
		"single runnable, with delay -> deadline exceeded": {
			runnables: []*mockRunnable{
				{delay: 20 * time.Millisecond},
			},
			timeout: 10 * time.Millisecond,
			wantErr: context.DeadlineExceeded,
		},
		"multiple runnables, no delay -> no error": {
			runnables: []*mockRunnable{
				{}, {}, {}, {}, {},
			},
			timeout: 10 * time.Millisecond,
			wantErr: nil,
		},
		"multiple runnables, delay, one fails -> all terminated, error returned": func() testcase {
			err := errors.New("3rd fails")

			return testcase{
				runnables: []*mockRunnable{
					{delay: 20 * time.Millisecond},
					{delay: 15 * time.Millisecond},
					{delay: 5 * time.Millisecond, returnErr: err},
					{delay: 10 * time.Millisecond},
					{delay: 1 * time.Millisecond},
				},
				timeout: 10 * time.Millisecond,
				wantErr: err,
			}
		}(),
		"multiple runnables, some too slow -> deadline exceeded": {
			runnables: []*mockRunnable{
				{delay: 20 * time.Millisecond},
				{delay: 15 * time.Millisecond},
				{delay: 25 * time.Millisecond},
				{delay: 10 * time.Millisecond},
				{delay: 5 * time.Millisecond},
			},
			timeout: 10 * time.Millisecond,
			wantErr: context.DeadlineExceeded,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			baseCtx, baseCancel := context.WithTimeout(t.Context(), tc.timeout)
			defer baseCancel()

			grp, ctx := rungroup.New(baseCtx)

			grp.Go(ctx, toRunnable(tc.runnables)...)

			err := grp.Wait()

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			for _, r := range tc.runnables {
				assert.Equalf(t, int64(1), r.started.Load(), "%+v has not started", r)
				assert.Equalf(t, int64(1), r.finished.Load(), "%+v has not finished", r)
			}
		})
	}
}

func toRunnable(input []*mockRunnable) []rungroup.Runnable {
	output := make([]rungroup.Runnable, len(input))

	for i := range input {
		output[i] = input[i]
	}

	return output
}
