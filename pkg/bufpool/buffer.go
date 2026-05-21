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

// Buffer is a thin wrapper on top of [bytes.Buffer] to enable release to [sync.Pool]
// when [Buffer.Close] method is called.
type Buffer struct {
	buf       *bytes.Buffer
	releaseFn func()
	once      sync.Once
}

// Write calls the underlying [bytes.Buffer.Write].
func (b *Buffer) Write(p []byte) (int, error) {
	return b.buf.Write(p)
}

// Read calls the underlying [bytes.Buffer.Read].
func (b *Buffer) Read(p []byte) (int, error) {
	return b.buf.Read(p)
}

// Reset calls the underlying [bytes.Buffer.Reset].
func (b *Buffer) Reset() {
	b.buf.Reset()
}

// Close implements release of the underlying [bytes.Buffer] to the [sync.Pool].
// It ensures that the release is performed only once via [sync.Once] primitive.
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

// Pool is a thin wrapper on top of [sync.Pool].
// Its purpose is to create [Buffer] object and supply then with proper
// release function.
type Pool struct {
	pool *sync.Pool
}

// New create an instance of [Pool].
func New() *Pool {
	return &Pool{
		pool: &sync.Pool{
			New: func() any {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

// Get returns a new instance of [Buffer] backed by a (possibly) reused [bytes.Buffer] from
// the underlying [sync.Pool].
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
