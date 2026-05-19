package servertest

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateHMACKeyBase64 generates random 32-byte key and encodes it with Base64 standard encoding.
// It will panic/crash the program if [crypto/rand.Read] encounters an error.
func GenerateHMACKeyBase64() string {
	buf := [32]byte{}

	// As per documentation to [rand.Read], it never returns an error but rather crashes the program irrecoverably.
	rand.Read(buf[:])

	return base64.StdEncoding.EncodeToString(buf[:])
}
