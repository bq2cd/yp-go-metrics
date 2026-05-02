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

func TestStep(t *testing.T) {
	testError := errors.New("step error")

	type testcase struct {
		processor        *MockProcessor[int]
		parallel         uint
		generator        iter.Seq[int]
		timeout          time.Duration
		wantProcessedMin int
		wantProcessedMax int
		wantErr          error
	}

	tests := map[string]testcase{
		"single thread step, fast, succeeds -- all processed": {
			processor:        NewMockProcessor[int](0, nil),
			parallel:         1,
			generator:        GeneratorInt(10),
			timeout:          15 * time.Millisecond,
			wantProcessedMin: 10,
			wantProcessedMax: 10,
			wantErr:          nil,
		},
		"single thread step, slow, succeeds -- part processed, deadline exceeded": {
			processor:        NewMockProcessor[int](10*time.Millisecond, nil),
			parallel:         1,
			generator:        GeneratorInt(10),
			timeout:          15 * time.Millisecond,
			wantProcessedMin: 1,
			wantProcessedMax: 1,
			wantErr:          context.DeadlineExceeded,
		},
		"single thread step, slow, fails -- nothing processed, error": {
			processor:        NewMockProcessor(10*time.Millisecond, WithErrorIfEqual(1, testError)),
			parallel:         1,
			generator:        GeneratorInt(10),
			timeout:          15 * time.Millisecond,
			wantProcessedMin: 0,
			wantProcessedMax: 0,
			wantErr:          testError,
		},
		"multi thread step, fast, succeeds -- all processed": {
			processor:        NewMockProcessor[int](0, nil),
			parallel:         5,
			generator:        GeneratorInt(10),
			timeout:          15 * time.Millisecond,
			wantProcessedMin: 10,
			wantProcessedMax: 10,
			wantErr:          nil,
		},
		"multi thread step, slow, succeeds -- part processed, deadline exceeded": {
			processor:        NewMockProcessor[int](10*time.Millisecond, nil),
			parallel:         5,
			generator:        GeneratorInt(10),
			timeout:          15 * time.Millisecond,
			wantProcessedMin: 0, // concurrency is hard to test
			wantProcessedMax: 5,
			wantErr:          context.DeadlineExceeded,
		},
		"multi thread step, fast, fails -- part processed, error": {
			processor:        NewMockProcessor(0, WithErrorIfEqual(5, testError)),
			parallel:         5,
			generator:        GeneratorInt(10),
			timeout:          30 * time.Millisecond,
			wantProcessedMin: 0, // concurrency is hard to test
			wantProcessedMax: 9,
			wantErr:          testError,
		},
		"multi thread step, slow, fails -- part processed, error": {
			processor:        NewMockProcessor(10*time.Millisecond, WithErrorIfEqual(4, testError)),
			parallel:         5,
			generator:        GeneratorInt(10),
			timeout:          30 * time.Millisecond,
			wantProcessedMin: 0, // concurrency is hard to test
			wantProcessedMax: 9,
			wantErr:          testError,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {

			ctx, cancel := context.WithTimeout(t.Context(), tc.timeout)
			defer cancel()

			ch := make(chan int, tc.parallel)

			step := pipeline.NewStep("step", ch, tc.processor, pipeline.Config{ParallelWorkers: tc.parallel})

			consumed := NewCounter[int]()

			grp := new(errgroup.Group)

			grp.Go(func() error {
				return DrainGenerator(ctx, tc.generator, ch)
			})

			grp.Go(func() error {
				return step.Run(ctx)
			})

			grp.Go(func() error {
				for item := range step.Out() {
					consumed.Add(item, 1)
				}

				return nil
			})

			err := grp.Wait()

			// Assert
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			tc.processor.processed.AssertNumItemsInRange(t, tc.wantProcessedMin, tc.wantProcessedMax)
		})
	}
}
