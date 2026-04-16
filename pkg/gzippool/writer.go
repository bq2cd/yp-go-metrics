package gzippool

import (
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

var _ io.WriteCloser = &Writer{}

type Writer struct {
	wgz       *gzip.Writer
	releaseFn func()
	once      sync.Once
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.wgz.Write(p)
}

func (w *Writer) Close() error {
	var err error

	w.once.Do(func() {
		err = w.wgz.Close()

		w.releaseFn()
	})

	return err
}

type WriterPool struct {
	pool *sync.Pool
}

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
