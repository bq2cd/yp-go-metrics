package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type middleware interface {
	Intercept(w http.ResponseWriter, r *http.Request, next http.Handler)
}

type middlewareHandler struct {
	impl middleware
	next http.Handler
}

// ServeHTTP implements http.Handler interface
func (m *middlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.impl.Intercept(w, r, m.next)
}

func createMiddleware(m middleware) Middleware {
	return func(next http.Handler) http.Handler {
		return &middlewareHandler{
			impl: m,
			next: next,
		}
	}
}
