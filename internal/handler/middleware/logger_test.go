package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type mockResponseWriter struct {
	mock.Mock
}

func (m *mockResponseWriter) Header() http.Header {
	m.Called()
	return http.Header{}
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.Called(data)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

func TestLogger(t *testing.T) {
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
				require.IsType(t, &loggerMiddleware{}, mh.impl)
				impl := mh.impl.(*loggerMiddleware)
				assert.Equal(t, args.l, impl.logger)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, Logger(tt.args.l))
		})
	}
}

func Test_loggerResponseWriter_Header(t *testing.T) {
	type fields struct {
		w *mockResponseWriter
	}
	tests := []struct {
		name   string
		fields fields
		want   http.Header
	}{
		{
			name: "called orig",
			fields: fields{
				w: &mockResponseWriter{},
			},
			want: http.Header{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lw := &loggerResponseWriter{
				w: tt.fields.w,
			}
			tt.fields.w.On("Header").Return(tt.want).Once()

			lw.Header()

			tt.fields.w.AssertExpectations(t)
		})
	}
}

func Test_loggerResponseWriter_Write(t *testing.T) {
	type fields struct {
		w    *mockResponseWriter
		size int
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      int
		wantTotal int
		wantErr   error
	}{
		{
			name: "no data",
			fields: fields{
				w: &mockResponseWriter{},
			},
			args: args{
				data: []byte{},
			},
			want:      0,
			wantTotal: 0,
			wantErr:   nil,
		},
		{
			name: "some data",
			fields: fields{
				w: &mockResponseWriter{},
			},
			args: args{
				data: []byte("1 2 3 4"),
			},
			want:      7,
			wantTotal: 7,
			wantErr:   nil,
		},
		{
			name: "append data",
			fields: fields{
				w:    &mockResponseWriter{},
				size: 10,
			},
			args: args{
				data: []byte("1 2 3 4"),
			},
			want:      7,
			wantTotal: 17,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lw := &loggerResponseWriter{
				w:    tt.fields.w,
				size: tt.fields.size,
			}
			tt.fields.w.On("Write", tt.args.data).Return(tt.want, tt.wantErr).Once()

			got, err := lw.Write(tt.args.data)

			tt.fields.w.AssertExpectations(t)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantTotal, lw.size)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func Test_loggerResponseWriter_WriteHeader(t *testing.T) {
	type fields struct {
		w      *mockResponseWriter
		status int
	}
	type args struct {
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "no previous status",
			fields: fields{
				w: &mockResponseWriter{},
			},
			args: args{statusCode: http.StatusOK},
			want: http.StatusOK,
		},
		{
			name: "override previous status",
			fields: fields{
				w:      &mockResponseWriter{},
				status: http.StatusOK,
			},
			args: args{statusCode: http.StatusInternalServerError},
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lw := &loggerResponseWriter{
				w:      tt.fields.w,
				status: tt.fields.status,
			}
			tt.fields.w.On("WriteHeader", tt.args.statusCode).Return().Once()

			lw.WriteHeader(tt.args.statusCode)

			tt.fields.w.AssertExpectations(t)
			assert.Equal(t, tt.want, lw.status)
		})
	}
}

func Test_loggerMiddleware_Intercept(t *testing.T) {
	type args struct {
		w    *httptest.ResponseRecorder
		r    *http.Request
		next http.Handler
	}
	type want struct {
		status  int
		size    int
		elapsed time.Duration
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "fast response",
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(10 * time.Millisecond)
					_, _ = w.Write([]byte("done!"))
					w.WriteHeader(http.StatusOK)
				}),
			},
			want: want{
				status:  http.StatusOK,
				size:    5,
				elapsed: 10 * time.Millisecond,
			},
		},
		{
			name: "slow response",
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					_, _ = w.Write([]byte("done!"))
					w.WriteHeader(http.StatusOK)
				}),
			},
			want: want{
				status:  http.StatusOK,
				size:    5,
				elapsed: 100 * time.Millisecond,
			},
		},
		{
			name: "POST request, no content",
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(5 * time.Millisecond)
					w.WriteHeader(http.StatusOK)
				}),
			},
			want: want{
				status:  http.StatusOK,
				size:    0,
				elapsed: 5 * time.Millisecond,
			},
		},
		{
			name: "POST request, bad gateway",
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(1 * time.Millisecond)
					w.WriteHeader(http.StatusBadGateway)
				}),
			},
			want: want{
				status:  http.StatusBadGateway,
				size:    0,
				elapsed: 1 * time.Millisecond,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewTestLogger()
			m := &loggerMiddleware{
				logger: logger,
			}

			m.Intercept(tt.args.w, tt.args.r, tt.args.next)

			resp := tt.args.w.Result()
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tt.want.status, resp.StatusCode)

			events := logger.RecordedEvents()
			require.Len(t, events, 1)
			e := events[0]

			assert.Equal(t, log.LevelInfo, e.Level())
			assert.Equal(t, "processed request", e.Message())

			assert.True(t, e.ContainsFields(
				log.Str("uri", tt.args.r.URL.Path),
				log.Str("method", tt.args.r.Method),
				log.Int("status", resp.StatusCode),
			))

			assert.NotEmpty(t, e.Fields().GetFieldByKey("request_id"))
			fsize := e.Fields().GetFieldByKey("size")
			require.NotNil(t, fsize)
			assert.Equal(t, fsize.Value, tt.want.size)

			fduration := e.Fields().GetFieldByKey("duration")
			assert.GreaterOrEqual(t, fduration.Value, tt.want.elapsed)
		})
	}
}
