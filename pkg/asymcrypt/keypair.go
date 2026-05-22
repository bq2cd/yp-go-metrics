package asymcrypt

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
)

// X25519KeyPair contains public/private X25519 keys, encoded in PEM format.
type X25519KeyPair struct {
	Public  []byte
	Private []byte
}

// NewX25519KeyPair generates a new X25519 key pair and encodes them into PEM format.
func NewX25519KeyPair() (*X25519KeyPair, error) {
	kp := new(X25519KeyPair)

	private, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("cannot generate X25519 private key: %w", err)
	}

	buf := bytes.NewBuffer(nil)

	err = EncodePrivateKey(buf, private)
	if err != nil {
		return nil, err
	}

	kp.Private = bytes.Clone(buf.Bytes())

	buf.Reset()

	err = EncodePublicKey(buf, private.PublicKey())
	if err != nil {
		return nil, err
	}

	kp.Public = bytes.Clone(buf.Bytes())

	return kp, nil
}

// EncodePrivateKey marshals key into PKCS8, ASN.1 DER form and encodes it into PEM format.
// The encoded key is written into provided [io.Writer].
func EncodePrivateKey(w io.Writer, key *ecdh.PrivateKey) error {
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("cannot marshal X25519 private key: %w", err)
	}

	err = pem.Encode(w, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8,
	})
	if err != nil {
		return fmt.Errorf("cannot PEM-encode private key: %w", err)
	}

	return nil
}

// EncodePublicKey marshals key into PKIX, ASN.1 DER form and encodes it into PEM format.
// The encoded key is written into provided [io.Writer].
func EncodePublicKey(w io.Writer, pubkey *ecdh.PublicKey) error {
	pkcs8, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return fmt.Errorf("cannot marshal X25519 public key: %w", err)
	}

	err = pem.Encode(w, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkcs8,
	})
	if err != nil {
		return fmt.Errorf("cannot PEM-encode public key: %w", err)
	}

	return nil
}
