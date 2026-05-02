package pipeline

import (
	"context"
)

// Consumer takes incoming data and possibly performs a final action on it.
type Consumer[I any] interface {
	InitCloser

	Consume(ctx context.Context, in I) error
}

// Sink describes a final pipeline stage that takes incoming data via channel, and passes it to a [Consumer].
// The consumer might perform any action on the data without returning any result (it might return an error, though).
type Sink[I any] struct {
	name     string
	config   Config
	in       <-chan I
	consumer Consumer[I]
}

// NewSink creates an instance of [Sink] with given `name` and [Consumer].
func NewSink[I any](name string, in <-chan I, consumer Consumer[I], config Config) *Sink[I] {
	return &Sink[I]{
		name:     name,
		config:   config,
		in:       in,
		consumer: consumer,
	}
}

// Run launches parallel processing of incoming channel's data.
func (s *Sink[I]) Run(ctx context.Context) error {
	return runStage(ctx, s.name, s.consumer, func(ctx context.Context) error {
		return runWorkers(ctx, s.config.ParallelWorkers, s.worker, nil)
	})
}

func (s *Sink[I]) worker(ctx context.Context) error {
	// use empty struct to pretend we return a result;
	// this will be discarded because we use `nil` output channel below.
	processorFn := func(ctx context.Context, item I) (struct{}, error) {
		err := s.consumer.Consume(ctx, item)

		return struct{}{}, err
	}

	// `nil` output channel will force [runProcessingLoop] to discard the result.
	return runProcessingLoop(ctx, s.name, s.in, nil, processorFn)
}
