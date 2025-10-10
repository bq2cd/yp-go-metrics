package contenttype

import (
	"net/http"
)

const (
	contentTypeHeaderKey = "Content-Type"
	contentTypeEmpty     = ContentType("")

	TextPlain       = ContentType("text/plain")
	TextPlainUTF8   = ContentType("text/plain; charset=utf-8")
	ApplicationJSON = ContentType("application/json")
)

type ContentType string

func (c ContentType) ApplyToRequest(r *http.Request) {
	if c == contentTypeEmpty {
		return
	}
	r.Header.Set(contentTypeHeaderKey, string(c))
}

func (c ContentType) ApplyToResponse(w http.ResponseWriter) {
	if c == contentTypeEmpty {
		return
	}
	w.Header().Set(contentTypeHeaderKey, string(c))
}

func (c ContentType) MatchesRequest(r *http.Request) bool {
	target := ContentType(r.Header.Get(contentTypeHeaderKey))
	return target == c
}

func (c ContentType) MatchesResponse(r *http.Response) bool {
	target := ContentType(r.Header.Get(contentTypeHeaderKey))
	return target == c
}
