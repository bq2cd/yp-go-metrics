package httpheaders

import (
	"encoding/hex"
	"net/http"
)

const (
	HeaderKeyHashSHA256 = "HashSHA265"
	HashSHA256Empty     = HashSHA256("")
)

type HashSHA256 string

func GetHashSHA256(header http.Header) HashSHA256 {
	return HashSHA256(header.Get(HeaderKeyHashSHA256))
}

func GetHashSHA256FromBytes(data []byte) HashSHA256 {
	return HashSHA256(hex.EncodeToString(data))
}

func (h HashSHA256) String() string {
	return string(h)
}

func (h HashSHA256) Bytes() ([]byte, error) {
	if h == HashSHA256Empty {
		return []byte{}, nil
	}
	return hex.DecodeString(h.String())
}

func (h HashSHA256) Matches(header http.Header) bool {
	return h == GetHashSHA256(header)
}

func (h HashSHA256) Apply(header http.Header) HashSHA256 {
	if h == HashSHA256Empty {
		header.Del(HeaderKeyHashSHA256)
	} else {
		header.Set(HeaderKeyHashSHA256, h.String())
	}
	return h
}
