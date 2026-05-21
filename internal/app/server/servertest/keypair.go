package servertest

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// X25519KeyPair contains public/private X25519 keys, encoded in PEM format.
type X25519KeyPair struct {
	Public  *bytes.Buffer
	Private *bytes.Buffer
}

// NewX25519KeyPair generates a new X25519 key pair and encodes them into PEM format.
func NewX25519KeyPair() (*X25519KeyPair, error) {
	private, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("cannot generate X25519 private key: %w", err)
	}

	marshalPrivate, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal X25519 private key: %w", err)
	}

	marshalPublic, err := x509.MarshalPKIXPublicKey(private.PublicKey())
	if err != nil {
		return nil, fmt.Errorf("cannot marshal X25519 public key: %w", err)
	}

	kp := &X25519KeyPair{
		Public:  bytes.NewBuffer(nil),
		Private: bytes.NewBuffer(nil),
	}

	err = pem.Encode(kp.Private, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: marshalPrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot PEM-encode private key: %w", err)
	}

	err = pem.Encode(kp.Public, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: marshalPublic,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot PEM-encode public key: %w", err)
	}

	return kp, nil
}
