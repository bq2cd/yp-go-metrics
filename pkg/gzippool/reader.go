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
	// ErrGzipReaderUnavailable is returned by [ReaderPool.Get] if it cannot obtain
	// a gzip reader object from the underlying [sync.Pool].
	ErrGzipReaderUnavailable = errors.New("cannot get gzip reader from pool")
)

var _ io.ReadCloser = &Reader{}

// Reader is a thin wrapper on top of [gzip.Reader] to enable its release to a [sync.Pool]
// when [Reader.Close] method is called.
type Reader struct {
	rgz       *gzip.Reader
	releaseFn func()
	once      sync.Once
}

// Read calls the underlying [gzip.Reader.Read] method.
func (r *Reader) Read(p []byte) (int, error) {
	return r.rgz.Read(p)
}

// Close calls the underlying [gzip.Reader.Close] method and returns
// the underlying reader to a [sync.Pool].
// The release function is ensured to be called only once by using
// [sync.Once] primitive.
func (r *Reader) Close() error {
	var err error

	r.once.Do(func() {
		err = r.rgz.Close()

		r.releaseFn()
	})

	return err
}

// ReaderPool is a thing wrapper on top of [sync.Pool] to create reusable [gzip.Reader] objects
// that support release to the underlying pool.
type ReaderPool struct {
	pool *sync.Pool
}

// NewReaderPool creates an instance of [ReaderPool].
// It will return an (unlikely) error, if the initial seed reader
// (containing only gzip header and footer) cannot be created
// for some reason.
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

// Get returns a new instance of [Reader] backed by a (possibly) reused [gzip.Reader] from
// the underlying [sync.Pool].
// The returned [gzip.Reader] is configured with the provided [io.Reader] via [gzip.Reader.Reset] method.
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

// Read returns the contents of gzip header by using [bytes.NewReader].
func (r *noopReader) Read(p []byte) (int, error) {
	return bytes.NewReader(r.gzipHeader).Read(p)
}
