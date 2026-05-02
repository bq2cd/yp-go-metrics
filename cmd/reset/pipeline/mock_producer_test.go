package pipeline_test

import (
	"context"
	"iter"
	"sync"
	"time"
)

// MockProducer implements [pipeline.Producer] interface.
// It uses "pull" iterator to produce items, records generated items, and
// optionally introduces delays and errors.
type MockProducer[T comparable] struct {
	noopInitCloser

	mu             sync.Mutex
	_generatorNext func() (T, bool)
	_generatorStop func()
	produced       *Counter[T]
	delay          time.Duration
	returnErr      func(T) error
}

func NewMockProducer[T comparable](generator iter.Seq[T], delay time.Duration, returnErr func(T) error) *MockProducer[T] {
	next, stop := iter.Pull(generator)

	return &MockProducer[T]{
		_generatorNext: next,
		_generatorStop: stop,
		produced:       NewCounter[T](),
		delay:          delay,
		returnErr:      returnErr,
	}
}

func (m *MockProducer[T]) Produce(ctx context.Context) (T, bool, error) {
	var (
		item T
		ok   bool
		err  error
	)

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-time.After(m.delay):
		err = ctx.Err()
		if err != nil {
			goto out
		}

		item, ok = m.nextItem()

		if m.returnErr != nil {
			err = m.returnErr(item)
		}

		if ok && err == nil {
			m.produced.Add(item, 1)
		}
	}

out:
	if err != nil {
		m.stopGenerator()
	}

	return item, ok, err
}

func (m *MockProducer[T]) nextItem() (T, bool) {
	// As per [iter.Pull] documentation:
	// "It is an error to call next or stop from multiple goroutines simultaneously".
	m.mu.Lock()
	defer m.mu.Unlock()

	return m._generatorNext()
}

func (m *MockProducer[T]) stopGenerator() {
	// As per [iter.Pull] documentation:
	// "It is an error to call next or stop from multiple goroutines simultaneously".
	m.mu.Lock()
	defer m.mu.Unlock()

	m._generatorStop()
}
