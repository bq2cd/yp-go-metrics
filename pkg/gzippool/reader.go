package gzippool

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	ErrGzipReaderUnavailable = errors.New("cannot get gzip reader from pool")
)

var _ io.ReadCloser = &Reader{}

type Reader struct {
	rgz       *gzip.Reader
	releaseFn func()
	once      sync.Once
}

func (r *Reader) Read(p []byte) (int, error) {
	return r.rgz.Read(p)
}

func (r *Reader) Close() error {
	var err error

	r.once.Do(func() {
		err = r.rgz.Close()

		r.releaseFn()
	})

	return err
}

type ReaderPool struct {
	pool *sync.Pool
}

func NewReaderPool() (*ReaderPool, error) {
	noop, err := newNoopReader()
	if err != nil {
		return nil, fmt.Errorf("cannot create noop gzip reader: %w", err)
	}

	pool := &ReaderPool{
		pool: &sync.Pool{
			New: func() any {
				// noop reader should produce valid gzip header, so no error is expected here,
				// but even if it happens, we'll check if rgz is `nil` in [ReaderPool.Get].
				rgz, _ := gzip.NewReader(noop)

				return rgz
			},
		},
	}

	return pool, nil
}

func (rp *ReaderPool) Get(r io.Reader) (*Reader, error) {
	rgz := rp.pool.Get().(*gzip.Reader)
	if rgz == nil {
		return nil, ErrGzipReaderUnavailable
	}

	err := rgz.Reset(r)
	if err != nil {
		return nil, fmt.Errorf("cannot reset gzip reader: %w", err)
	}

	wrapped := &Reader{
		rgz: rgz,
		releaseFn: func() {
			rp.pool.Put(rgz)
		},
	}

	return wrapped, nil
}

type noopReader struct {
	gzipHeader []byte
}

func newNoopReader() (*noopReader, error) {
	r := &noopReader{}

	buf := bytes.NewBuffer(nil)
	wgz := gzip.NewWriter(buf)

	_, err := wgz.Write(nil)
	if err != nil {
		return nil, fmt.Errorf("cannot gzip empty data into buffer (write error): %w", err)
	}

	err = wgz.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot gzip empty data into buffer (close error): %w", err)
	}

	r.gzipHeader = buf.Bytes()

	return r, nil
}

func (r *noopReader) Read(p []byte) (int, error) {
	return bytes.NewReader(r.gzipHeader).Read(p)
}
