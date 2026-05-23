package agent

import (
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner/hmacsignertest"
)

func TestGetOutgoingIPv4(t *testing.T) {
	type testcase struct {
		remoteAddr string
		wantIP     string
		wantErr    error
	}

	tests := map[string]testcase{
		"connection to 127.0.0.1 returns 127.0.0.1": {
			remoteAddr: "127.0.0.1:12345",
			wantIP:     "127.0.0.1",
		},
		"connection to localhost returns 127.0.0.1": {
			remoteAddr: "localhost:12345",
			wantIP:     "127.0.0.1",
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ip, err := getOutgoingIPv4(tc.remoteAddr)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			if err != nil {
				return
			}

			assert.Equal(t, tc.wantIP, ip.String())
		})
	}
}

func TestPrepareRealIPHeader(t *testing.T) {
	type testcase struct {
		signer        func(*gomock.Controller) *hmacsignertest.MockHMACSigner
		remoteAddr    string
		wantIP        httpheaders.XRealIP
		wantErrString string
	}

	tests := map[string]testcase{
		"empty remote address": {
			signer: func(ctrl *gomock.Controller) *hmacsignertest.MockHMACSigner {
				return hmacsignertest.NewMockHMACSigner(ctrl)
			},
			remoteAddr:    "",
			wantIP:        httpheaders.XRealIP{},
			wantErrString: "cannot detect outgoing IPv4:",
		},
		"invalid remote address (no port)": {
			signer: func(ctrl *gomock.Controller) *hmacsignertest.MockHMACSigner {
				return hmacsignertest.NewMockHMACSigner(ctrl)
			},
			remoteAddr:    "127.0.0.1",
			wantIP:        httpheaders.XRealIP{},
			wantErrString: "cannot detect outgoing IPv4:",
		},
		"valid remote address but signer misses secret key": {
			signer: func(ctrl *gomock.Controller) *hmacsignertest.MockHMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().Sign(net.ParseIP("127.0.0.1").To4()).Return(nil, hmacsigner.ErrMissingSecretKey)
				return m
			},
			remoteAddr: "127.0.0.1:23456",
			wantIP: httpheaders.XRealIP{
				IP: net.ParseIP("127.0.0.1").To4(),
			},
			wantErrString: "",
		},
		"valid remote address and signer configured with secret key": {
			signer: func(ctrl *gomock.Controller) *hmacsignertest.MockHMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().Sign(net.ParseIP("127.0.0.1").To4()).Return([]byte(`dummy hash`), nil)
				return m
			},
			remoteAddr: "127.0.0.1:23456",
			wantIP: httpheaders.XRealIP{
				IP:   net.ParseIP("127.0.0.1").To4(),
				Hash: []byte(`dummy hash`),
			},
			wantErrString: "",
		},
		"valid remote address but signer fails on signing": {
			signer: func(ctrl *gomock.Controller) *hmacsignertest.MockHMACSigner {
				m := hmacsignertest.NewMockHMACSigner(ctrl)
				m.EXPECT().Sign(net.ParseIP("127.0.0.1").To4()).Return(nil, fmt.Errorf("IP signing error"))
				return m
			},
			remoteAddr: "127.0.0.1:23456",
			wantIP: httpheaders.XRealIP{
				IP: net.ParseIP("127.0.0.1").To4(),
			},
			wantErrString: "IP signing error",
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			gotIP, err := prepareRealIPHeader(tc.remoteAddr, tc.signer(ctrl))

			if tc.wantErrString == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErrString)
			}

			assert.Equal(t, tc.wantIP, gotIP)
		})
	}

}

func BenchmarkGetOutgoingIPv4(b *testing.B) {
	var (
		ip  net.IP
		err error
	)

	b.ReportAllocs()

	for b.Loop() {
		ip, err = getOutgoingIPv4("localhost:12345")
		if err != nil {
			panic(err)
		}
	}

	fmt.Fprint(io.Discard, ip)
}
