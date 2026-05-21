// Package asymcrypt defines an asymmetric (one-way) hybrid encryption scheme based on
// X25519 public/private key pair, ECDH and AES-GCM.
//
// Initial requirement is a static X25519 public/private key pair: a sender uses the public key,
// and a receiver uses the private key.
//
// The sender generates ephemeral X25519 key pair, uses the static public key to derive a symmetric encryption key
// using ECDH and HKDF, then encrypts plain text data with AES-GCM, producing a ciphertext.
// The sender then composes a message consisting of the sender's ephemeral public key, HKDF salt, AES-GCM nonce
// and the ciphertext; the resulting message is sent over the wire.
//
// The receiver decomposes received message into parts (the sender's ephemeral public key, HKDF salt, AES-GCM nonce,
// the ciphertext), derives symmetric encryption key using ECDH and HKDF, then decrypts the ciphertext into plain
// text data.
//
// The ephemeral X25519 public key, HKDF salt and AES-GCM nonce are sent over the wire in clear text (unencrypted)
// so that the receiver could derive the same symmetric encryption key.
package asymcrypt
