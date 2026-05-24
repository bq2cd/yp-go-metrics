package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type trustedSubnetMiddleware struct {
	logger        log.Logger
	trustedSubnet net.IPNet
	signer        hmacsigner.HMACSigner
}

// TrustedSubnet creates a middleware responsible for validation of `X-Real-IP` header.
// The header must contain an IP address that belongs to a configured trusted subnet;
// if that is not the case, processing of the request is aborted and `403 Forbidden` is returned.
func TrustedSubnet(l log.Logger, trustedSubnet net.IPNet, signer hmacsigner.HMACSigner) Middleware {
	if l == nil {
		l = log.NewNoopLogger()
	}

	m := &trustedSubnetMiddleware{
		logger:        l.With(log.Str("middleware", "trusted_subnet")),
		trustedSubnet: trustedSubnet,
		signer:        signer,
	}

	return createMiddleware(m)
}

// Intercept attempts to validate `X-Real-IP` header.
// The header must contain an IP address that belongs to a configured trusted subnet;
// if that is not the case, processing of the request is aborted and `403 Forbidden` is returned.
// The header might also contain HMAC signature of IP address bytes; if HMAC secret key is configured,
// then requests with invalid HMAC signature are also rejected with `403 Forbidden`.
func (m *trustedSubnetMiddleware) Intercept(w http.ResponseWriter, r *http.Request, next http.Handler) {
	ok, err := m.validateXRealIP(r)
	if err != nil {
		m.logger.Error().WithErr(err).Msg("cannot validate X-Real-IP")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if !ok {
		w.WriteHeader(http.StatusForbidden)

		return
	}

	next.ServeHTTP(w, r)
}

func (m *trustedSubnetMiddleware) isTrustedSubnetConfigured() bool {
	return len(m.trustedSubnet.IP) > 0 && len(m.trustedSubnet.Mask) > 0
}

func (m *trustedSubnetMiddleware) validateXRealIP(r *http.Request) (bool, error) {
	if !m.isTrustedSubnetConfigured() {
		return true, nil // allow all requests when trusted subnet is not configured
	}

	realIP := httpheaders.GetXRealIP(r.Header)
	if realIP.Empty() {
		return false, nil // requests without header are rejected
	}

	ok, err := m.verifyXRealIPHash(realIP)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil // reject requests with invalid hash
	}

	if !m.trustedSubnet.Contains(realIP.IP) {
		return false, nil // reject requests with IP not in trusted subnet
	}

	return true, nil
}

func (m *trustedSubnetMiddleware) verifyXRealIPHash(realIP httpheaders.XRealIP) (bool, error) {
	if !m.signer.HasKey() {
		return true, nil // ignore hash if no secret key is configured; assume IP is valid
	}

	err := m.signer.Verify(realIP.IP.To16(), realIP.Hash) // ensure we verify longest possible IP bytes; the sender must sign the same length
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, hmacsigner.ErrSignatureMismatch):
		return false, nil
	default:
		return false, fmt.Errorf("cannot verify X-Real-IP hash: %w", err)
	}
}
