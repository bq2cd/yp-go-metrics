package middleware

import (
	"net/http"
)

// Middleware defines a function that wraps provided HTTP handler and returns new HTTP handler.
type Middleware func(http.Handler) http.Handler

type middleware interface {
	Intercept(w http.ResponseWriter, r *http.Request, next http.Handler)
}

type middlewareHandler struct {
	impl middleware
	next http.Handler
}

// ServeHTTP implements [http.Handler] interface.
// It will call underlying middleware implementation and pass to it the next
// HTTP handler.
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
