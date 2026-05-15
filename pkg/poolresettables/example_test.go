package poolresettables_test

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/bq2cd/yp-go-metrics/pkg/poolresettables"
)

func Example() {
	n := 5 // initial pool capacity

	// create a pool
	pool := poolresettables.New[bytes.Buffer](poolresettables.WithInitialCapacity(uint(n)))
	fmt.Printf("initial pool size: %d\n", pool.Size())

	// empty pool returns pointer to newly-allocated zero value (empty buffer)
	buf := pool.Get()
	fmt.Printf("first=%s,len=%d,cap=%d\n", buf.String(), buf.Len(), buf.Cap())

	// populate the pool with pre-allocated buffers
	for range n {
		pool.Put(bytes.NewBufferString("initial value is reset on Put")) // 29 bytes -> rounded to 32 bytes for buffer capacity
	}
	fmt.Printf("populated pool size: %d\n", pool.Size())

	// use the pool in goroutines
	wg := new(sync.WaitGroup)
	for i := range n * 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			buf := pool.Get()
			defer pool.Put(buf)

			beforeValue := buf.String()
			beforeLen := buf.Len()
			beforeCap := buf.Cap()

			fmt.Fprintf(buf, "%03d", i+1)

			// Avoid printing value after writing to the buffer to maintain predictable output.
			// Due to concurrency, different goroutines will pick pre-allocated buffers each time,
			// thus leading to different output (because capacity of pre-allocated buffers differ
			// from the capacity of newly allocated buffers).
			fmt.Printf("goroutine: before(value=%s,len=%d,cap=%d) -> after(len=%d,cap=%d)\n", beforeValue, beforeLen, beforeCap, buf.Len(), buf.Cap())

			// prevent goroutine from returning its buffer to the pool too fast
			// to make sure all buffers in the pool are used.
			time.Sleep(5 * time.Millisecond)
		}()
	}
	wg.Wait()

	// number idle buffers in the pool increased because we launched twice as many goroutines as
	// the number of the pre-allocated buffers.
	fmt.Printf("final pool size: %d\n", pool.Size())

	// drain pool
	for range pool.Size() {
		buf := pool.Get()
		fmt.Printf("draining: value=%s,len=%d,cap=%d\n", buf.String(), buf.Len(), buf.Cap())
	}
	fmt.Printf("drained pool size: %d\n", pool.Size())

	// Unordered output:
	// initial pool size: 0
	// first=,len=0,cap=0
	// populated pool size: 5
	// goroutine: before(value=,len=0,cap=32) -> after(len=3,cap=32)
	// goroutine: before(value=,len=0,cap=32) -> after(len=3,cap=32)
	// goroutine: before(value=,len=0,cap=32) -> after(len=3,cap=32)
	// goroutine: before(value=,len=0,cap=32) -> after(len=3,cap=32)
	// goroutine: before(value=,len=0,cap=32) -> after(len=3,cap=32)
	// goroutine: before(value=,len=0,cap=0) -> after(len=3,cap=64)
	// goroutine: before(value=,len=0,cap=0) -> after(len=3,cap=64)
	// goroutine: before(value=,len=0,cap=0) -> after(len=3,cap=64)
	// goroutine: before(value=,len=0,cap=0) -> after(len=3,cap=64)
	// goroutine: before(value=,len=0,cap=0) -> after(len=3,cap=64)
	// final pool size: 10
	// draining: value=,len=0,cap=32
	// draining: value=,len=0,cap=32
	// draining: value=,len=0,cap=32
	// draining: value=,len=0,cap=32
	// draining: value=,len=0,cap=32
	// draining: value=,len=0,cap=64
	// draining: value=,len=0,cap=64
	// draining: value=,len=0,cap=64
	// draining: value=,len=0,cap=64
	// draining: value=,len=0,cap=64
	// drained pool size: 0
}
