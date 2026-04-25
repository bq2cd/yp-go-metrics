package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/gzippool"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type compressorResponseWriter struct {
	http.ResponseWriter

	compressor      io.WriteCloser
	contentEncoding httpheaders.ContentEncoding
	_shouldCompress bool
}

// Write conditionally forwards write requests either to gzip compressor or
// to the normal HTTP response writer.
func (cw *compressorResponseWriter) Write(data []byte) (int, error) {
	if cw._shouldCompress {
		return cw.compressor.Write(data)
	}
	return cw.ResponseWriter.Write(data)
}

// WriteHeader determines whether response compression needs to be enabled
// based on the content type, and sets corresponding flag for [compressorResponseWriter.Write]
// method to check during writing.
func (cw *compressorResponseWriter) WriteHeader(statusCode int) {
	header := cw.Header()
	shouldCompress := false

	for _, ct := range []httpheaders.ContentType{
		httpheaders.ContentTypeApplicationJSON,
		httpheaders.ContentTypeTextHTML,
		httpheaders.ContentTypeTextPlain,
	} {
		if ct.Matches(header) {
			shouldCompress = true
			break
		}
	}
	if shouldCompress && statusCode < 300 {
		cw._shouldCompress = true
		cw.contentEncoding.Apply(header)
	}
	cw.ResponseWriter.WriteHeader(statusCode)
}

// Close flushes compressed data when the current response processing is completed.
// It does nothing, if compression is not enabled for the response.
func (cw *compressorResponseWriter) Close(l log.Logger) {
	if !cw._shouldCompress {
		// prevent compressor from writing gzip headers into writer when we should not compress
		return
	}
	err := cw.compressor.Close()
	if err != nil {
		l.Error().WithErr(err).Msg("compressor close failed")
	}
}

type compressorFactory interface {
	Create(w io.Writer) (io.WriteCloser, error)
	ContentEncoding() httpheaders.ContentEncoding
}

type compressorGzipFactory struct {
	level    int
	pool     *gzippool.WriterPool
	poolOnce sync.Once
}

// Create returns an instance of gzip writer from [gzippool.WriterPool].
// It may reuse previously allocated gzip writer if pool contains enough
// idle instances.
func (f *compressorGzipFactory) Create(w io.Writer) (io.WriteCloser, error) {
	var err error

	if f.pool == nil {
		f.poolOnce.Do(func() {
			f.pool, err = gzippool.NewWriterPool(f.level)
		})
		if err != nil {
			return nil, fmt.Errorf("cannot initialize gzip writer pool: %w", err)
		}
	}

	return f.pool.Get(w), nil
}

// ContentEncoding returns content encoding type for the compressed data.
// Currently, only `gzip` encoding is supported.
func (f *compressorGzipFactory) ContentEncoding() httpheaders.ContentEncoding {
	return httpheaders.ContentEncodingGzip
}

// Compressor defines middleware that performs compression of the responses based on their content type,
// and decompression of the requests.
func Compressor(l log.Logger) Middleware {
	m := &compressorMiddleware{
		logger:            l.With(log.Str("middleware", "compressor")),
		compressorFactory: &compressorGzipFactory{level: gzip.BestSpeed},
	}
	return createMiddleware(m)
}

type compressorMiddleware struct {
	logger            log.Logger
	compressorFactory compressorFactory
	decompressorPool  *gzippool.ReaderPool
	poolOnce          sync.Once
}

func (m *compressorMiddleware) decompressRequest(r *http.Request) error {
	if !httpheaders.ContentEncodingGzip.Matches(r.Header) {
		return nil
	}

	var err error

	if m.decompressorPool == nil {
		m.poolOnce.Do(func() {
			m.decompressorPool, err = gzippool.NewReaderPool()
		})
		if err != nil {
			return fmt.Errorf("cannot initialize gzip reader pool: %w", err)
		}
	}

	rbody := r.Body
	rgz, err := m.decompressorPool.Get(rbody)
	if err != nil {
		return err
	}

	r.Body = rgz

	return nil
}

func (m *compressorMiddleware) shouldCompress(r *http.Request) bool {
	return m.compressorFactory.ContentEncoding().Accepted(r.Header)
}

func (m *compressorMiddleware) wrapResponseWriter(w http.ResponseWriter) (*compressorResponseWriter, bool) {
	wgz, err := m.compressorFactory.Create(w)
	if err != nil {
		m.logger.Error().WithErr(err).Msg("compressed writer creation failed")
		return nil, false
	}

	cw := &compressorResponseWriter{
		ResponseWriter:  w,
		compressor:      wgz,
		contentEncoding: m.compressorFactory.ContentEncoding(),
	}

	return cw, true
}

// Intercept defines actual middleware implementation.
// It will call next HTTP handler after processing.
func (m *compressorMiddleware) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	if err := m.decompressRequest(r); err != nil {
		http.Error(w, "cannot decompress request", http.StatusInternalServerError)
		return
	}

	if !m.shouldCompress(r) {
		next.ServeHTTP(w, r)
		return
	}

	cw, ok := m.wrapResponseWriter(w)
	if !ok || cw == nil {
		next.ServeHTTP(w, r)
		return
	}
	defer cw.Close(m.logger)

	next.ServeHTTP(cw, r)
}
