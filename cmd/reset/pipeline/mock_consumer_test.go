package pipeline_test

import (
	"context"
	"time"
)

// MockConsumer implements [pipeline.Consumer] interface.
// It records consumed items and optionally introduces delays and errors.
type MockConsumer[T comparable] struct {
	noopInitCloser

	consumed  *Counter[T]
	delay     time.Duration
	returnErr func(T) error
}

func NewMockConsumer[T comparable](delay time.Duration, returnErr func(T) error) *MockConsumer[T] {
	return &MockConsumer[T]{
		consumed:  NewCounter[T](),
		delay:     delay,
		returnErr: returnErr,
	}
}

func (m *MockConsumer[T]) Consume(ctx context.Context, item T) error {
	var err error

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.delay):
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if m.returnErr != nil {
			err = m.returnErr(item)
		}

		if err == nil {
			m.consumed.Add(item, 1)
		}

		return err
	}
}
