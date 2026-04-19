package gzippool

import (
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

var _ io.WriteCloser = &Writer{}

// Writer is thin wrapper on top of [gzip.Writer] to enable its release to a [sync.Pool]
// when [Writer.Close] method is called.
type Writer struct {
	wgz       *gzip.Writer
	releaseFn func()
	once      sync.Once
}

// Write calls the underlying [gzip.Writer.Write] method.
func (w *Writer) Write(p []byte) (int, error) {
	return w.wgz.Write(p)
}

// Close calls the underlying [gzip.Writer.Close] method and returns the underlying writer
// to a [sync.Pool].
// The release function is ensured to be called once by using [sync.Once] primitive.
func (w *Writer) Close() error {
	var err error

	w.once.Do(func() {
		err = w.wgz.Close()

		w.releaseFn()
	})

	return err
}

// WriterPool is a thin wrapper on top of [sync.Pool] to create reusable [gzip.Writer] objects
// that support release to the underlying pool.
type WriterPool struct {
	pool *sync.Pool
}

// NewWriterPool creates an instance of [WriterPool] for a given gzip level.
// If the provided level is invalid, an error will be returned.
func NewWriterPool(level int) (*WriterPool, error) {
	_, err := gzip.NewWriterLevel(io.Discard, level)
	if err != nil {
		return nil, fmt.Errorf("invalid gzip level: %w", err)
	}

	pool := &WriterPool{
		pool: &sync.Pool{
			New: func() any {
				// we already validated level, so can ignore error here
				wgz, _ := gzip.NewWriterLevel(io.Discard, level)

				return wgz
			},
		},
	}

	return pool, nil
}

// Get returns a new instance of [Writer] backed by a (possibly) reused [gzip.Writer] from
// the underlying [sync.Pool].
// The returned [gzip.Writer] is configured with the provided [io.Writer] via [gzip.Writer.Reset] method.
func (wp *WriterPool) Get(w io.Writer) *Writer {
	wgz := wp.pool.Get().(*gzip.Writer)
	wgz.Reset(w)

	return &Writer{
		wgz: wgz,
		releaseFn: func() {
			wp.pool.Put(wgz)
		},
	}
}
