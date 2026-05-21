package asymcrypt_test

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
)

var (
	benchmarkMessage32 = bytes.Repeat([]byte(`1234`), 8) // 32 bytes (e.g. key length for AES-256)
)

func TestEncryptDecrypt(t *testing.T) {
	type testcase struct {
		cleartext      []byte
		wantEncryptErr error
		wantDecryptErr error
	}

	tests := map[string]testcase{
		"empty input": {
			cleartext: nil,
		},
		"input < 100 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 10), // 90 bytes
		},
		"input < 400 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 40), // 360 bytes
		},
		"input < 500 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 50), // 450 bytes
		},
		"input < 1000 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 100), // 900 bytes
		},
		"input < 5000 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 500), // 4500 bytes
		},
		"input < 10_000 bytes": {
			cleartext: bytes.Repeat([]byte(`1 2 3 4 5`), 1_000), // 9000 bytes
		},
	}

	static, err := ecdh.X25519().GenerateKey(rand.Reader)
	require.NoErrorf(t, err, "cannot generate X25519 key pair")

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			enc, err := asymcrypt.NewEncryptor(static.PublicKey())
			require.NoErrorf(t, err, "encryptor creation failed")

			ciphertext, err := enc.Encrypt(tc.cleartext)
			if tc.wantEncryptErr == nil {
				require.NoErrorf(t, err, "cannot encrypt")
			} else {
				require.ErrorIsf(t, err, tc.wantEncryptErr, "unexpected encryption error")
			}

			if err != nil { // when encryption fails, no point in attempting decryption
				return
			}

			assert.NotEqualf(t, tc.cleartext, ciphertext, "cleartext not encrypted")

			dec := asymcrypt.NewDecryptor(static)

			cleartext, err := dec.Decrypt(ciphertext)
			if tc.wantDecryptErr == nil {
				require.NoErrorf(t, err, "cannot decrypt")
			} else {
				require.ErrorIsf(t, err, tc.wantDecryptErr, "unexpected decrypte error")
			}

			assert.Equalf(t, tc.cleartext, cleartext, "unexpected decrypted value")

			t.Logf("ciphertext size: %d, cleartext size: %d", len(ciphertext), len(cleartext))
		})
	}
}

func BenchmarkEncrypt_32bytes(b *testing.B) {
	msg := bytes.Clone(benchmarkMessage32)

	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	enc, err := asymcrypt.NewEncryptor(key.PublicKey())
	if err != nil {
		panic(err)
	}

	var out []byte

	b.Run("", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			out, err = enc.Encrypt(msg)
			if err != nil {
				panic(err)
			}
		}

		b.ReportMetric(float64(len(out)), "outbytes")
	})

	fmt.Fprint(io.Discard, string(out))
}

func BenchmarkDecrypt_32bytes(b *testing.B) {
	msg := bytes.Clone(benchmarkMessage32)

	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	enc, err := asymcrypt.NewEncryptor(key.PublicKey())
	if err != nil {
		panic(err)
	}

	ciphertext, err := enc.Encrypt(msg)
	if err != nil {
		panic(err)
	}

	dec := asymcrypt.NewDecryptor(key)

	var out []byte

	b.Run("", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			out, err = dec.Decrypt(ciphertext)
			if err != nil {
				panic(err)
			}
		}

		b.ReportMetric(float64(len(out)), "outbytes")
	})

	fmt.Fprint(io.Discard, string(out))
}
