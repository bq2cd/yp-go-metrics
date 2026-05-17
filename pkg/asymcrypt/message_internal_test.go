package asymcrypt

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageFromWire(t *testing.T) {
	type testcase struct {
		wiredata []byte
		wantMsg  *Message
		wantErr  error
	}

	tests := map[string]testcase{
		"empty wire data": {
			wiredata: []byte{},
			wantMsg:  nil,
			wantErr:  ErrMessageTooShort,
		},
		"wire data too short": {
			wiredata: []byte(`1 2 3 4 5`),
			wantMsg:  nil,
			wantErr:  ErrMessageTooShort,
		},
		"wire data contains header only": {
			wiredata: bytes.Repeat([]byte(`1 2 3`), 12), // 60 bytes, exactly size of the header
			wantMsg:  nil,
			wantErr:  ErrEmptyCipherText,
		},
		"wire data with cipher text": {
			wiredata: bytes.Repeat([]byte(`abcD`), 20), // 80 bytes
			wantMsg: &Message{
				publicKey:  bytes.Repeat([]byte(`abcD`), 8), // 32 bytes
				salt:       bytes.Repeat([]byte(`abcD`), 4), // 16 bytes
				nonce:      bytes.Repeat([]byte(`abcD`), 3), // 12 bytes
				ciphertext: bytes.Repeat([]byte(`abcD`), 5), // 20 bytes
			},
			wantErr: nil,
		},
		"wire data with cipher text (random)": func() testcase {
			key := make([]byte, messagePublicKeySize)
			salt := make([]byte, messageSaltSize)
			nonce := make([]byte, messageNonceSize)
			ciphertext := make([]byte, 100)

			for _, buf := range [][]byte{key, salt, nonce, ciphertext} {
				rand.Read(buf) // rand.Read never returns an error but crashes instead
			}

			return testcase{
				wiredata: slices.Concat(key, salt, nonce, ciphertext),
				wantMsg: &Message{
					publicKey:  key,
					salt:       salt,
					nonce:      nonce,
					ciphertext: ciphertext,
				},
				wantErr: nil,
			}
		}(),
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			t.Logf("wire data size: %d bytes", len(tc.wiredata))

			msg, err := MessageFromWire(tc.wiredata)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			if tc.wantMsg == nil {
				assert.Nilf(t, msg, "expected <nil> instead of message")
			} else {
				assert.Equal(t, tc.wantMsg, msg)
			}
		})
	}
}

func TestMessageToWire(t *testing.T) {
	type testcase struct {
		msg      Message
		wantWire []byte
	}

	tests := map[string]testcase{
		"empty message produces empty wire data": {
			msg:      Message{},
			wantWire: []byte{},
		},
		"message without cipher text": {
			msg: Message{
				publicKey: []byte(`public key;`),
				salt:      []byte(`salt;`),
				nonce:     []byte(`nonce;`),
			},
			wantWire: []byte(`public key;salt;nonce;`),
		},
		"message with cipher text": {
			msg: Message{
				publicKey:  []byte(`public key;`),
				salt:       []byte(`salt;`),
				nonce:      []byte(`nonce;`),
				ciphertext: []byte(`cipher text`),
			},
			wantWire: []byte(`public key;salt;nonce;cipher text`),
		},
		"message with cipher text (random)": func() testcase {
			key := make([]byte, 100)
			salt := make([]byte, 90)
			nonce := make([]byte, 70)
			ciphertext := make([]byte, 1000)

			for _, buf := range [][]byte{key, salt, nonce, ciphertext} {
				rand.Read(buf) // rand.Read never returns error but crashes instead
			}

			return testcase{
				msg: Message{
					publicKey:  key,
					salt:       salt,
					nonce:      nonce,
					ciphertext: ciphertext,
				},
				wantWire: slices.Concat(key, salt, nonce, ciphertext),
			}
		}(),
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			data := tc.msg.ToWire()

			assert.Equal(t, tc.wantWire, data)
		})
	}
}

func TestMessageInitSalt(t *testing.T) {
	type testcase struct {
		initial []byte
	}

	tests := map[string]testcase{
		"empty salt": {
			initial: []byte{},
		},
		"prepopulated salt": {
			initial: []byte(`1 2 3 4 5`),
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			msg := Message{
				salt: tc.initial,
			}

			msg.InitSalt()

			assert.Lenf(t, msg.salt, messageSaltSize, "incorrect salt size")
			assert.NotEqualf(t, tc.initial, msg.salt, "salt has not changed")
		})
	}
}

func TestMessageInitNonce(t *testing.T) {
	type testcase struct {
		initial []byte
	}

	tests := map[string]testcase{
		"empty nonce": {
			initial: []byte{},
		},
		"prepopulated nonce": {
			initial: []byte(`1 2 3 4 5`),
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			msg := Message{
				nonce: tc.initial,
			}

			msg.InitNonce()

			assert.Lenf(t, msg.nonce, messageNonceSize, "incorrect nonce size")
			assert.NotEqualf(t, tc.initial, msg.nonce, "nonce has not changed")
		})
	}
}

func TestMessageSeal(t *testing.T) {
	aead := new(mockAEAD)
	msg := new(Message)

	msg.InitNonce()

	cleartext := []byte(`1 2 3 4 5`)
	want := slices.Concat([]byte(`mock!`), msg.nonce, cleartext)

	msg.Seal(aead, cleartext)

	assert.Equalf(t, want, msg.ciphertext, "message ciphertext is not sealed")
}

func TestMessageOpen(t *testing.T) {
	type testcase struct {
		aead          *mockAEAD
		ciphertext    []byte
		wantClearText []byte
		wantErr       error
	}

	tests := map[string]testcase{
		"decryption succeeds": {
			aead: &mockAEAD{
				wantClearText: []byte(`some clear text`),
			},
			ciphertext:    []byte(`1 2 3 4 5`),
			wantClearText: []byte(`some clear text`),
			wantErr:       nil,
		},
		"decryption fails": func() testcase {
			err := errors.New("some decryption error")
			return testcase{
				aead: &mockAEAD{
					wantOpenErr: err,
				},
				ciphertext:    []byte(`1 2 3 4 5`),
				wantClearText: nil,
				wantErr:       err,
			}
		}(),
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			msg := &Message{
				ciphertext: tc.ciphertext,
			}

			msg.InitNonce()

			cleartext, err := msg.Open(tc.aead)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			assert.Equalf(t, tc.wantClearText, cleartext, "incorrect clear text returned by Open")
		})
	}
}

type mockAEAD struct {
	wantOpenErr   error
	wantClearText []byte
}

// Seal mimics behavior of [crypto/cipher.AEAD.Seal] method, but without any encryption/authentication.
func (m *mockAEAD) Seal(dst, nonce, cleartext, additionalData []byte) []byte {
	dst = append(dst, []byte(`mock!`)...)
	dst = append(dst, nonce...)
	dst = append(dst, cleartext...)
	dst = append(dst, additionalData...)

	return dst
}

func (m *mockAEAD) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	if m.wantOpenErr != nil {
		return nil, fmt.Errorf("open error: %w", m.wantOpenErr)
	}

	return m.wantClearText, nil
}

func (m *mockAEAD) NonceSize() int {
	return messageNonceSize
}

func (m *mockAEAD) Overhead() int {
	return 0
}
