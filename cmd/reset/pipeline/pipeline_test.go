package pipeline_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/cmd/reset/pipeline"
	"github.com/bq2cd/yp-go-metrics/cmd/reset/rungroup"
)

func TestPipeline(t *testing.T) {
	testError := errors.New("pipeline error")

	type testcase struct {
		producer        *MockProducer[int]
		processors      []*MockProcessor[int]
		consumer        *MockConsumer[int]
		parallel        uint
		timeout         time.Duration
		wantConsumedMin int
		wantConsumedMax int
		wantErr         error
	}

	tests := map[string]testcase{
		"single thread, source + sink, fast, no errors -- all processed": {
			producer:        NewMockProducer(GeneratorInt(10), 0, nil),
			processors:      []*MockProcessor[int]{},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        1,
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 10,
			wantConsumedMax: 10,
			wantErr:         nil,
		},
		"single thread, source + steps + sink, fast, no errors -- all processed": {
			producer: NewMockProducer(GeneratorInt(10), 0, nil),
			processors: []*MockProcessor[int]{
				NewMockProcessor[int](0, nil),
				NewMockProcessor[int](0, nil),
				NewMockProcessor[int](0, nil),
			},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        1,
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 10,
			wantConsumedMax: 10,
			wantErr:         nil,
		},
		"single thread, source + steps + sink, slow step, no errors -- part processed, deadline exceeded": {
			producer: NewMockProducer(GeneratorInt(10), 0, nil),
			processors: []*MockProcessor[int]{
				NewMockProcessor[int](0, nil),
				NewMockProcessor[int](10*time.Millisecond, nil),
				NewMockProcessor[int](0, nil),
			},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        1,
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 1,
			wantConsumedMax: 1,
			wantErr:         context.DeadlineExceeded,
		},
		"single thread, source + steps + sink, slow step, fails -- nothing processed, error": {
			producer: NewMockProducer(GeneratorInt(10), 0, nil),
			processors: []*MockProcessor[int]{
				NewMockProcessor[int](0, nil),
				NewMockProcessor[int](10*time.Millisecond, nil),
				NewMockProcessor(0, WithErrorIfEqual(1, testError)),
			},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        1,
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 0,
			wantConsumedMax: 0,
			wantErr:         testError,
		},
		"multi thread, source + sink, fast, no errors -- all processed": {
			producer:        NewMockProducer(GeneratorInt(10), 0, nil),
			processors:      []*MockProcessor[int]{},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        5,
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 10,
			wantConsumedMax: 10,
			wantErr:         nil,
		},
		"multi thread, source + steps + sink, fast, no errors -- all processed": {
			producer: NewMockProducer(GeneratorInt(10), 0, nil),
			processors: []*MockProcessor[int]{
				NewMockProcessor[int](0, nil),
				NewMockProcessor[int](0, nil),
				NewMockProcessor[int](0, nil),
			},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        5,
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 10,
			wantConsumedMax: 10,
			wantErr:         nil,
		},
		"multi thread, source + steps + sink, slow step, no errors -- part processed, deadline exceeded": {
			producer: NewMockProducer(GeneratorInt(10), 0, nil),
			processors: []*MockProcessor[int]{
				NewMockProcessor[int](0, nil),
				NewMockProcessor[int](10*time.Millisecond, nil),
				NewMockProcessor[int](0, nil),
			},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        5,
			timeout:         15 * time.Millisecond,
			wantConsumedMin: 5,
			wantConsumedMax: 5,
			wantErr:         context.DeadlineExceeded,
		},
		"multi thread, source + steps + sink, fast, fails -- part processed, error": {
			producer: NewMockProducer(GeneratorInt(10), 0, nil),
			processors: []*MockProcessor[int]{
				NewMockProcessor[int](0, nil),
				NewMockProcessor(0, WithErrorIfEqual(5, testError)),
				NewMockProcessor[int](0, nil),
			},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        5,
			timeout:         30 * time.Millisecond,
			wantConsumedMin: 0, // concurrency is hard to test
			wantConsumedMax: 9,
			wantErr:         testError,
		},
		"multi thread, source + steps + sink, slow step, fails -- part processed, error": {
			producer: NewMockProducer(GeneratorInt(10), 0, nil),
			processors: []*MockProcessor[int]{
				NewMockProcessor[int](10*time.Millisecond, nil),
				NewMockProcessor(0, WithErrorIfEqual(4, testError)),
				NewMockProcessor(0, WithErrorIfEqual(6, testError)),
			},
			consumer:        NewMockConsumer[int](0, nil),
			parallel:        5,
			timeout:         30 * time.Millisecond,
			wantConsumedMin: 0, // concurrency is hard to test
			wantConsumedMax: 8,
			wantErr:         testError,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			baseCtx, baseCancel := context.WithTimeout(t.Context(), tc.timeout)
			defer baseCancel()

			source := pipeline.NewSource("source", tc.producer, pipeline.Config{ParallelWorkers: tc.parallel})

			ch := source.Out()

			steps := make([]*pipeline.Step[int, int], len(tc.processors))
			for i := range steps {
				name := fmt.Sprintf("step %d", i+1)
				steps[i] = pipeline.NewStep(name, ch, tc.processors[i], pipeline.Config{ParallelWorkers: tc.parallel})
				ch = steps[i].Out()
			}

			sink := pipeline.NewSink("sink", ch, tc.consumer, pipeline.Config{ParallelWorkers: tc.parallel})

			grp, ctx := rungroup.New(baseCtx)

			grp.Go(ctx, source, sink)
			for i := range steps {
				grp.Go(ctx, steps[i])
			}

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
