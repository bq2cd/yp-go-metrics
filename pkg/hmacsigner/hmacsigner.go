package hmacsigner

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

var (
	// ErrMissingSecretKey is returned by [HMACSigner.Sign] if there is not secret key configured.
	ErrMissingSecretKey = errors.New("secret key is missing")
	// ErrSignatureMismatch is returned by [HMACSigner.Verify] if presumably signed data does not match expected signature.
	ErrSignatureMismatch = errors.New("signature mismatch")
)

//go:generate go tool mockgen -destination=hmacsignertest/mock_hmacsigner.go -package hmacsignertest github.com/bq2cd/yp-go-metrics/pkg/hmacsigner HMACSigner

// HMACSigner performs signing/verification of a given message.
type HMACSigner interface {
	Sign(signingMessage []byte) ([]byte, error)
	Verify(signedMessage []byte, signature []byte) error
	HasKey() bool
}

type hmacSigner struct {
	secretKey []byte
}

// NewHMACSigner creates an instance of HMAC signer with the given secret key.
func NewHMACSigner(secretKey []byte) *hmacSigner {
	return &hmacSigner{
		secretKey: secretKey,
	}
}

// Sign calculates HMAC-SHA256 hash on the provided message.
// It will return [ErrMissingSecretKey] if the key is missing.
func (hs *hmacSigner) Sign(signingMessage []byte) ([]byte, error) {
	if !hs.HasKey() {
		return nil, ErrMissingSecretKey
	}
	if len(signingMessage) == 0 {
		return []byte{}, nil
	}
	hasher := hmac.New(sha256.New, hs.secretKey)
	// as per documentation, [hash.Hash.Write] never returns an error.
	hasher.Write(signingMessage)
	return hasher.Sum(nil), nil
}

// Verify calculates HMAC-SHA256 hash on the provided message and compares it with the provided signature.
// If calculated hash matches signature, `nil` is returned.
// It will return [ErrMissingSecretKey] if the key is missing.
func (hs *hmacSigner) Verify(signedMessage []byte, expectedSignature []byte) error {
	if !hs.HasKey() {
		return ErrMissingSecretKey
	}
	if len(signedMessage) == 0 {
		return nil
	}
	hasher := hmac.New(sha256.New, hs.secretKey)
	// as per documentation, [hash.Hash.Write] never returns an error.
	hasher.Write(signedMessage)
	if !hmac.Equal(expectedSignature, hasher.Sum(nil)) {
		return ErrSignatureMismatch
	}
	return nil
}

func (hs *hmacSigner) HasKey() bool {
	return len(hs.secretKey) > 0
}
