package pipeline

import (
	"context"
	"fmt"
)

// Producer generates arbitrary data.
// Its [Producer.Produce] method is called concurrently and expected to return
// a single data item, completion flag, and, potentially, a generation error.
// When generation has been finished, the producer will return `false` as the completion flag,
// zero value for the data item, and no error.
// The producer is expected to handle context cancellation properly and
// stop data generation as soon as possible.
type Producer[O any] interface {
	InitCloser

	Produce(ctx context.Context) (O, bool, error)
}

// Source describe an initial pipeline stage that generates data and sends it via outgoing channel
// for consumption by other stages ([Step], [Sink]).
type Source[O any] struct {
	name     string
	config   Config
	out      chan O
	producer Producer[O]
}

// NewSource creates an instance of [Source] with given `name` and [Producer].
func NewSource[O any](name string, producer Producer[O], config Config) *Source[O] {
	return &Source[O]{
		name:     name,
		config:   config,
		out:      make(chan O, config.OutputBufferSize),
		producer: producer,
	}
}

// Out returns outgoing channel of the [Source] for reading.
func (s *Source[O]) Out() <-chan O {
	return s.out
}

// Run launches parallel generation of the data.
func (s *Source[O]) Run(ctx context.Context) error {
	return runStage(ctx, s.name, s.producer, func(ctx context.Context) error {
		return runWorkers(ctx, s.config.ParallelWorkers, s.worker, s.cleanup)
	})
}

func (s *Source[O]) worker(ctx context.Context) error {
loop:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s aborted: %w", s.name, ctx.Err())
		default:
		}

		item, ok, err := s.producer.Produce(ctx)
		if err != nil { // generation error
			return err
		}

		if !ok { // producer finished generation
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("%s aborted: %w", s.name, ctx.Err())
		case s.out <- item:
			continue loop
		}
	}
}

func (s *Source[O]) cleanup() {
	close(s.out)
}
