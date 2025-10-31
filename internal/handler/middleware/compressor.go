package middleware

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type compressorResponseWriter struct {
	http.ResponseWriter
	compressor      io.WriteCloser
	contentEncoding httpheaders.ContentEncoding
	_shouldCompress bool
}

func (cw *compressorResponseWriter) Write(data []byte) (int, error) {
	if cw._shouldCompress {
		return cw.compressor.Write(data)
	}
	return cw.ResponseWriter.Write(data)
}

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

func (cw *compressorResponseWriter) Close(l log.Logger) {
	if !cw._shouldCompress {
		// prevent compressor from writing headers into writer when we should not compress
		return
	}
	err := cw.compressor.Close()
	if err != nil {
		l.Error().Err("error", err).Msg("compressor close failed")
	}
}

type compressorFactory interface {
	Create(w io.Writer) (io.WriteCloser, error)
	ContentEncoding() httpheaders.ContentEncoding
}

type compressorGzipFactory struct {
	level int
}

func (f *compressorGzipFactory) Create(w io.Writer) (io.WriteCloser, error) {
	wgz, err := gzip.NewWriterLevel(w, f.level)
	if err != nil {
		return nil, err
	}
	return wgz, nil
}

func (f *compressorGzipFactory) ContentEncoding() httpheaders.ContentEncoding {
	return httpheaders.ContentEncodingGzip
}

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
}

func (m *compressorMiddleware) decompressRequest(r *http.Request) error {
	if !httpheaders.ContentEncodingGzip.Matches(r.Header) {
		return nil
	}
	rbody := r.Body
	rgz, err := gzip.NewReader(rbody)
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
		m.logger.Error().Err("error", err).Msg("compressed writer creation failed")
		return nil, false
	}

	cw := &compressorResponseWriter{
		ResponseWriter:  w,
		compressor:      wgz,
		contentEncoding: m.compressorFactory.ContentEncoding(),
	}

	return cw, true
}

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
