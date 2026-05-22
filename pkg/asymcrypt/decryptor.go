package asymcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"
)

//go:generate go tool mockgen -destination=asymcrypttest/mock_decryptor.go -package asymcrypttest github.com/bq2cd/yp-go-metrics/pkg/asymcrypt Decryptor

// Decryptor defines an entity capable of decrypting provided ciphertext into clear text data.
type Decryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

type decryptor struct {
	local *ecdh.PrivateKey
}

// NewDecryptor creates an instance of [Decryptor] that implements AES-GCM decryption using
// provided static X25519 private key, ECDH and HKDF to derive a symmetric encryption key.
// For ECDH to work, [Decryptor] receives sender's X25519 public key as part of an encrypted message.
// For HKDF to work, a salt is provided as part of the encrypted message.
// For AES-GCM to work, a nonce (initialization vector) is provided as part of the encrypted message.
func NewDecryptor(local *ecdh.PrivateKey) *decryptor {
	return &decryptor{
		local: local,
	}
}

// Decrypt takes provided ciphertext and decrypts it with AES-GCM, deriving symmetric encryption key
// from local (static, provided externally) X25519 private key and a remote X25519 public key (ephemeral,
// provided as port of the ciphertext) using ECDH and HKDF.
// HKDF salt and AES-GCM nonce are also provided as part of the ciphertext.
func (d *decryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	msg, err := MessageFromWire(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Decryptor(wire): %w", err)
	}

	remote, err := ecdh.X25519().NewPublicKey(msg.publicKey)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Decryptor(public key): %w", err)
	}

	aead, err := d.prepareAEAD(remote, msg.salt)
	if err != nil {
		return nil, err
	}

	cleartext, err := msg.Open(aead)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Decryptor(open): %w", err)
	}

	return cleartext, nil
}

func (d *decryptor) prepareAEAD(remote *ecdh.PublicKey, salt []byte) (cipher.AEAD, error) {
	shared, err := d.local.ECDH(remote)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Decryptor(ecdh): %w", err)
	}

	symmetric, err := hkdf.Key(sha256.New, shared, salt, "", symmetricKeySize)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Decryptor(hkdf): %w", err)
	}

	block, err := aes.NewCipher(symmetric)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Decryptor(aes): %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Decryptor(gcm): %w", err)
	}

	return aead, nil
}
