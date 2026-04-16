package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWriteCloser struct {
	wantErr error
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	return 0, m.wantErr
}
func (m *mockWriteCloser) Close() error {
	return m.wantErr
}

type faultyCompressorFactory struct {
	contentEncoding httpheaders.ContentEncoding
}

func (f *faultyCompressorFactory) Create(w io.Writer) (io.WriteCloser, error) {
	return nil, errors.New("something went wrong")
}

func (f *faultyCompressorFactory) ContentEncoding() httpheaders.ContentEncoding {
	return f.contentEncoding
}

func Test_compressorResponseWriter_Write(t *testing.T) {
	type fields struct {
		compressor     *mockWriteCloser
		shouldCompress bool
	}
	type args struct {
		data []byte
	}
	type want struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "no data",
			args: args{data: []byte{}},
			want: want{data: []byte{}},
		},
		{
			name: "some data, no compression",
			args: args{data: []byte("1 2 3 4")},
			want: want{data: []byte("1 2 3 4")},
		},
		{
			name:   "some data, with compression",
			fields: fields{shouldCompress: true},
			args:   args{data: []byte("1 2 3 4")},
			want:   want{data: []byte("1 2 3 4")},
		},
		{
			name:    "some data, compression error",
			fields:  fields{shouldCompress: true, compressor: &mockWriteCloser{wantErr: errors.New("compressor error")}},
			args:    args{data: []byte("1 2 3 4")},
			want:    want{data: []byte{}},
			wantErr: true,
		},
		{
			name: "bulk data, no compression",
			args: args{data: []byte(strings.Repeat(" 1 2 3 4 ", 1000))},
			want: want{data: []byte(strings.Repeat(" 1 2 3 4 ", 1000))},
		},
		{
			name:   "bulk data, with compression",
			fields: fields{shouldCompress: true},
			args:   args{data: []byte(strings.Repeat(" 1 2 3 4 ", 1000))},
			want:   want{data: []byte(strings.Repeat(" 1 2 3 4 ", 1000))},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var compressor io.WriteCloser
			rw := httptest.NewRecorder()
			if tt.fields.compressor != nil {
				compressor = tt.fields.compressor
			} else {
				compressor = gzip.NewWriter(&buf)
			}
			cw := &compressorResponseWriter{
				ResponseWriter:  rw,
				compressor:      compressor,
				_shouldCompress: tt.fields.shouldCompress,
			}

			got, err := cw.Write(tt.args.data)

			if tt.wantErr {
				require.Error(t, err)
				require.Error(t, compressor.Close())
			} else {
				require.NoError(t, err)
				require.NoError(t, compressor.Close())
			}
			assert.Equal(t, len(tt.want.data), got)

			var rbody io.ReadCloser
			if tt.fields.shouldCompress {
				assert.Empty(t, rw.Body.Bytes())
				rbody, err = gzip.NewReader(&buf)
				if tt.wantErr {
					require.Error(t, err)
					rbody = nil
				} else {
					require.NoError(t, err)
				}
			} else {
				resp := rw.Result()
				defer func() { assert.NoError(t, resp.Body.Close()) }()
				rbody = resp.Body
			}
			if rbody != nil {
				body, err := io.ReadAll(rbody)
				assert.NoError(t, err)
				assert.NoError(t, rbody.Close())
				assert.Equal(t, tt.want.data, body)
			}
		})
	}
}

func TestCompressor(t *testing.T) {
	type args struct {
		l log.Logger
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, Middleware)
	}{
		{
			name: "default",
			args: args{l: log.NewNoopLogger()},
			assertion: func(t *testing.T, args args, got Middleware) {
				next := &middlewareHandler{}
				m := got(next)
				require.IsType(t, &middlewareHandler{}, m)
				mh := m.(*middlewareHandler)
				require.IsType(t, &compressorMiddleware{}, mh.impl)
				impl := mh.impl.(*compressorMiddleware)
				assert.Equal(t, args.l, impl.logger)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, Compressor(tt.args.l))
		})
	}
}

func Test_compressorMiddleware_Intercept(t *testing.T) {
	type fields struct {
		factory compressorFactory
	}
	type args struct {
		r    func() *http.Request
		next http.Handler
	}
	type want struct {
		contentEncoding httpheaders.ContentEncoding
		dataDecoded     []byte
		status          int
	}
	type innerTest struct {
		name   string
		fields fields
		args   args
		want   want
	}
	type outerTest struct {
		name  string
		cases []innerTest
	}
	tests := []outerTest{
		{
			name: "server responds to",
			cases: []innerTest{
				{
					name: "no compression requested",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte("done!"))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("done!"),
						status:          http.StatusOK,
					},
				},
				{
					name: "gzip compression requested",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header)
							httpheaders.ContentTypeTextPlain.Apply(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte("done!"))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingGzip,
						dataDecoded:     []byte("done!"),
						status:          http.StatusOK,
					},
				},
				{
					name: "deflate compression requested",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							httpheaders.ContentEncodingDeflate.MakeAccepted(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte("done!"))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("done!"),
						status:          http.StatusOK,
					},
				},
			},
		},
		{
			name: "server responds to based on content type",
			cases: []innerTest{
				{
					name: "gzip compression, application/json",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte(`{"id": 1}`))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingGzip,
						dataDecoded:     []byte(`{"id": 1}`),
						status:          http.StatusOK,
					},
				},
				{
					name: "gzip compression, text/html",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextHTML.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte("<html></html>"))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingGzip,
						dataDecoded:     []byte("<html></html>"),
						status:          http.StatusOK,
					},
				},
				{
					name: "gzip compression, text/plain",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header)
							httpheaders.ContentTypeTextPlain.Apply(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte("done!"))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingGzip,
						dataDecoded:     []byte("done!"),
						status:          http.StatusOK,
					},
				},
				{
					name: "no compression, some binary format",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentType("image/png").Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte("some binary data"))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("some binary data"),
						status:          http.StatusOK,
					},
				},
				{
					name: "no compression, text/plain, status code >= 300",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusBadRequest)
							_, err := w.Write([]byte("bad request"))
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("bad request"),
						status:          http.StatusBadRequest,
					},
				},
			},
		},
		{
			name: "client sends with",
			cases: []innerTest{
				{
					name: "no compression",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							var buf bytes.Buffer
							buf.WriteString("hello me!")
							r := httptest.NewRequest(http.MethodPost, "/", &buf)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := io.Copy(w, r.Body)
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("hello me!"),
						status:          http.StatusOK,
					},
				},
				{
					name: "gzip compression",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							var buf bytes.Buffer
							wgz := gzip.NewWriter(&buf)
							_, err := wgz.Write([]byte("hello me!"))
							require.NoError(t, err)
							require.NoError(t, wgz.Close())
							r := httptest.NewRequest(http.MethodPost, "/", &buf)
							httpheaders.ContentEncodingGzip.Apply(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := io.Copy(w, r.Body)
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("hello me!"),
						status:          http.StatusOK,
					},
				},
				{
					name: "deflate compression",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							var buf bytes.Buffer
							wgz, err := flate.NewWriter(&buf, flate.BestSpeed)
							require.NoError(t, err)
							_, err = wgz.Write([]byte("hello me!"))
							require.NoError(t, err)
							require.NoError(t, wgz.Close())
							r := httptest.NewRequest(http.MethodPost, "/", &buf)
							httpheaders.ContentEncodingDeflate.Apply(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := io.Copy(w, r.Body)
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded: func() []byte {
							var buf bytes.Buffer
							wgz, err := flate.NewWriter(&buf, flate.BestSpeed)
							require.NoError(t, err)
							_, err = wgz.Write([]byte("hello me!"))
							require.NoError(t, err)
							require.NoError(t, wgz.Close())
							return buf.Bytes()
						}(),
						status: http.StatusOK,
					},
				},
			},
		},
		{
			name: "client sends with and want response with",
			cases: []innerTest{
				{
					name: "gzip compression, gzip compression",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							var buf bytes.Buffer
							wgz := gzip.NewWriter(&buf)
							_, err := wgz.Write([]byte("hello me!"))
							require.NoError(t, err)
							require.NoError(t, wgz.Close())
							r := httptest.NewRequest(http.MethodPost, "/", &buf)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header).Apply(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := io.Copy(w, r.Body)
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingGzip,
						dataDecoded:     []byte("hello me!"),
						status:          http.StatusOK,
					},
				},
				{
					name: "gzip compression, deflate compression",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							var buf bytes.Buffer
							wgz := gzip.NewWriter(&buf)
							_, err := wgz.Write([]byte("hello me!"))
							require.NoError(t, err)
							require.NoError(t, wgz.Close())
							r := httptest.NewRequest(http.MethodPost, "/", &buf)
							httpheaders.ContentEncodingGzip.Apply(r.Header)
							httpheaders.ContentEncodingDeflate.MakeAccepted(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := io.Copy(w, r.Body)
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("hello me!"),
						status:          http.StatusOK,
					},
				},
			},
		},
		{
			name: "client sends compressed data in wrong format",
			cases: []innerTest{
				{
					name: "deflate compression, claims gzip compression",
					fields: fields{
						factory: &compressorGzipFactory{level: gzip.DefaultCompression},
					},
					args: args{
						r: func() *http.Request {
							var buf bytes.Buffer
							wgz, err := flate.NewWriter(&buf, flate.BestSpeed)
							require.NoError(t, err)
							_, err = wgz.Write([]byte("hello me!"))
							require.NoError(t, err)
							require.NoError(t, wgz.Close())
							r := httptest.NewRequest(http.MethodPost, "/", &buf)
							httpheaders.ContentEncodingGzip.Apply(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := io.Copy(w, r.Body)
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("cannot decompress request\n"),
						status:          http.StatusInternalServerError,
					},
				},
			},
		},
		{
			name: "server unable to compress",
			cases: []innerTest{
				{
					name: "sends gzip compression, wants gzip compression",
					fields: fields{
						factory: &faultyCompressorFactory{},
					},
					args: args{
						r: func() *http.Request {
							var buf bytes.Buffer
							wgz := gzip.NewWriter(&buf)
							_, err := wgz.Write([]byte("hello me!"))
							require.NoError(t, err)
							require.NoError(t, wgz.Close())
							r := httptest.NewRequest(http.MethodPost, "/", &buf)
							httpheaders.ContentEncodingGzip.MakeAccepted(r.Header).Apply(r.Header)
							return r
						},
						next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							httpheaders.ContentTypeTextPlain.Apply(w.Header())
							w.WriteHeader(http.StatusOK)
							_, err := io.Copy(w, r.Body)
							assert.NoError(t, err)
						}),
					},
					want: want{
						contentEncoding: httpheaders.ContentEncodingEmpty,
						dataDecoded:     []byte("hello me!"),
						status:          http.StatusOK,
					},
				},
			},
		},
	}
	for _, outer := range tests {
		t.Run(outer.name, func(t *testing.T) {
			for _, tt := range outer.cases {
				t.Run(tt.name, func(t *testing.T) {
					logger := log.NewTestLogger()
					m := &compressorMiddleware{
						logger:            logger,
						compressorFactory: tt.fields.factory,
					}
					rw := httptest.NewRecorder()

					m.Intercept(rw, tt.args.r(), tt.args.next)

					resp := rw.Result()
					defer func() { assert.NoError(t, resp.Body.Close()) }()
					require.True(t, tt.want.contentEncoding.Matches(resp.Header))

					var body []byte
					if tt.want.contentEncoding == httpheaders.ContentEncodingGzip {
						rgz, err := gzip.NewReader(resp.Body)
						require.NoError(t, err)
						body, err = io.ReadAll(rgz)
						require.NoError(t, err)
						require.NoError(t, rgz.Close())
					} else {
						var err error
						body, err = io.ReadAll(resp.Body)
						require.NoError(t, err)
					}
					assert.Equal(t, tt.want.dataDecoded, body)
				})
			}
		})
	}
}

func Test_compressorResponseWriter_Close(t *testing.T) {
	type fields struct {
		compressor     io.WriteCloser
		shouldCompress bool
	}
	tests := []struct {
		name      string
		fields    fields
		assertion func(*testing.T, log.TestLogEventSet)
	}{
		{
			name:   "no error, should not compress",
			fields: fields{compressor: &mockWriteCloser{}},
			assertion: func(t *testing.T, events log.TestLogEventSet) {
				assert.Empty(t, events)
			},
		},
		{
			name:   "no error, should compress",
			fields: fields{compressor: &mockWriteCloser{}},
			assertion: func(t *testing.T, events log.TestLogEventSet) {
				assert.Empty(t, events)
			},
		},
		{
			name:   "close failed, should not compress",
			fields: fields{compressor: &mockWriteCloser{wantErr: errors.New("oops")}},
			assertion: func(t *testing.T, events log.TestLogEventSet) {
				assert.Empty(t, events)
			},
		},
		{
			name:   "close failed, should compress",
			fields: fields{compressor: &mockWriteCloser{wantErr: errors.New("oops")}, shouldCompress: true},
			assertion: func(t *testing.T, events log.TestLogEventSet) {
				require.Len(t, events, 1)
				assert.Equal(t, log.LevelError, events[0].Level())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw := &compressorResponseWriter{
				compressor:      tt.fields.compressor,
				_shouldCompress: tt.fields.shouldCompress,
			}
			logger := log.NewTestLogger()
			cw.Close(logger)
		})
	}
}

func Test_compressorMiddleware_decompressRequest(t *testing.T) {
	type fields struct {
		logger log.Logger
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &compressorMiddleware{
				logger: tt.fields.logger,
			}
			tt.assertion(t, m.decompressRequest(tt.args.r))
		})
	}
}

func Test_compressorMiddleware_wrapResponseWriter(t *testing.T) {
	type fields struct {
		logger log.Logger
	}
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      bool
		assertion func(*testing.T, *compressorResponseWriter)
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &compressorMiddleware{
				logger: tt.fields.logger,
			}
			got, ok := m.wrapResponseWriter(tt.args.w)
			assert.Equal(t, tt.want, ok)
			tt.assertion(t, got)
		})
	}
}

func Test_compressorMiddleware_shouldCompress(t *testing.T) {
	type fields struct {
		logger log.Logger
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &compressorMiddleware{
				logger: tt.fields.logger,
			}
			assert.Equal(t, tt.want, m.shouldCompress(tt.args.r))
		})
	}
}

func Test_compressorGzipFactory_Create(t *testing.T) {
	type fields struct {
		level int
	}
	tests := []struct {
		name      string
		fields    fields
		assertion func(*testing.T, io.WriteCloser, error)
	}{
		{
			name:   "valid level",
			fields: fields{level: gzip.BestCompression},
			assertion: func(t *testing.T, got io.WriteCloser, err error) {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Implements(t, (*io.WriteCloser)(nil), got)
			},
		},
		{
			name:   "invalid level",
			fields: fields{level: -5},
			assertion: func(t *testing.T, got io.WriteCloser, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &compressorGzipFactory{
				level: tt.fields.level,
			}
			w := &bytes.Buffer{}
			got, err := f.Create(w)
			tt.assertion(t, got, err)
		})
	}
}

func Test_compressorResponseWriter_WriteHeader(t *testing.T) {
	type args struct {
		statusCode int
	}
	type want struct {
		shouldCompress bool
	}
	tests := []struct {
		name         string
		headerSetup  func(http.Header)
		args         args
		want         want
		headerAssert func(*testing.T, http.Header)
	}{
		{
			name: "status OK, text/plain, should compress",
			headerSetup: func(h http.Header) {
				httpheaders.ContentTypeTextPlain.Apply(h)
			},
			args: args{statusCode: http.StatusOK},
			want: want{shouldCompress: true},
			headerAssert: func(t *testing.T, h http.Header) {
				assert.True(t, httpheaders.ContentEncodingGzip.Matches(h))
			},
		},
		{
			name: "status NotFound, text/plain, should not compress",
			headerSetup: func(h http.Header) {
				httpheaders.ContentTypeTextPlain.Apply(h)
			},
			args: args{statusCode: http.StatusNotFound},
			want: want{shouldCompress: false},
			headerAssert: func(t *testing.T, h http.Header) {
				assert.True(t, httpheaders.ContentEncodingEmpty.Matches(h))
			},
		},
		{
			name: "status OK, text/html, should compress",
			headerSetup: func(h http.Header) {
				httpheaders.ContentTypeTextHTML.Apply(h)
			},
			args: args{statusCode: http.StatusOK},
			want: want{shouldCompress: true},
			headerAssert: func(t *testing.T, h http.Header) {
				assert.True(t, httpheaders.ContentEncodingGzip.Matches(h))
			},
		},
		{
			name: "status OK, application/json, should compress",
			headerSetup: func(h http.Header) {
				httpheaders.ContentTypeApplicationJSON.Apply(h)
			},
			args: args{statusCode: http.StatusOK},
			want: want{shouldCompress: true},
			headerAssert: func(t *testing.T, h http.Header) {
				assert.True(t, httpheaders.ContentEncodingGzip.Matches(h))
			},
		},
		{
			name: "status OK, image/png, should not compress",
			headerSetup: func(h http.Header) {
				httpheaders.ContentType("image/png").Apply(h)
			},
			args: args{statusCode: http.StatusOK},
			want: want{shouldCompress: false},
			headerAssert: func(t *testing.T, h http.Header) {
				assert.True(t, httpheaders.ContentEncodingEmpty.Matches(h))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			cw := &compressorResponseWriter{
				ResponseWriter:  rw,
				compressor:      &mockWriteCloser{},
				contentEncoding: httpheaders.ContentEncodingGzip,
			}
			tt.headerSetup(cw.Header())

			cw.WriteHeader(tt.args.statusCode)

			assert.Equal(t, tt.want.shouldCompress, cw._shouldCompress)
			resp := rw.Result()
			defer func() { assert.NoError(t, resp.Body.Close()) }()
			assert.Equal(t, tt.args.statusCode, resp.StatusCode)
			tt.headerAssert(t, cw.Header())
		})
	}
}

func Test_compressorGzipFactory_ContentEncoding(t *testing.T) {
	type fields struct {
		level int
	}
	tests := []struct {
		name   string
		fields fields
		want   httpheaders.ContentEncoding
	}{
		{
			name:   "sets gzip content encoding",
			fields: fields{level: gzip.DefaultCompression},
			want:   httpheaders.ContentEncodingGzip,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &compressorGzipFactory{
				level: tt.fields.level,
			}
			assert.Equal(t, tt.want, f.ContentEncoding())
		})
	}
}
