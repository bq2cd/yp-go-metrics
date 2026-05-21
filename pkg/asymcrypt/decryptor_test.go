package asymcrypt_test

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
)

func TestDecryptFailure(t *testing.T) {
	type testcase struct {
		ciphertext    []byte
		wantErrString string
	}

	tests := map[string]testcase{
		"empty ciphertext": {
			ciphertext:    []byte{},
			wantErrString: "asymcrypt.Decryptor(wire):",
		},
		"public key too short": {
			ciphertext:    bytes.Repeat([]byte(`abc`), 10), // 30 bytes < 32 bytes (X25519 key size)
			wantErrString: "asymcrypt.Decryptor(wire):",
		},
		"incomplete salt": {
			ciphertext:    bytes.Repeat([]byte(`abc`), 15), // 45 bytes < 32 + 16 bytes (X25519 key size + salt size)
			wantErrString: "asymcrypt.Decryptor(wire):",
		},
		"missing encrypted body": {
			ciphertext:    bytes.Repeat([]byte(`abc`), 20), // 60 bytes, exactly header size
			wantErrString: "asymcrypt.Decryptor(wire):",
		},
		"invalid decryption params": {
			ciphertext:    bytes.Repeat([]byte(`abc`), 30), // 90 bytes, header + 30 bytes of encrypted body
			wantErrString: "asymcrypt.Decryptor(open):",
		},
	}

	static, err := ecdh.X25519().GenerateKey(rand.Reader)
	require.NoErrorf(t, err, "cannot generate X25519 key pair")

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			dec := asymcrypt.NewDecryptor(static)

			_, err := dec.Decrypt(tc.ciphertext)

			t.Logf("decrypt error: %v (%T)", err, err)

			require.NotEmptyf(t, tc.wantErrString, "wantErrString must not be empty")
			require.ErrorContainsf(t, err, tc.wantErrString, "unexpected decryption error")
		})
	}
}
