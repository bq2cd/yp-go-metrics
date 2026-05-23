package httpheaders

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetXRealIPFromBytes(t *testing.T) {
	type testcase struct {
		input []byte
		want  XRealIP
	}

	tests := map[string]testcase{
		"empty input": {
			input: []byte{},
			want:  XRealIP{},
		},
		"invalid IP address": {
			input: []byte(`300.0.1234.500`),
			want:  XRealIP{},
		},
		"only IP address": {
			input: []byte(`8.9.10.11`),
			want: XRealIP{
				IP: net.ParseIP("8.9.10.11"),
			},
		},
		"IP address with spaces": {
			input: []byte(` 8 . 9.10.  11 `),
			want: XRealIP{
				IP: net.ParseIP("8.9.10.11"),
			},
		},
		"IP address with spaces and param delimiters": {
			input: []byte(` 8.9.10.11 ; ; `),
			want: XRealIP{
				IP: net.ParseIP("8.9.10.11"),
			},
		},
		"IP address with spaces and param delimiters before IP": {
			input: []byte(` ; 8.9.10.11 ; `),
			want:  XRealIP{},
		},
		"IP address with spaces and random parameter": {
			input: []byte(` 8.9.10.11 ; p1=123 `),
			want: XRealIP{
				IP: net.ParseIP("8.9.10.11"),
			},
		},
		"IP address with spaces and multiple random parameters": {
			input: []byte(` 8.9.10.11 ; p1=123 ; p2=456 `),
			want: XRealIP{
				IP: net.ParseIP("8.9.10.11"),
			},
		},
		"IP address with spaces and hash parameter (invalid hex)": {
			input: []byte(` 8.9.10.11 ; hash=zzz`),
			want: XRealIP{
				IP:   net.ParseIP("8.9.10.11"),
				Hash: []byte{},
			},
		},
		"IP address with spaces and hash parameter (incomplete hex)": {
			input: []byte(` 8.9.10.11 ; hash=123456789`),
			want: XRealIP{
				IP:   net.ParseIP("8.9.10.11"),
				Hash: []byte("\x124Vx"),
			},
		},
		"IP address with spaces and hash parameter (valid hex)": {
			input: []byte(` 8.9.10.11 ; hash =  484d4143207369676e6174757265 ; `),
			want: XRealIP{
				IP:   net.ParseIP("8.9.10.11"),
				Hash: []byte(`HMAC signature`),
			},
		},
		"IP address with spaces and hash parameter (valid hex) the first": {
			input: []byte(` 8.9.10.11 ; hash=484d4143207369676e6174757265 ; other=123 ; `),
			want: XRealIP{
				IP:   net.ParseIP("8.9.10.11"),
				Hash: []byte(`HMAC signature`),
			},
		},
		"IP address with spaces and hash parameter (valid hex) but not the first": {
			input: []byte(` 8.9.10.11 ; other=123 ; hash=484d4143207369676e6174757265 ; extra=456 `),
			want: XRealIP{
				IP:   net.ParseIP("8.9.10.11"),
				Hash: []byte(`HMAC signature`),
			},
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			realIP := GetXRealIPFromBytes(tc.input)

			assert.Equal(t, tc.want, realIP)
		})
	}
}

func TestXRealIP_String(t *testing.T) {
	type testcase struct {
		realIP     XRealIP
		wantString string
	}

	tests := map[string]testcase{
		"empty real IP -> empty string": {
			realIP:     XRealIP{},
			wantString: "",
		},
		"invalid IP address -> empty string": {
			realIP: XRealIP{
				IP: net.ParseIP("300.0.0.300"),
			},
			wantString: "",
		},
		"only IP address -> no hash parameter": {
			realIP: XRealIP{
				IP: net.ParseIP("192.168.0.1"),
			},
			wantString: "192.168.0.1",
		},
		"IP address + hash -> hash parameter present": {
			realIP: XRealIP{
				IP:   net.ParseIP("192.168.0.1"),
				Hash: []byte("abcd"),
			},
			wantString: "192.168.0.1;hash=61626364", // abcd -> 61626364 (in hex)
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			out := tc.realIP.String()

			assert.Equal(t, tc.wantString, out)
		})
	}
}

func TestXRealIP_Matches(t *testing.T) {
	type testcase struct {
		realIP XRealIP
		header http.Header
		want   bool
	}

	tests := map[string]testcase{
		"empty header matches empty real IP": {
			realIP: XRealIP{},
			header: http.Header{},
			want:   true,
		},
		"header with invalid IP does not match real IP": {
			realIP: XRealIP{
				IP: net.ParseIP("10.0.0.1"),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, "500.0.1000.2.222")
				return h
			}(),
			want: false,
		},
		"header with different IP does not match real IP": {
			realIP: XRealIP{
				IP: net.ParseIP("10.0.0.1"),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, "10.0.0.2")
				return h
			}(),
			want: false,
		},
		"header with the same IP matches real IP": {
			realIP: XRealIP{
				IP: net.ParseIP("10.0.0.1"),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, "  10.0 . 0.1 ")
				return h
			}(),
			want: true,
		},
		"header with the same IP but without hash does not match real IP": {
			realIP: XRealIP{
				IP:   net.ParseIP("10.0.0.1"),
				Hash: []byte(`10-0-0-1`),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, "10.0.0.1")
				return h
			}(),
			want: false,
		},
		"header with the same IP but different hash does not match real IP": {
			realIP: XRealIP{
				IP:   net.ParseIP("10.0.0.1"),
				Hash: []byte(`10-0-0-1`),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, " 10.0.0.1 ; hash = 31302d302d302d32") // 10-0-0-2
				return h
			}(),
			want: false,
		},
		"header with the same IP and same hash matches real IP": {
			realIP: XRealIP{
				IP:   net.ParseIP("10.0.0.1"),
				Hash: []byte(`10-0-0-1`),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, " 10. 0.0.1 ; hash =  31302d302d302d31;; ") // 10-0-0-1
				return h
			}(),
			want: true,
		},
		"header with the same IP and same hash but not the first position matches real IP": {
			realIP: XRealIP{
				IP:   net.ParseIP("10.0.0.1"),
				Hash: []byte(`10-0-0-1`),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, "10.0.0.1;p=1;;hash=31302d302d302d31") // 10-0-0-1
				return h
			}(),
			want: true,
		},
		"header with the same IP and two hashes with the first valid hash matches real IP": {
			realIP: XRealIP{
				IP:   net.ParseIP("10.0.0.1"),
				Hash: []byte(`10-0-0-1`),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, "10.0.0.1; p=1; hash = 31302d302d302d31 ; hash=31302d302d302d32; p=2") // 10-0-0-1, 10-0-0-2
				return h
			}(),
			want: true,
		},
		"header with the same IP and two hashes with the first invalid hash does not match real IP": {
			realIP: XRealIP{
				IP:   net.ParseIP("10.0.0.1"),
				Hash: []byte(`10-0-0-1`),
			},
			header: func() http.Header {
				h := http.Header{}
				h.Add(HeaderKeyXRealIP, "10.0.0.1; p=1; hash=31302d302d302d32; hash=31302d302d302d31; p=2") // 10-0-0-2, 10-0-0-1
				return h
			}(),
			want: false,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			equal := tc.realIP.Matches(tc.header)

			assert.Equalf(t, tc.want, equal, "real IP does not match HTTP header")
		})
	}
}
