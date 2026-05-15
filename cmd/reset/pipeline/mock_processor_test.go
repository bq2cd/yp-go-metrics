package pipeline_test

import (
	"context"
	"time"
)

// MockProcessor implements [pipeline.Processor] interface.
// It takes incoming items and returns them as is, while optionally adding delays and errors.
type MockProcessor[T comparable] struct {
	noopInitCloser

	processed *Counter[T]
	delay     time.Duration
	returnErr func(T) error
}

func NewMockProcessor[T comparable](delay time.Duration, returnErr func(T) error) *MockProcessor[T] {
	return &MockProcessor[T]{
		processed: NewCounter[T](),
		delay:     delay,
		returnErr: returnErr,
	}
}

func (m *MockProcessor[T]) Process(ctx context.Context, item T) (T, error) {
	var (
		zero T
		err  error
	)

	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-time.After(m.delay):
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		if m.returnErr != nil {
			err = m.returnErr(item)
		}

		if err == nil {
			m.processed.Add(item, 1)
		}

		return item, err
	}
}
