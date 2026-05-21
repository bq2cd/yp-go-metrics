package asymcrypt

import (
	"crypto/ecdh"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	// ErrEmptyKeyBytes is returned for empty input.
	ErrEmptyKeyBytes = errors.New("key bytes are missing")
	// ErrInvalidKeyEncoding is returned when input is not PEM-encoded.
	ErrInvalidKeyEncoding = errors.New("invalid key encoding (not PEM)")
	// ErrInvalidKeyType is returned when parsed key is of different type or has unsupported curve (not X25519).
	ErrInvalidKeyType = errors.New("invalid key type or unsupported curve (not X25519)")
)

// ParsePrivateKey parses PEM-encoded unencrypted private key in PKCS #8, ASN.1 DER form.
// It expects the key to be X25519 private key, otherwise it will return [ErrInvalidKeyType] error.
func ParsePrivateKey(input []byte) (*ecdh.PrivateKey, error) {
	if len(input) == 0 {
		return nil, ErrEmptyKeyBytes
	}

	block, _ := pem.Decode(input)
	if block == nil {
		return nil, ErrInvalidKeyEncoding
	}

	cand, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := cand.(*ecdh.PrivateKey)
	if !ok {
		return nil, ErrInvalidKeyType
	}

	if key.Curve() != ecdh.X25519() {
		return nil, ErrInvalidKeyType
	}

	return key, nil
}

// ParsePublicKey parses PEM-encoded public key in PKIX, ASN.1 DER form.
// It expects the key to be X25519 public key, otherwise it will return [ErrInvalidKeyType] error.
func ParsePublicKey(input []byte) (*ecdh.PublicKey, error) {
	if len(input) == 0 {
		return nil, ErrEmptyKeyBytes
	}

	block, _ := pem.Decode(input)
	if block == nil {
		return nil, ErrInvalidKeyEncoding
	}

	cand, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := cand.(*ecdh.PublicKey)
	if !ok {
		return nil, ErrInvalidKeyType
	}

	if key.Curve() != ecdh.X25519() {
		return nil, ErrInvalidKeyType
	}

	return key, nil
}
