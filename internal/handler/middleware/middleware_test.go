package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMiddlewareImpl struct {
	mock.Mock
}

func (m *mockMiddlewareImpl) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	m.Called(w, r, next)
}

func Test_middlewareHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		impl *mockMiddlewareImpl
		next http.Handler
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "middleware called",
			fields: fields{
				impl: &mockMiddlewareImpl{},
				next: &middlewareHandler{}, // it implements http.Handler, but we're not going to call it
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
					require.NoError(t, err)
					return req
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &middlewareHandler{
				impl: tt.fields.impl,
				next: tt.fields.next,
			}
			tt.fields.impl.On("Intercept",
				tt.args.w,
				tt.args.r,
				tt.fields.next,
			).Return().Once()

			m.ServeHTTP(tt.args.w, tt.args.r)

			tt.fields.impl.AssertExpectations(t)
		})
	}
}

func Test_createMiddleware(t *testing.T) {
	type args struct {
		m middleware
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, Middleware)
	}{
		{
			name: "mock middleware impl",
			args: args{m: &mockMiddlewareImpl{}},
			assertion: func(t *testing.T, args args, got Middleware) {
				next := &middlewareHandler{}
				m := got(next)
				require.IsType(t, &middlewareHandler{}, m)
				mh := m.(*middlewareHandler)
				assert.Equal(t, args.m, mh.impl)
				assert.Equal(t, next, mh.next)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, createMiddleware(tt.args.m))
		})
	}
}
