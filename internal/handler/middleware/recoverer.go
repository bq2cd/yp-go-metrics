package middleware

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/log"
)

// Recoverer is a middleware that implements recovery from panics.
// Once panic occurs inside HTTP handler, it will log a stack trace
// and return 500 status code (internal server error).
func Recoverer(l log.Logger) Middleware {
	m := &recovererMiddleware{
		logger: l.With(log.Str("middleware", "recoverer")),
	}
	return createMiddleware(m)
}

type recovererMiddleware struct {
	logger log.Logger
}

func (m *recovererMiddleware) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	defer func() {
		if err := recover(); err != nil {
			m.logger.Error().Any("panic", err).Msg("internal error")
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}()
	next.ServeHTTP(w, r)
}
