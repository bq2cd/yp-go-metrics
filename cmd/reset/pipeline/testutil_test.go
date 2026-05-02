package pipeline_test

import (
	"context"
	"iter"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Counter is a thread-safe map with integer values; its primary purpose is
// to count occurrences of its keys.
type Counter[T comparable] struct {
	mu   sync.RWMutex
	data map[T]int
}

type noopInitCloser struct{}

// NewCounter creates an instance of [Counter].
func NewCounter[T comparable]() *Counter[T] {
	return &Counter[T]{data: make(map[T]int)}
}

// Add increments counter for given key with provided delta.
func (c *Counter[T]) Add(key T, delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] += delta
}

// AssertNumItemsInRange validates that number of items (keys) in [Counter] is between `minItems` and `maxItems`.
// Each item must have count equal to `1`.
// [assert.Assertions] from `testify` package are used for validation.
func (c *Counter[T]) AssertNumItemsInRange(t *testing.T, minItems, maxItems int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	assert.GreaterOrEqualf(t, len(c.data), minItems, "not enough items")
	assert.LessOrEqualf(t, len(c.data), maxItems, "too many items")

	for item, count := range c.data {
		assert.Equalf(t, 1, count, "item %v: counter must equal to 1", item)
	}
}

// GeneratorInt returns integers from 1 to N.
func GeneratorInt(limit int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := range limit {
			if !yield(i + 1) {
				return
			}
		}
	}
}

// DrainGenerator iterates over provided generator and sends values to a channel.
// It might return an error if provided context is canceled before the generator is exhausted.
func DrainGenerator[T any](ctx context.Context, generator iter.Seq[T], out chan<- T) error {
	defer close(out)

	for v := range generator {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- v:
			// send away!
		}
	}

	return nil
}

// WithErrorIfEqual returns a function that takes a value and returns provided error
// only if the value equals provided condition.
func WithErrorIfEqual[T comparable](cond T, err error) func(T) error {
	return func(v T) error {
		if v == cond {
			return err
		}

		return nil
	}
}

// Init is a no-op method required for implementation of [pipeline.InitCloser] interface.
func (s *noopInitCloser) Init(ctx context.Context) error {
	return nil
}

// Close is a no-op method required for implementation of [pipeline.StartStoppable] interface.
func (s *noopInitCloser) Close(ctx context.Context) error {
	return nil
}
