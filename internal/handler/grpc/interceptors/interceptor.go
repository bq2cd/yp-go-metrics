package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryInterceptor defines an interface for unary server gRPC interceptors.
type UnaryInterceptor interface {
	Intercept(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error)
}
