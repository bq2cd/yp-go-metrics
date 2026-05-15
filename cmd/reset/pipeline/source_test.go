package pipeline_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/cmd/reset/pipeline"
)

func TestSource(t *testing.T) {
	testError := errors.New("source error")

	type testcase struct {
		producer        *MockProducer[int]
		parallel        uint
		timeout         time.Duration
		wantProducedMin int
		wantProducedMax int
		wantErr         error
	}

	tests := map[string]testcase{
		"single thread source, fast, succeeds -- all generated": {
			producer:        NewMockProducer(GeneratorInt(10), 0, nil),
			parallel:        1,
			timeout:         15 * time.Millisecond,
			wantProducedMin: 10,
			wantProducedMax: 10,
			wantErr:         nil,
		},
		"single thread source, slow, succeeds -- part generated, deadline exceeded": {
			producer:        NewMockProducer(GeneratorInt(10), 10*time.Millisecond, nil),
			parallel:        1,
			timeout:         15 * time.Millisecond,
			wantProducedMin: 1,
			wantProducedMax: 1,
			wantErr:         context.DeadlineExceeded,
		},
		"single thread source, slow, fails -- nothing generated, error": {
			producer:        NewMockProducer(GeneratorInt(10), 10*time.Millisecond, WithErrorIfEqual(1, testError)),
			parallel:        1,
			timeout:         15 * time.Millisecond,
			wantProducedMin: 0,
			wantProducedMax: 0,
			wantErr:         testError,
		},
		"multi thread source, fast, succeeds -- all generated": {
			producer:        NewMockProducer(GeneratorInt(10), 0, nil),
			parallel:        5,
			timeout:         15 * time.Millisecond,
			wantProducedMin: 10,
			wantProducedMax: 10,
			wantErr:         nil,
		},
		"multi thread source, slow, succeeds -- part generated, deadline exceeded": {
			producer:        NewMockProducer(GeneratorInt(10), 10*time.Millisecond, nil),
			parallel:        5,
			timeout:         15 * time.Millisecond,
			wantProducedMin: 0, // concurrency is hard to test
			wantProducedMax: 5,
			wantErr:         context.DeadlineExceeded,
		},
		"multi thread source, fast, fails -- part generated, error": {
			producer:        NewMockProducer(GeneratorInt(10), 0, WithErrorIfEqual(5, testError)),
			parallel:        5,
			timeout:         30 * time.Millisecond,
			wantProducedMin: 0, // concurrency is hard to test
			wantProducedMax: 9,
			wantErr:         testError,
		},
		"multi thread source, slow, fails -- part generated, error": {
			producer:        NewMockProducer(GeneratorInt(10), 10*time.Millisecond, WithErrorIfEqual(4, testError)),
			parallel:        5,
			timeout:         30 * time.Millisecond,
			wantProducedMin: 0, // concurrency is hard to test
			wantProducedMax: 9,
			wantErr:         testError,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tc.timeout)
			defer cancel()

			source := pipeline.NewSource("source", tc.producer, pipeline.Config{ParallelWorkers: tc.parallel})

			consumed := NewCounter[int]()

			grp := new(errgroup.Group)

			grp.Go(func() error {
				return source.Run(ctx)
			})

			grp.Go(func() error {
				for item := range source.Out() {
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

			consumed.AssertNumItemsInRange(t, tc.wantProducedMin, tc.wantProducedMax)

			// producer can generate more items that won't be sent to outgoing channel due to context cancellation.
			tc.producer.produced.AssertNumItemsInRange(t, tc.wantProducedMin, tc.wantProducedMax+1)
		})
	}
}
