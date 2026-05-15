package poolresettables_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bq2cd/yp-go-metrics/pkg/poolresettables"
)

// Counter implements [poolresettables.Resettable] interface and serves an object for tests.
// It intentionally does not use any concurrency barriers to demonstrate that no data races occur
// when `Reset()` method is called from multiple goroutines.
type Counter struct {
	resets int
	value  int
}

func (c *Counter) Reset() {
	c.resets++
	c.value = 0
}

func TestEmptyPoolGet(t *testing.T) {
	pool := poolresettables.New[Counter]()

	c := pool.Get()

	assert.Equal(t, 0, c.resets)
	assert.Equal(t, 0, c.value)
}

func TestEmptyPoolGetAndPut(t *testing.T) {
	pool := poolresettables.New[Counter]()

	c := pool.Get()

	c.value = 25 // simulate work

	pool.Put(c)

	assert.Equal(t, 1, c.resets)
	assert.Equal(t, 0, c.value)
}

func TestPopulatedPoolGet(t *testing.T) {
	n := 5

	pool := poolresettables.New[Counter]()

	for i := range n {
		pool.Put(&Counter{value: i + 101})
	}

	assert.Equal(t, n, pool.Size())

	for range n {
		c := pool.Get()

		assert.Equal(t, 1, c.resets)
		assert.Equal(t, 0, c.value)
	}

	assert.Equal(t, 0, pool.Size())

	for range n {
		c := pool.Get()

		assert.Equal(t, 0, c.resets)
		assert.Equal(t, 0, c.value)
	}

	assert.Equal(t, 0, pool.Size())
}

func TestPopulatedPoolGetAndPutConcurrently(t *testing.T) {
	n := 10

	pool := poolresettables.New[Counter]()

	for i := range n {
		pool.Put(&Counter{value: i + 101})
	}

	assert.Equal(t, n, pool.Size())

	wg := new(sync.WaitGroup)

	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := pool.Get()
			defer pool.Put(c)

			c.value = i * 1000

			// make each goroutine hold its counter long enough to allow others to get
			// different counters from the pool, not the same one (if returned fast enough).
			time.Sleep(5 * time.Millisecond)
		}()
	}

	wg.Wait()

	for range n {
		c := pool.Get()

		assert.Equal(t, 2, c.resets)
		assert.Equal(t, 0, c.value)
	}
}
