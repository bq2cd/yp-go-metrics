package middleware

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
	"github.com/bq2cd/yp-go-metrics/pkg/bufpool"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

var (
	// ErrRequestDecryptionFailed is returned when request is not properly encrypted (wrong encryption key, etc.).
	ErrRequestDecryptionFailed = errors.New("request decryption failed")
)

type requestDecryptorMiddleware struct {
	logger    log.Logger
	decryptor asymcrypt.Decryptor
	bufpool   *bufpool.Pool
}

// RequestDecryptor creates a middleware responsible for decrypting incoming requests but only
// if actual decryptor is configured (otherwise, requests are passed down the chain as is).
// Provided [service.Decryptor] can be `nil` to disable request decryption.
func RequestDecryptor(l log.Logger, decryptor asymcrypt.Decryptor) Middleware {
	if l == nil { // a bit of defense-in-depth
		l = log.NewNoopLogger()
	}

	m := &requestDecryptorMiddleware{
		logger:    l.With(log.Str("middleware", "decryptor")),
		decryptor: decryptor,
		bufpool:   bufpool.New(),
	}

	return createMiddleware(m)
}

// Intercept attempts to decrypt incoming request if the middleware has a [Decryptor] configured,
// i.e. `decryptor` field is not `nil`.
// Otherwise, the request is passed to the next handler as is.
func (m *requestDecryptorMiddleware) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	if m.decryptor == nil {
		next.ServeHTTP(w, r)

		return
	}

	err := m.decryptRequest(r)
	if err == nil {
		next.ServeHTTP(w, r)

		return
	}

	m.logger.Error().WithErr(err).Msg("cannot decrypt incoming request")

	switch {
	case errors.Is(err, ErrRequestDecryptionFailed):
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *requestDecryptorMiddleware) decryptRequest(r *http.Request) error {
	buf := m.bufpool.Get()

	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return fmt.Errorf("cannot read request body: %w", err)
	}

	cleartext, err := m.decryptor.Decrypt(buf.Bytes())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestDecryptionFailed, err)
	}

	// reuse buffer
	buf.Reset()
	_, err = buf.Write(cleartext)
	if err != nil { // unlikely, but let's handle
		return fmt.Errorf("cannot write decrypted request body: %w", err)
	}

	// we have new request body now, can ignore closing error
	_ = r.Body.Close()

	r.Body = buf // buf will return to the pool when HTTP server will call `Close()` on the body.

	return nil
}
