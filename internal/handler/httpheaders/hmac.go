package httpheaders

import (
	"encoding/hex"
	"net/http"
)

// HeaderKeyHashSHA256 represents the key for a custom HTTP header which contains HMAC signature
// of the request/response data.
const HeaderKeyHashSHA256 = "HashSHA265"

const (
	// HashSHA256Empty represents empty HMAC signature.
	HashSHA256Empty = HashSHA256("")
)

// HashSHA256 represents HMAC signature value as a hex-encoded string.
type HashSHA256 string

// GetHashSHA256 extracts the value of HMAC signature from the given HTTP headers.
func GetHashSHA256(header http.Header) HashSHA256 {
	return HashSHA256(header.Get(HeaderKeyHashSHA256))
}

// GetHashSHA256FromBytes converts provided data bytes into hex-encoded string.
func GetHashSHA256FromBytes(data []byte) HashSHA256 {
	return HashSHA256(hex.EncodeToString(data))
}

// String returns string representation of HMAC signature.
func (h HashSHA256) String() string {
	return string(h)
}

// Bytes decode HMAC signature string from hex representation.
// An error is returned if the string is not a valid hex string.
func (h HashSHA256) Bytes() ([]byte, error) {
	if h == HashSHA256Empty {
		return []byte{}, nil
	}
	return hex.DecodeString(h.String())
}

// Matches returns `true` if current HMAC signature matches the signature in the provided HTTP headers.
func (h HashSHA256) Matches(header http.Header) bool {
	return h == GetHashSHA256(header)
}

// Apply sets current HMAC signature as an HTTP header.
// If the current signature is empty, it will remove the corresponding header,
// otherwise it will override all prior values with the value of the current signature.
func (h HashSHA256) Apply(header http.Header) HashSHA256 {
	if h == HashSHA256Empty {
		header.Del(HeaderKeyHashSHA256)
	} else {
		header.Set(HeaderKeyHashSHA256, h.String())
	}
	return h
}
