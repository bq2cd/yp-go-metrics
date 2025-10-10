package contenttype

import (
	"net/http"
)

const (
	HeaderKey       = "Content-Type"
	Empty           = ContentType("")
	TextPlain       = ContentType("text/plain")
	TextPlainUTF8   = ContentType("text/plain; charset=utf-8")
	ApplicationJSON = ContentType("application/json")
)

type ContentType string

func (c ContentType) ApplyToRequest(r *http.Request) {
	if c == Empty {
		return
	}
	r.Header.Set(HeaderKey, string(c))
}

func (c ContentType) ApplyToResponse(w http.ResponseWriter) {
	if c == Empty {
		return
	}
	w.Header().Set(HeaderKey, string(c))
}

func (c ContentType) MatchesRequest(r *http.Request) bool {
	target := ContentType(r.Header.Get(HeaderKey))
	return target == c
}

func (c ContentType) MatchesResponse(r *http.Response) bool {
	target := ContentType(r.Header.Get(HeaderKey))
	return target == c
}

func (c ContentType) String() string {
	return string(c)
}
