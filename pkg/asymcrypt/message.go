package asymcrypt

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const (
	messagePublicKeySize = 32
	messageSaltSize      = 16
	messageNonceSize     = 12
	messageHeaderSize    = messagePublicKeySize + messageSaltSize + messageNonceSize
)

var (
	// ErrMessageTooShort is returned when wire data length is less than [messageHeaderSize].
	ErrMessageTooShort = errors.New("encrypted message too short")
	// ErrEmptyCipherText is returned when wire data has only message header (but no ciphertext).
	ErrEmptyCipherText = errors.New("empty ciphertext")
)

// Message describes decomposed wire data, consisting of ciphertext and auxiliary info (header)
// required for decryption.
type Message struct {
	publicKey  []byte
	salt       []byte
	nonce      []byte
	ciphertext []byte
}

// MessageFromWire takes bytes from wire and attempts to decompose them into [Message].
// Important: as there is no copying, so the same backing array is shared between wire `data` and
// the resulting [Message] parts.
func MessageFromWire(data []byte) (*Message, error) {
	if len(data) < messageHeaderSize {
		return nil, ErrMessageTooShort
	}

	msg := &Message{
		publicKey:  extractPrefix(&data, messagePublicKeySize),
		salt:       extractPrefix(&data, messageSaltSize),
		nonce:      extractPrefix(&data, messageNonceSize),
		ciphertext: data,
	}

	if len(msg.ciphertext) == 0 {
		return nil, ErrEmptyCipherText
	}

	return msg, nil
}

// ToWire composes wire data from message by concatenating its parts.
func (m *Message) ToWire() []byte {
	out := make([]byte, len(m.publicKey)+len(m.salt)+len(m.nonce)+len(m.ciphertext))

	pos := 0
	for _, data := range [][]byte{
		m.publicKey,
		m.salt,
		m.nonce,
		m.ciphertext,
	} {
		pos += copy(out[pos:pos+len(data)], data)
	}

	return out
}

// InitSalt generates [messageSaltSize] random bytes using [crypto/rand.Reader] and updates
// message's `salt` field in place.
func (m *Message) InitSalt() error {
	m.salt = make([]byte, messageSaltSize)

	_, err := io.ReadFull(rand.Reader, m.salt)
	if err != nil {
		return err
	}

	return nil
}

// InitNonce generates [messageNonceSize] random bytes using [crypto/rand.Reader] and updates
// message's `nonce` field in place.
func (m *Message) InitNonce() error {
	m.nonce = make([]byte, messageNonceSize)

	_, err := io.ReadFull(rand.Reader, m.nonce)
	if err != nil {
		return err
	}

	return nil
}

// Seal encrypts and authenticates clear text using provided [cipher.AEAD].
// The resulting ciphertext is stored in message's `ciphertext` field.
func (m *Message) Seal(aead cipher.AEAD, cleartext []byte) {
	m.ciphertext = aead.Seal(m.ciphertext, m.nonce, cleartext, nil)
}

// Open decrypts and authenticates message's ciphertext using provided [cipher.AEAD].
func (m *Message) Open(aead cipher.AEAD) ([]byte, error) {
	return aead.Open(nil, m.nonce, m.ciphertext, nil)
}

func extractPrefix(data *[]byte, size int) []byte {
	out := (*data)[:size]
	*data = (*data)[size:]

	return out
}
