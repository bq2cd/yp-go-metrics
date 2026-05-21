package asymcrypt_test

import (
	"crypto/ecdh"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
)

func TestParsePrivateKey(t *testing.T) {
	type testcase struct {
		input         []byte
		wantErr       error
		wantErrString string
	}

	tests := map[string]testcase{
		"empty input": {
			input:   []byte{},
			wantErr: asymcrypt.ErrEmptyKeyBytes,
		},
		"invalid key encoding (not PEM)": {
			input: []byte(`
MC4CAQAwBQYDK2VuBCIEIBOrVDQQCTds+Z4foo8YwVQQlW3qn/iYqWBYB9NEYInO
				`),
			wantErr: asymcrypt.ErrInvalidKeyEncoding,
		},
		"invalid key format (random bytes)": {
			input: []byte(`
-----BEGIN PRIVATE KEY-----
uREkYiIx5Cxxl4x+vSGpP+ywiJXO2EKX3N+iroWYMJEsF6mTBX2+Ik/Od1TQz3Yk
sAoO2v32yxNugU8yr2w/zw==
-----END PRIVATE KEY-----
				`),
			wantErrString: "asn1: structure error: tags don't match",
		},
		"invalid key type (ecdsa)": {
			input: []byte(`
-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgNn/D2nplofWr0GK7
aodTxCXbCGT+09wwwyXamsGlSIShRANCAASFIxULepm9T1APxjyKX3QhsksbNkw+
G+4Kcv+T/7WABg+lJ5euEmY1ah6F273QS4E1GsLvECsOugAE3dovgWOT
-----END PRIVATE KEY-----
				`),
			wantErr: asymcrypt.ErrInvalidKeyType,
		},
		"invalid key curve (NIST P-256)": {
			input: []byte(`
-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgYF0t32NP5bg/m7f7
nJafOdzDuni18F02tzW+OvBA67GhRANCAARxAn5vNMOCXADz94HLypEuFFkKkWiD
A3GpCRO+lbL3hvzj8xlqfU5GzCJ2kuDxJXf4i/bBObgDiifRiMWT7K1a
-----END PRIVATE KEY-----
				`),
			wantErr: asymcrypt.ErrInvalidKeyType,
		},
		"valid X25519 key": {
			input: []byte(`
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VuBCIEIFbnyK1qIh28hRwoWA8yqZTgbIRTzjqG9tlg1cSUer9K
-----END PRIVATE KEY-----
				`),
			wantErr: nil,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			key, err := asymcrypt.ParsePrivateKey(tc.input)

			if tc.wantErr == nil {
				if tc.wantErrString == "" {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, tc.wantErrString)
				}
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			if err != nil {
				return
			}

			require.NotNilf(t, key, "expected non-nil key")

			assert.Equalf(t, ecdh.X25519(), key.Curve(), "incorrect curve type")
		})
	}
}

func TestParsePublicKey(t *testing.T) {
	type testcase struct {
		input         []byte
		wantErr       error
		wantErrString string
	}

	tests := map[string]testcase{
		"empty input": {
			input:   []byte{},
			wantErr: asymcrypt.ErrEmptyKeyBytes,
		},
		"invalid key encoding (not PEM)": {
			input: []byte(`
MCowBQYDK2VuAyEA5m3YMw1LGYjn3wjp3+xN6PkQ6nCFwL6aTCN6ey0+NWU=
				`),
			wantErr: asymcrypt.ErrInvalidKeyEncoding,
		},
		"invalid key format (random bytes)": {
			input: []byte(`
-----BEGIN PUBLIC KEY-----
Lyuaa7cA2ow81gjYGUycXEZFdxukwBIlEJaxbBHeoos=
-----END PUBLIC KEY-----
				`),
			wantErrString: "asn1: structure error: tags don't match",
		},
		"invalid key type (ecdsa)": {
			input: []byte(`
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEl33pLmqVVaqZkIm9CfBQgIutPVtC
6xL2OcODgsVbTMWAUPFpov5RDjCENbPkLFm5mzWkpXXABMV8WGeb2/lUpQ==
-----END PUBLIC KEY-----
				`),
			wantErr: asymcrypt.ErrInvalidKeyType,
		},
		"invalid key curve (NIST P-256)": {
			input: []byte(`
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQpESANEhy93RTS9E5v/UKSXESiwO
trnoLrONMH4Ogu3G6cfPsf9eaDP1KPVntGeqZTf8dlCBn5G0FRBIOJPPDA==
-----END PUBLIC KEY-----
				`),
			wantErr: asymcrypt.ErrInvalidKeyType,
		},
		"valid X25519 key": {
			input: []byte(`
-----BEGIN PUBLIC KEY-----
MCowBQYDK2VuAyEAU2Ps2cuU7yMrHHCV+FZJC0QkGowFW8oq2bqo2kj1wVI=
-----END PUBLIC KEY-----
				`),
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			key, err := asymcrypt.ParsePublicKey(tc.input)

			if tc.wantErr == nil {
				if tc.wantErrString == "" {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, tc.wantErrString)
				}
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			if err != nil {
				return
			}

			require.NotNilf(t, key, "expected non-nil key")

			assert.Equalf(t, ecdh.X25519(), key.Curve(), "incorrect curve type")
		})
	}
}
