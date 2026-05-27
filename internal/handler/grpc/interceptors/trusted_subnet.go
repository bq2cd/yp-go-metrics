package interceptors

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type trustedSubnet struct {
	logger log.Logger
	subnet net.IPNet
	signer hmacsigner.HMACSigner
}

// TrustedSubnet creates an instance of [UnaryInterceptor] that validates incoming requests
// by checking if IP address from `x-real-ip` metadata belongs to trusted subnet.
func TrustedSubnet(logger log.Logger, subnet net.IPNet, signer hmacsigner.HMACSigner) *trustedSubnet {
	if logger == nil {
		logger = log.NewNoopLogger()
	}

	return &trustedSubnet{
		logger: logger.With(log.Str("interceptor", "trusted_subnet")),
		subnet: subnet,
		signer: signer,
	}
}

// Intercept implements trusted subnet validation.
// It extracts IP address from `x-real-ip` metadata and checks if the IP address belongs to the trusted
// subnet. If HMAC secret key is configured, [Intercept] will require that `x-real-ip` metadata
// contains valid HMAC signature.
// On any mismatch, [status.PermissionDenied] error is returned; otherwise, execution proceeds to
// the actual gRPC handler.
func (m *trustedSubnet) Intercept(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	ok, err := m.validateXRealIP(ctx)
	if err != nil {
		m.logger.Error().WithErr(err).Msg("cannot validate X-Real-IP")

		return nil, status.Error(codes.Internal, "")
	}

	if !ok {
		return nil, status.Error(codes.PermissionDenied, "untrusted subnet")
	}

	return handler(ctx, req)
}

func (m *trustedSubnet) validateXRealIP(ctx context.Context) (bool, error) {
	if !m.isTrustedSubnetConfigured() {
		return true, nil // allow all requests when trusted subnet is not configured
	}

	realIP, ok := m.extractXRealIPFromMetadata(ctx)
	if !ok {
		return false, nil
	}

	ok, err := m.verifyXRealIPHash(realIP)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil // reject requests with invalid hash
	}

	if !m.subnet.Contains(realIP.IP) {
		return false, nil // reject requests with IP not in trusted subnet
	}

	return true, nil
}

func (m *trustedSubnet) isTrustedSubnetConfigured() bool {
	return len(m.subnet.IP) > 0 && len(m.subnet.Mask) > 0
}

func (m *trustedSubnet) extractXRealIPFromMetadata(ctx context.Context) (httpheaders.XRealIP, bool) {
	var zero httpheaders.XRealIP

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return zero, false // requests without metadata are rejected
	}

	values := md.Get(strings.ToLower(httpheaders.HeaderKeyXRealIP))
	if len(values) != 1 {
		return zero, false // requests with empty x-real-ip or with multi-value x-real-ip are rejected
	}

	realIP := httpheaders.GetXRealIPFromBytes([]byte(values[0]))
	if realIP.Empty() {
		return zero, false // requests with invalid IP address are rejected
	}

	return realIP, true
}

func (m *trustedSubnet) verifyXRealIPHash(realIP httpheaders.XRealIP) (bool, error) {
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
