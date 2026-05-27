package interceptors

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner/hmacsignertest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func Test_trustedSubnet_Intercept(t *testing.T) {
	mirrorHandler := func(_ context.Context, in any) (any, error) {
		return in, nil
	}

	type want struct {
		code codes.Code
		resp any
	}
	type testcase struct {
		trustedSubnet net.IPNet
		signer        func(*gomock.Controller) hmacsigner.HMACSigner
		next          grpc.UnaryHandler
		req           func() (context.Context, any)
		want          want
	}

	tests := map[string]testcase{
		"no trusted subnet allows all requests": {
			trustedSubnet: net.IPNet{},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				return context.TODO(), "a request"
			},
			want: want{
				code: codes.OK,
				resp: "a request",
			},
		},
		"trusted subnet rejects requests without IP header": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				return context.TODO(), "a request"
			},
			want: want{
				code: codes.PermissionDenied,
				resp: nil,
			},
		},
		"trusted subnet rejects requests with invalid IP header": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "not-an-ip"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.PermissionDenied,
				resp: nil,
			},
		},
		"trusted subnet rejects requests with IP from different subnet": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "10.1.1.1"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.PermissionDenied,
				resp: nil,
			},
		},
		"trusted subnet rejects requests with IP from different subnet (hash ignored)": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "10.1.1.1;hash=valid-hash"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.PermissionDenied,
				resp: nil,
			},
		},
		"trusted subnet accepts requests with IP from the same subnet": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "1.2.3.45"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.OK,
				resp: "a request",
			},
		},
		"trusted subnet accepts requests with IP from the same subnet (hash ignored)": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "1.2.3.45;hash=valid-hash"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.OK,
				resp: "a request",
			},
		},
		"trusted subnet with signer rejects requests without IP hash": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(hmacsigner.ErrSignatureMismatch)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "1.2.3.45"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.PermissionDenied,
				resp: nil,
			},
		},
		"trusted subnet with signer rejects requests with incorrect IP hash": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(hmacsigner.ErrSignatureMismatch)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "1.2.3.45;hash=incorrect"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.PermissionDenied,
				resp: nil,
			},
		},
		"trusted subnet with signer rejects requests with correct IP hash but with IP from different subnet": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "10.1.1.1;hash=correct"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.PermissionDenied,
				resp: nil,
			},
		},
		"trusted subnet with signer accepts requests with correct IP hash and with IP from the same subnet": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil)
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "1.2.3.45;hash=correct"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.OK,
				resp: "a request",
			},
		},
		"trusted subnet with signer fails on IP hash verification": {
			trustedSubnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(fmt.Errorf("unexpected error"))
				return m
			},
			next: mirrorHandler,
			req: func() (context.Context, any) {
				md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): "1.2.3.45;hash=correct"})
				ctx := metadata.NewIncomingContext(context.TODO(), md)
				return ctx, "a request"
			},
			want: want{
				code: codes.Internal,
				resp: nil,
			},
		},
		"trusted subnet with real signer fails on IP hash verification with different IP bytes length": func() testcase {
			signer := hmacsigner.NewHMACSigner([]byte(`super-secret-key`))

			return testcase{
				trustedSubnet: net.IPNet{
					IP:   net.ParseIP("1.2.3.0"),
					Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
				},
				signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
					return signer
				},
				next: mirrorHandler,
				req: func() (context.Context, any) {
					ip := net.ParseIP("1.2.3.45").To4()
					hash, err := signer.Sign(ip)
					require.NoErrorf(t, err, "cannot sign IP bytes")
					realIP := httpheaders.XRealIP{IP: ip, Hash: hash}
					md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): realIP.String()})
					ctx := metadata.NewIncomingContext(context.TODO(), md)
					return ctx, "a request"
				},
				want: want{
					code: codes.PermissionDenied,
					resp: nil,
				},
			}
		}(),
		"trusted subnet with real signer succeeds with IP hash verification and accepts IP from the same subnet": func() testcase {
			signer := hmacsigner.NewHMACSigner([]byte(`super-secret-key`))

			return testcase{
				trustedSubnet: net.IPNet{
					IP:   net.ParseIP("1.2.3.0"),
					Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
				},
				signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
					return signer
				},
				next: mirrorHandler,
				req: func() (context.Context, any) {
					ip := net.ParseIP("1.2.3.45").To16()
					hash, err := signer.Sign(ip)
					require.NoErrorf(t, err, "cannot sign IP bytes")
					realIP := httpheaders.XRealIP{IP: ip, Hash: hash}
					md := metadata.New(map[string]string{strings.ToLower(httpheaders.HeaderKeyXRealIP): realIP.String()})
					ctx := metadata.NewIncomingContext(context.TODO(), md)
					return ctx, "a request"
				},
				want: want{
					code: codes.OK,
					resp: "a request",
				},
			}
		}(),
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := log.NewTestLogger()

			m := TrustedSubnet(logger, tc.trustedSubnet, tc.signer(ctrl))

			ctx, req := tc.req()

			// Act
			resp, err := m.Intercept(ctx, req, nil, tc.next)

			// Assert
			st, _ := status.FromError(err)
			assert.Equalf(t, tc.want.code, st.Code(), "unexpected gRPC status code")
			assert.Equalf(t, tc.want.resp, resp, "unexpected gRPC response")
		})
	}
}
