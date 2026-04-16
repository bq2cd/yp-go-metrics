package bufpool

import (
	"bytes"
	"io"
	"sync"
)

var (
	_ io.Writer     = &Buffer{}
	_ io.ReadCloser = &Buffer{}
)

type Buffer struct {
	buf       *bytes.Buffer
	releaseFn func()
	once      sync.Once
}

func (b *Buffer) Write(p []byte) (int, error) {
	return b.buf.Write(p)
}

func (b *Buffer) Read(p []byte) (int, error) {
	return b.buf.Read(p)
}

func (b *Buffer) Close() error {
	b.once.Do(b.releaseFn)

	return nil
}

// Bytes returns internal buffer slice for direct (read) access.
// To obtain the slice, it calls [bytes.Buffer.Bytes] method under the hood.
// The slice is valid for use only until the next buffer modification
// (see documentation of [bytes.Buffer.Bytes] for more details).
func (b *Buffer) Bytes() []byte {
	return b.buf.Bytes()
}

// Reader wraps internal buffer slice into [bytes.NewReader].
// To obtain the slice, it calls [bytes.Buffer.Bytes] method under the hood.
// The slice is valid for use only until the next buffer modification
// (see documentation of [bytes.Buffer.Bytes] for more details).
func (b *Buffer) Reader() io.Reader {
	return bytes.NewReader(b.buf.Bytes())
}

type Pool struct {
	pool *sync.Pool
}

func New() *Pool {
	return &Pool{
		pool: &sync.Pool{
			New: func() any {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

func (p *Pool) Get() *Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset()

	return &Buffer{
		buf: buf,
		releaseFn: func() {
			p.pool.Put(buf)
		},
	}
}
