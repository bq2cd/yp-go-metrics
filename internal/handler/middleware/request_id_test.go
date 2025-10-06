package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(*testing.T, Middleware)
	}{
		{
			name: "default",
			assertion: func(t *testing.T, got Middleware) {
				next := &middlewareHandler{}
				m := got(next)
				require.IsType(t, &middlewareHandler{}, m)
				mh := m.(*middlewareHandler)
				require.IsType(t, &requestIDMiddleware{}, mh.impl)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, RequestID())
		})
	}
}

func Test_generateRequestID(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(*testing.T, string)
	}{
		{
			name: "not empty",
			assertion: func(t *testing.T, got string) {
				assert.Greater(t, len(got), 1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, generateRequestID())
		})
	}
}

func Test_getOrGenerateRequestID(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, string)
	}{
		{
			name: "nil context",
			args: args{ctx: nil},
			assertion: func(t *testing.T, got string) {
				assert.NotEqual(t, emptyRequestID, got)
			},
		},
		{
			name: "context without request id",
			args: args{ctx: context.Background()},
			assertion: func(t *testing.T, got string) {
				assert.NotEqual(t, emptyRequestID, got)
			},
		},
		{
			name: "context with request id",
			args: args{ctx: context.WithValue(context.Background(), defaultRequestIDKey, "super-puper-request-id")},
			assertion: func(t *testing.T, got string) {
				assert.Equal(t, "super-puper-request-id", got)
			},
		},
		{
			name: "context with wrong value type",
			args: args{ctx: context.WithValue(context.Background(), defaultRequestIDKey, []byte("super-puper-request-id"))},
			assertion: func(t *testing.T, got string) {
				assert.Equal(t, emptyRequestID, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, getOrGenerateRequestID(tt.args.ctx))
		})
	}
}

func Test_requestIDMiddleware_Intercept(t *testing.T) {
	type args struct {
		w    *httptest.ResponseRecorder
		r    *http.Request
		next http.Handler
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "has request id",
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rID := getOrGenerateRequestID(r.Context())
					w.Header().Set("X-Request-ID", rID)
					w.WriteHeader(http.StatusOK)
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &requestIDMiddleware{}

			m.Intercept(tt.args.w, tt.args.r, tt.args.next)

			resp := tt.args.w.Result()
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.NotEqual(t, emptyRequestID, resp.Header.Get("X-Request-ID"))
		})
	}
}
