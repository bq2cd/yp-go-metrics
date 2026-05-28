package middleware

import (
	"net"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/validators"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type trustedSubnetMiddleware struct {
	logger    log.Logger
	validator *validators.TrustedSubnet
}

// TrustedSubnet creates a middleware responsible for validation of `X-Real-IP` header.
// The header must contain an IP address that belongs to a configured trusted subnet;
// if that is not the case, processing of the request is aborted and `403 Forbidden` is returned.
func TrustedSubnet(l log.Logger, trustedSubnet net.IPNet, signer hmacsigner.HMACSigner) Middleware {
	if l == nil {
		l = log.NewNoopLogger()
	}

	m := &trustedSubnetMiddleware{
		logger:    l.With(log.Str("middleware", "trusted_subnet")),
		validator: validators.NewTrustedSubnet(trustedSubnet, signer),
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

func (m *trustedSubnetMiddleware) validateXRealIP(r *http.Request) (bool, error) {
	realIP := httpheaders.GetXRealIP(r.Header)

	return m.validator.IsXRealIPTrusted(realIP)
}
