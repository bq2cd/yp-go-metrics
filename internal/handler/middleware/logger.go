package middleware

import (
	"net/http"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/log"
)

// Logger is a middleware that implements request and response
// logging.
func Logger(l log.Logger) Middleware {
	m := &loggerMiddleware{
		logger: l.With(log.Str("middleware", "logger")),
	}
	return createMiddleware(m)
}

type loggerResponseWriter struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (lw *loggerResponseWriter) Header() http.Header {
	return lw.w.Header()
}

func (lw *loggerResponseWriter) Write(data []byte) (int, error) {
	size, err := lw.w.Write(data)
	lw.size += size
	return size, err
}

func (lw *loggerResponseWriter) WriteHeader(statusCode int) {
	lw.status = statusCode
	lw.w.WriteHeader(statusCode)
}

type loggerMiddleware struct {
	logger log.Logger
}

func (m *loggerMiddleware) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	rID := getOrGenerateRequestID(r.Context())
	l := m.logger.With(
		log.Str("request_id", rID),
		log.Str("uri", r.URL.String()),
		log.Str("method", r.Method),
	)
	lw := &loggerResponseWriter{w: w}

	start := time.Now()
	next.ServeHTTP(lw, r)
	elapsed := time.Since(start)

	l.Info().Int("status", lw.status).Int("size", lw.size).Dur("duration", elapsed).Msg("processed request")
}
