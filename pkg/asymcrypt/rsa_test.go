package asymcrypt_test

// Test RSA asymmetric encryption with regards to payload size.

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rsaTestCase struct {
	cleartext      []byte
	wantEncryptErr map[int]error
	wantDecryptErr map[int]error
}

func TestRSAEncryptDecryptPayloadSize(t *testing.T) {
	tests := map[string]rsaTestCase{
		"empty input": {
			cleartext: []byte{},
		},
		"input < 100 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 10), // 90 bytes
		},
		"input < 200 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 20), // 180 bytes
		},
		"input < 400 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 40), // 360 bytes
			wantEncryptErr: map[int]error{
				2048: rsa.ErrMessageTooLong,
			},
		},
		"input < 500 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 50), // 450 bytes
			wantEncryptErr: map[int]error{
				2048: rsa.ErrMessageTooLong,
				4096: rsa.ErrMessageTooLong,
			},
		},
		"input < 1000 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 100), // 900 bytes
			wantEncryptErr: map[int]error{
				2048: rsa.ErrMessageTooLong,
				4096: rsa.ErrMessageTooLong,
			},
		},
	}

	keys := map[int]*rsa.PrivateKey{
		2048: generateRSAKey(t, 2048),
		4096: generateRSAKey(t, 4096),
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			for bits, key := range keys {
				t.Run(fmt.Sprintf("bits=%d", bits), func(t *testing.T) {
					runRSATestCase(t, tc, bits, key)
				})
			}
		})
	}
}

func generateRSAKey(t *testing.T, bits int) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)

	return key
}

func runRSATestCase(t *testing.T, tc rsaTestCase, bits int, key *rsa.PrivateKey) {
	t.Helper()

	t.Logf("cleartext size: %d bytes, key modulus size: %d bytes", len(tc.cleartext), key.Size())

	// encrypt
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, tc.cleartext, nil)
	if tc.wantEncryptErr[bits] == nil {
		require.NoErrorf(t, err, "unexpected encryption error")
	} else {
		require.ErrorIsf(t, err, tc.wantEncryptErr[bits], "unexpected encryption error")
	}

	if err != nil { // encryption failed, no point in attempting decryption
		return
	}

	t.Logf("ciphertext size: %d bytes", len(ciphertext))

	// decrypt
	cleartext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, ciphertext, nil)
	if tc.wantDecryptErr[bits] == nil {
		require.NoErrorf(t, err, "unexpected decryption error")
	} else {
		require.ErrorIsf(t, err, tc.wantDecryptErr[bits], "unexpected decryption error")
	}

	// assert
	assert.Equalf(t, tc.cleartext, cleartext, "decrypted cleartext does not match original cleartext")
}

func BenchmarkRSAEncrypt_32bytes(b *testing.B) {
	msg := bytes.Clone(benchmarkMessage32)

	for _, bits := range []int{2048, 3072, 4096} {
		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			panic(err)
		}

		var out []byte

		b.Run(fmt.Sprintf("bits=%d", bits), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				out, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, msg, nil)
				if err != nil {
					panic(err)
				}
			}

			b.ReportMetric(float64(len(out)), "outbytes")
		})

		fmt.Fprint(io.Discard, string(out))
	}
}

func BenchmarkRSADecrypt_32bytes(b *testing.B) {
	msg := bytes.Clone(benchmarkMessage32)

	for _, bits := range []int{2048, 3072, 4096} {
		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			panic(err)
		}

		ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, msg, nil)
		if err != nil {
			panic(err)
		}

		var out []byte

		b.Run(fmt.Sprintf("bits=%d", bits), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				out, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, key, ciphertext, nil)
				if err != nil {
					panic(err)
				}
			}

			b.ReportMetric(float64(len(out)), "outbytes")
		})

		fmt.Fprint(io.Discard, string(out))
	}
}
