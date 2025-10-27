package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverer(t *testing.T) {
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
				require.IsType(t, &recovererMiddleware{}, mh.impl)
				impl := mh.impl.(*recovererMiddleware)
				assert.Equal(t, args.l, impl.logger)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, Recoverer(tt.args.l))
		})
	}
}

func Test_recovererMiddleware_Intercept(t *testing.T) {
	type args struct {
		w    *httptest.ResponseRecorder
		r    *http.Request
		next http.Handler
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, log.TestLogEventSet, *httptest.ResponseRecorder)
	}{
		{
			name: "no panic",
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			},
			assertion: func(t *testing.T, events log.TestLogEventSet, rec *httptest.ResponseRecorder) {
				resp := rec.Result()
				defer func() { _ = resp.Body.Close() }()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Empty(t, events)
			},
		},
		{
			name: "panic",
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic("oops, I did it again!")
				}),
			},
			assertion: func(t *testing.T, events log.TestLogEventSet, rec *httptest.ResponseRecorder) {
				resp := rec.Result()
				defer func() { _ = resp.Body.Close() }()
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				require.Len(t, events, 1)
				e := events[0]

				assert.Equal(t, log.LevelError, e.Level())
				assert.Equal(t, "recovered from panic", e.Message())
				fp := e.Fields().GetFieldByKey("panic")

				require.NotNil(t, fp)
				assert.Equal(t, "oops, I did it again!", fp.Value)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewTestLogger()
			m := &recovererMiddleware{
				logger: logger,
			}

			m.Intercept(tt.args.w, tt.args.r, tt.args.next)

			tt.assertion(t, logger.RecordedEvents(), tt.args.w)
		})
	}
}
