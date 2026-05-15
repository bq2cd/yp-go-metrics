// Package poolresettables provides a container (pool) for object implementing `interface { Reset() }`.
// Such objects can be put into the pool and have their `Reset()` method called automatically by the pool.
// When an object is taken from the pool, it should be in a ready-to-use state (on the condition that
// the object's `Reset()` method is implemented properly).
package poolresettables

import (
	"sync"

	"github.com/bq2cd/yp-go-metrics/pkg/option"
)

const (
	defaultInitialPoolCapacity = 16
)

// Resettable describe an object that can be "reset" to a clean state and reused.
// The object must have `Reset()` method implemented for a pointer receiver.
type Resettable[T any] interface {
	Reset()
	*T
}

// Pool is a container for storing and retrieving [Resettable] objects in a concurrently-safe manner.
// When stored in the [Pool], each object is "reset" to a clean state by calling its `Reset()` method.
// The object should be designed to be ready for use after zero value allocation (with `new()`);
// `bytes.Buffer` constitutes an example of such object.
// If object requires any additional initialization on top of zero value allocation, it is
// responsibility of the caller to either finish the object's initialization after zero value is
// received or pre-populate pool with properly initialized objects.
type Pool[T any, R Resettable[T]] struct {
	mu      sync.RWMutex
	objects []*T
}

type poolConfig struct {
	InitialPoolCapacity uint
}

// New creates an instance of a [Pool].
// Initial pool capacity can be configured by using [WithInitialCapacity] option.
func New[T any, R Resettable[T]](opts ...option.Option[poolConfig]) *Pool[T, R] {
	cfg := &poolConfig{
		InitialPoolCapacity: defaultInitialPoolCapacity,
	}

	option.Apply(cfg, opts...)

	return &Pool[T, R]{
		objects: make([]*T, 0, cfg.InitialPoolCapacity),
	}
}

// WithInitialCapacity configures [Pool] initial capacity.
func WithInitialCapacity(capacity uint) option.Option[poolConfig] {
	return func(c *poolConfig) {
		c.InitialPoolCapacity = capacity
	}
}

// Put stores provided [Resettable] object into the pool (`nil` objects are ignored).
// The object's `Reset()` method is called before the object is stored in the pool.
// The caller **must not** use the object directly after it has been [Put] into the pool.
func (p *Pool[T, R]) Put(obj R) {
	if obj == nil {
		return
	}

	// It is extremely unlikely that the same object can be [Put] into the pool from
	// multiple goroutines, so it should be safe to call [Reset] without mutex protection.
	// The object can implement its own concurrency safeguards internally, if that is needed.
	obj.Reset()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.objects = append(p.objects, obj)
}

// Get retrieves a previously "reset" object from the pool if the pool is not empty.
// Otherwise, it allocates a zero value (with `new()`) and returns a pointer to it.
// It is a responsibility of the caller to finish object's initialization if zero value is received
// (if such action is required).
// The caller is also responsible for returning the object to the pool after the work is finished,
// or discarding the object if it is no longer necessary.
func (p *Pool[T, R]) Get() *T {
	p.mu.Lock()
	defer p.mu.Unlock()

	last := len(p.objects) - 1
	if last < 0 {
		return new(T) // allocate zero value
	}

	obj := p.objects[last]
	p.objects = p.objects[:last]

	return obj
}

// Size returns number of idle (obtainable with [Pool.Get]) objects in the pool.
func (p *Pool[T, R]) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.objects)
}
