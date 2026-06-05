package middleware

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner/hmacsignertest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func Test_trustedSubnet_Intercept(t *testing.T) {
	mirrorHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.Copy(w, r.Body)
	}

	type want struct {
		status int
		body   []byte
	}
	type testcase struct {
		trustedSubnet net.IPNet
		signer        func(*gomock.Controller) hmacsigner.HMACSigner
		next          http.HandlerFunc
		req           func() *http.Request
		want          want
	}

	tests := map[string]testcase{
		"no trusted subnet allows all requests": {
			trustedSubnet: net.IPNet{},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				return hmacsignertest.NewMockHMACSigner(ctrl)
			},
			next: mirrorHandler,
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			},
			want: want{
				status: http.StatusOK,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				return r
			},
			want: want{
				status: http.StatusForbidden,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "not-an-ip")
				return r
			},
			want: want{
				status: http.StatusForbidden,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "10.1.1.1")
				return r
			},
			want: want{
				status: http.StatusForbidden,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "10.1.1.1;hash=valid-hash")
				return r
			},
			want: want{
				status: http.StatusForbidden,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "1.2.3.45")
				return r
			},
			want: want{
				status: http.StatusOK,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "1.2.3.45;hash=valid-hash")
				return r
			},
			want: want{
				status: http.StatusOK,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "1.2.3.45")
				return r
			},
			want: want{
				status: http.StatusForbidden,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "1.2.3.45;hash=incorrect")
				return r
			},
			want: want{
				status: http.StatusForbidden,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "10.1.1.1;hash=correct")
				return r
			},
			want: want{
				status: http.StatusForbidden,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "1.2.3.45;hash=correct")
				return r
			},
			want: want{
				status: http.StatusOK,
				body:   []byte{},
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
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				r.Header.Set(httpheaders.HeaderKeyXRealIP, "1.2.3.45;hash=correct")
				return r
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte{},
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
				req: func() *http.Request {
					r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
					ip := net.ParseIP("1.2.3.45").To4()
					hash, err := signer.Sign(ip)
					require.NoErrorf(t, err, "cannot sign IP bytes")
					realIP := httpheaders.XRealIP{IP: ip, Hash: hash}
					r.Header.Set(httpheaders.HeaderKeyXRealIP, realIP.String())
					return r
				},
				want: want{
					status: http.StatusForbidden,
					body:   []byte{},
				},
			}
		}(),
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := log.NewTestLogger()

			rw := httptest.NewRecorder()
			m := TrustedSubnet(logger, tc.trustedSubnet, tc.signer(ctrl))(tc.next)

			// Act
			m.ServeHTTP(rw, tc.req())

			// Assert
			resp := rw.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer func() { assert.NoError(t, resp.Body.Close()) }()

			assert.Equal(t, tc.want.status, resp.StatusCode)
			assert.Equal(t, tc.want.body, body)
		})
	}
}
