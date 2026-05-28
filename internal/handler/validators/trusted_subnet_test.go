package validators

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner/hmacsignertest"
)

func TestTrustedSubnet_IsXRealIPTrusted(t *testing.T) {
	type testcase struct {
		subnet        net.IPNet
		signer        func(*gomock.Controller) hmacsigner.HMACSigner
		realIP        httpheaders.XRealIP
		wantTrusted   bool
		wantErrString string
	}

	tests := map[string]testcase{
		"trusted subnet is not configured, any IP is trusted": {
			subnet: net.IPNet{},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				return hmacsignertest.NewMockHMACSigner(ctrl)
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("9.9.9.9")},
			wantTrusted: true,
		},
		"trusted subnet is not configured, empty IP is allowed": {
			subnet: net.IPNet{},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				return hmacsignertest.NewMockHMACSigner(ctrl)
			},
			realIP:      httpheaders.XRealIP{},
			wantTrusted: true,
		},
		"trusted subnet rejects empty IP": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				return hmacsignertest.NewMockHMACSigner(ctrl)
			},
			realIP:      httpheaders.XRealIP{},
			wantTrusted: false,
		},
		"trusted subnet rejects IP from different subnet": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("9.9.9.9")},
			wantTrusted: false,
		},
		"trusted subnet rejects IP from different subnet (hash ignored)": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("9.9.9.9"), Hash: []byte(`anything`)},
			wantTrusted: false,
		},
		"trusted subnet accepts IP from the same subnet": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("1.2.3.4")},
			wantTrusted: true,
		},
		"trusted subnet accepts IP from the same subnet (hash ignored)": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(false)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("1.2.3.4"), Hash: []byte(`anything`)},
			wantTrusted: true,
		},
		"HMAC + trusted subnet rejects empty IP": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				return m
			},
			realIP:      httpheaders.XRealIP{},
			wantTrusted: false,
		},
		"HMAC + trusted subnet rejects IP without hash": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(net.ParseIP("1.2.3.4").To16(), nil).Return(hmacsigner.ErrSignatureMismatch)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("1.2.3.4")},
			wantTrusted: false,
		},
		"HMAC + trusted subnet rejects IP with incorrect hash": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(net.ParseIP("1.2.3.4").To16(), []byte(`anything`)).Return(hmacsigner.ErrSignatureMismatch)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("1.2.3.4"), Hash: []byte(`anything`)},
			wantTrusted: false,
		},
		"HMAC + trusted subnet rejects IP with correct hash but from different subnet": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(net.ParseIP("1.2.99.4").To16(), []byte(`anything`)).Return(nil)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("1.2.99.4"), Hash: []byte(`anything`)},
			wantTrusted: false,
		},
		"HMAC + trusted subnet accepts IP with correct hash and the same subnet": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(net.ParseIP("1.2.3.4").To16(), []byte(`anything`)).Return(nil)
				return m
			},
			realIP:      httpheaders.XRealIP{IP: net.ParseIP("1.2.3.4"), Hash: []byte(`anything`)},
			wantTrusted: true,
		},
		"HMAC + trusted subnet returns error on HMACSigner failure": {
			subnet: net.IPNet{
				IP:   net.ParseIP("1.2.3.0"),
				Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
			},
			signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().HasKey().Return(true)
				m.EXPECT().Verify(net.ParseIP("1.2.3.4").To16(), []byte(`anything`)).Return(fmt.Errorf("oops, no signature validation today"))
				return m
			},
			realIP:        httpheaders.XRealIP{IP: net.ParseIP("1.2.3.4"), Hash: []byte(`anything`)},
			wantTrusted:   false,
			wantErrString: "oops, no signature validation today",
		},
		"HMAC (real signer) + trusted subnet rejects IP from the same subnet but with hash obtained from 4-byte IP slice": func() testcase {
			signer := hmacsigner.NewHMACSigner([]byte(`super-secret-key`))

			return testcase{
				subnet: net.IPNet{
					IP:   net.ParseIP("1.2.3.0"),
					Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
				},
				signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
					return signer
				},
				realIP: func() httpheaders.XRealIP {
					ip := net.ParseIP("1.2.3.4")
					hash, err := signer.Sign(ip.To4())
					require.NoErrorf(t, err, "cannot sign IP bytes")
					return httpheaders.XRealIP{IP: ip, Hash: hash}
				}(),
				wantTrusted: false,
			}
		}(),
		"HMAC (real signer) + trusted subnet accepts IP from the same subnet and with hash obtained from 16-byte IP slice": func() testcase {
			signer := hmacsigner.NewHMACSigner([]byte(`super-secret-key`))

			return testcase{
				subnet: net.IPNet{
					IP:   net.ParseIP("1.2.3.0"),
					Mask: net.CIDRMask(24, 32), // 1.2.3.0/24
				},
				signer: func(ctrl *gomock.Controller) hmacsigner.HMACSigner {
					return signer
				},
				realIP: func() httpheaders.XRealIP {
					ip := net.ParseIP("1.2.3.4")
					hash, err := signer.Sign(ip.To16())
					require.NoErrorf(t, err, "cannot sign IP bytes")
					return httpheaders.XRealIP{IP: ip, Hash: hash}
				}(),
				wantTrusted: true,
			}
		}(),
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			validator := NewTrustedSubnet(tc.subnet, tc.signer(ctrl))

			// Act
			trusted, err := validator.IsXRealIPTrusted(tc.realIP)

			// Assert
			if tc.wantErrString == "" {
				require.NoErrorf(t, err, "unexpected validator error")
			} else {
				require.ErrorContainsf(t, err, tc.wantErrString, "unexpected validator error")
			}

			assert.Equalf(t, tc.wantTrusted, trusted, "unexpected validator decision")
		})
	}
}
