package hmacsigner

import (
	"encoding/base64"
	"os"
)

// LoadSecretKey attempts to load secret key from a given string by trying the following:
// base64-decoding the string, reading from a file pointed to by the string,
// or returning the string as a byte slice (as a last resort).
// Empty string is considered valid and corresponds to a  `nil` byte slice.
func LoadSecretKey(source string) []byte {
	if source == "" {
		return nil
	}

	// try base64 decode
	if data, err := base64.StdEncoding.DecodeString(source); err == nil {
		return data
	}

	// try loading from file
	if _, err := os.Stat(source); err == nil {
		if data, err := os.ReadFile(source); err == nil {
			return data
		}
	}

	return []byte(source)
}
