package interceptors

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/validators"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type trustedSubnet struct {
	logger    log.Logger
	validator *validators.TrustedSubnet
}

// TrustedSubnet creates an instance of [UnaryInterceptor] that validates incoming requests
// by checking if IP address from `x-real-ip` metadata belongs to trusted subnet.
func TrustedSubnet(logger log.Logger, subnet net.IPNet, signer hmacsigner.HMACSigner) *trustedSubnet {
	if logger == nil {
		logger = log.NewNoopLogger()
	}

	return &trustedSubnet{
		logger:    logger.With(log.Str("interceptor", "trusted_subnet")),
		validator: validators.NewTrustedSubnet(subnet, signer),
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
	realIP := m.extractXRealIPFromMetadata(ctx)

	return m.validator.IsXRealIPTrusted(realIP)
}

func (m *trustedSubnet) extractXRealIPFromMetadata(ctx context.Context) httpheaders.XRealIP {
	var zero httpheaders.XRealIP

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return zero
	}

	values := md.Get(strings.ToLower(httpheaders.HeaderKeyXRealIP))
	if len(values) != 1 {
		return zero // requests with multi-value x-real-ip are treated as not having x-real-ip at all
	}

	return httpheaders.GetXRealIPFromBytes([]byte(values[0]))
}
