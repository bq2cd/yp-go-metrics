package pipeline_test

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/cmd/reset/pipeline"
)

func TestSink(t *testing.T) {
	testError := errors.New("sink error")

	type testcase struct {
		consumer        *MockConsumer[int]
		parallel        uint
		generator       iter.Seq[int]
		timeout         time.Duration
		wantConsumedMin int
		wantConsumedMax int
		wantErr         error
	}

	tests := map[string]testcase{
		"single thread sink, fast, succeeds -- all consumed": {
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        1,
			generator:       GeneratorInt(10),
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 10,
			wantConsumedMax: 10,
			wantErr:         nil,
		},
		"single thread sink, slow, succeeds -- part consumed, deadline exceeded": {
			consumer:        NewMockConsumer[int](10*time.Millisecond, nil),
			parallel:        1,
			generator:       GeneratorInt(10),
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 1,
			wantConsumedMax: 1,
			wantErr:         context.DeadlineExceeded,
		},
		"single thread sink, slow, fails -- nothing consumed, error": {
			consumer:        NewMockConsumer(10*time.Millisecond, WithErrorIfEqual(1, testError)),
			parallel:        1,
			generator:       GeneratorInt(10),
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 0,
			wantConsumedMax: 0,
			wantErr:         testError,
		},
		"multi thread sink, fast, succeeds -- all consumed": {
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        5,
			generator:       GeneratorInt(10),
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 10,
			wantConsumedMax: 10,
			wantErr:         nil,
		},
		"multi thread sink, slow, succeeds -- part consumed, deadline exceeded": {
			consumer:        NewMockConsumer[int](10*time.Millisecond, nil),
			parallel:        5,
			generator:       GeneratorInt(10),
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 0, // concurrency is hard to test
			wantConsumedMax: 5,
			wantErr:         context.DeadlineExceeded,
		},
		"multi thread sink, fast, fails -- part consumed, error": {
			consumer:        NewMockConsumer(0, WithErrorIfEqual(5, testError)),
			parallel:        5,
			generator:       GeneratorInt(10),
			timeout:         30 * time.Millisecond,
			wantConsumedMin: 0, // concurrency is hard to test
			wantConsumedMax: 9,
			wantErr:         testError,
		},
		"multi thread sink, slow, fails -- part consumed, error": {
			consumer:        NewMockConsumer(10*time.Millisecond, WithErrorIfEqual(4, testError)),
			parallel:        5,
			generator:       GeneratorInt(10),
			timeout:         30 * time.Millisecond,
			wantConsumedMin: 0, // concurrency is hard to test
			wantConsumedMax: 9,
			wantErr:         testError,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tc.timeout)
			defer cancel()

			ch := make(chan int, tc.parallel)

			sink := pipeline.NewSink("sink", ch, tc.consumer, pipeline.Config{ParallelWorkers: tc.parallel})

			grp := new(errgroup.Group)

			grp.Go(func() error {
				return DrainGenerator(ctx, tc.generator, ch)
			})

			grp.Go(func() error {
				return sink.Run(ctx)
			})

			err := grp.Wait()

			// Assert
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			tc.consumer.consumed.AssertNumItemsInRange(t, tc.wantConsumedMin, tc.wantConsumedMax)
		})
	}
}
