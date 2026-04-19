package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
)

const (
	defaultRequestIDKey = requestIDKey("request_id")
	emptyRequestID      = ""
)

type requestIDKey string

// RequestID is middleware that generates a random request ID and passes it to the next HTTP handler via request context.
func RequestID() Middleware {
	m := &requestIDMiddleware{}
	return createMiddleware(m)
}

func generateRequestID() string {
	var buf [16]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		return emptyRequestID
	}
	return hex.EncodeToString(buf[:])
}

func getOrGenerateRequestID(ctx context.Context) string {
	if ctx == nil {
		return generateRequestID()
	}
	v := ctx.Value(defaultRequestIDKey)
	if v == nil {
		return generateRequestID()
	}
	if rID, ok := v.(string); ok {
		return rID
	}
	return emptyRequestID
}

type requestIDMiddleware struct{}

// Intercept defines actual middleware implementation.
// It will call next HTTP handler after processing.
func (m *requestIDMiddleware) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	ctx := r.Context()
	rID := getOrGenerateRequestID(ctx)
	if rID != emptyRequestID {
		ctx = context.WithValue(ctx, defaultRequestIDKey, rID)
	}
	next.ServeHTTP(w, r.WithContext(ctx))
}
