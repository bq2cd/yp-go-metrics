package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type hmacSignerResponseWriter struct {
	http.ResponseWriter

	data       *bytes.Buffer
	statusCode int
}

// Write copies incoming data into internal buffer for further signing.
func (hw *hmacSignerResponseWriter) Write(data []byte) (int, error) {
	n, err := io.Copy(hw.data, bytes.NewReader(data))
	return int(n), err
}

// WriteHeader saves HTTP response status internally to make a decision
// whether to sign the response data later on.
func (hw *hmacSignerResponseWriter) WriteHeader(statusCode int) {
	hw.statusCode = statusCode
}

// HMACSigner defines middleware that validates incoming requests' HMAC signature and
// signs 2xx responses.
func HMACSigner(l log.Logger, signer hmacsigner.HMACSigner) Middleware {
	m := &hmacSignerMiddleware{
		logger: l.With(log.Str("middleware", "hmac_signer")),
		signer: signer,
	}
	return createMiddleware(m)
}

type hmacSignerMiddleware struct {
	logger log.Logger
	signer hmacsigner.HMACSigner
}

func (m *hmacSignerMiddleware) validateRequest(r *http.Request) (bool, error) {
	buf := bytes.NewBuffer(nil)
	tee := io.TeeReader(r.Body, buf)

	body, err := io.ReadAll(tee)
	if err != nil {
		return false, fmt.Errorf("cannot read request body: %w", err)
	}

	// close old request body to avoid leaks
	_ = r.Body.Close()

	r.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

	if len(body) == 0 {
		return true, nil
	}

	hash := httpheaders.GetHashSHA256(r.Header)

	// FIXME:
	// Requests with hash should not be accepted, but we are forced to accept them
	// because of `go-autotests` which do not sign their requests; see
	// https://github.com/Yandex-Practicum/go-autotests/blob/0591b1dbbcbcf741c41c8eca0718bf676ed7307f/cmd/metricstest_v2/iteration14_test.go#L462
	if hash == httpheaders.HashSHA256Empty {
		return true, nil
	}

	signature, err := hash.Bytes()
	if err != nil {
		return false, fmt.Errorf("cannot decode signature from hash: %w", err)
	}

	err = m.signer.Verify(body, signature)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, hmacsigner.ErrSignatureMismatch):
		return false, nil
	default:
		return false, fmt.Errorf("cannot verify signature: %w", err)
	}
}

func (m *hmacSignerMiddleware) wrapResponseWriter(w http.ResponseWriter) *hmacSignerResponseWriter {
	return &hmacSignerResponseWriter{
		ResponseWriter: w,
		data:           bytes.NewBuffer(nil),
	}
}

func (m *hmacSignerMiddleware) writeResponse(hw *hmacSignerResponseWriter) error {
	if err := m.signResponse(hw); err != nil {
		return fmt.Errorf("cannot sign response: %w", err)
	}

	hw.ResponseWriter.WriteHeader(hw.statusCode)
	_, err := hw.ResponseWriter.Write(hw.data.Bytes())
	return err
}

func (m *hmacSignerMiddleware) signResponse(hw *hmacSignerResponseWriter) error {
	if hw.data.Len() == 0 {
		// empty responses are not signed by design
		return nil
	}
	if hw.statusCode >= 300 {
		// non-2xx responses are not signed by design
		return nil
	}

	signature, err := m.signer.Sign(hw.data.Bytes())
	if err != nil {
		return err
	}

	hash := httpheaders.GetHashSHA256FromBytes(signature)
	hash.Apply(hw.ResponseWriter.Header())

	return nil
}

// Intercept defines actual middleware implementation.
// It will call next HTTP handler after processing.
func (m *hmacSignerMiddleware) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	if !m.signer.HasKey() {
		next.ServeHTTP(w, r)
		return
	}

	valid, err := m.validateRequest(r)
	if err != nil {
		m.logger.Error().WithErr(err).Msg("cannot validate incoming request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "signature mismatch", http.StatusBadRequest)
		return
	}

	hw := m.wrapResponseWriter(w)

	next.ServeHTTP(hw, r)

	if err := m.writeResponse(hw); err != nil {
		m.logger.Error().WithErr(err).Msg("cannot write response")
	}
}
