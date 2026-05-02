package pipeline

import (
	"context"
)

// Processor performs an action on incoming data and return a result or error.
type Processor[I, O any] interface {
	InitCloser

	Process(ctx context.Context, in I) (O, error)
}

// Step describes pipeline stage that takes incoming data via channel, applies provided [Processor] to it,
// then sends the result to outgoing channel.
// The stage aborts its execution if [Processor] returns an error.
type Step[I, O any] struct {
	name      string
	config    Config
	in        <-chan I
	out       chan O
	processor Processor[I, O]
}

// NewStep creates an instance of [Step] with given `name` and [Processor].
func NewStep[I, O any](name string, in <-chan I, processor Processor[I, O], config Config) *Step[I, O] {
	return &Step[I, O]{
		name:      name,
		config:    config,
		in:        in,
		out:       make(chan O, config.OutputBufferSize),
		processor: processor,
	}
}

// Out returns outgoing channel of the [Step] for reading.
func (s *Step[I, O]) Out() <-chan O {
	return s.out
}

// Run launches parallel processing of incoming channel's data.
func (s *Step[I, O]) Run(ctx context.Context) error {
	return runStage(ctx, s.name, s.processor, func(ctx context.Context) error {
		return runWorkers(ctx, s.config.ParallelWorkers, s.worker, s.cleanup)
	})
}

func (s *Step[I, O]) worker(ctx context.Context) error {
	return runProcessingLoop(ctx, s.name, s.in, s.out, s.processor.Process)
}

func (s *Step[I, O]) cleanup() {
	close(s.out)
}
