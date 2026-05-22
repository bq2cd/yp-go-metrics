package asymcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

//go:generate go tool mockgen -destination=asymcrypttest/mock_encryptor.go -package asymcrypttest github.com/bq2cd/yp-go-metrics/pkg/asymcrypt Encryptor

// Encryptor defines an entity capable of encrypting provided clear text data.
type Encryptor interface {
	Encrypt(cleartext []byte) ([]byte, error)
}

const (
	symmetricKeySize = 32 // AES-256
)

type encryptor struct {
	remote *ecdh.PublicKey // static, provided externally
}

// NewEncryptor creates an instance of [Encryptor] that implements AES-GCM encryption using
// provided static X25519 public key, ECDH and HKDF to derive a symmetric encryption key.
// For ECDH to work, [Encryptor] generates ephemeral (in-memory only) X25519 key pair on initialization.
// The ephemeral public key is then sent over the wire in clear text, for a [Decryptor] to be able to use ECDH
// to derive the same symmetric encryption key.
func NewEncryptor(remote *ecdh.PublicKey) (*encryptor, error) {
	enc := &encryptor{
		remote: remote,
	}

	return enc, nil
}

// Encrypt takes provided clear text and encrypts it with AES-GCM, deriving symmetric encryption key
// from local (ephemeral) X25519 private key and remote (static, provided externally) X25519 public key
// using ECDH and HKDF.
// Resulting ciphertext is combined with local (ephemeral) X25519 public key, HKDF salt and AES-GCM nonce
// into a single message, ready to be sent over the wire.
func (e *encryptor) Encrypt(cleartext []byte) ([]byte, error) {
	ephemeral, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Encryptor(ephemeral): %w", err)
	}

	msg, err := e.prepareMessage(ephemeral.PublicKey())
	if err != nil {
		return nil, err
	}

	aead, err := e.prepareAEAD(ephemeral, msg.salt)
	if err != nil {
		return nil, err
	}

	msg.Seal(aead, cleartext)

	return msg.ToWire(), nil
}

func (e *encryptor) prepareMessage(ephemeralPub *ecdh.PublicKey) (*Message, error) {
	var err error

	msg := &Message{
		publicKey: ephemeralPub.Bytes(),
	}

	err = msg.InitSalt()
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Encryptor(salt): %w", err)
	}

	err = msg.InitNonce()
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Encryptor(nonce): %w", err)
	}

	return msg, nil
}

func (e *encryptor) prepareAEAD(ephemeral *ecdh.PrivateKey, salt []byte) (cipher.AEAD, error) {
	shared, err := ephemeral.ECDH(e.remote)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Encryptor(ecdh): %w", err)
	}

	symmetric, err := hkdf.Key(sha256.New, shared, salt, "", symmetricKeySize)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Encryptor(hkdf): %w", err)
	}

	block, err := aes.NewCipher(symmetric)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Encryptor(aes): %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("asymcrypt.Encryptor(gcm): %w", err)
	}

	return aead, nil
}
