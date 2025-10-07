package handler

import (
	"net/http"
)

const (
	_contentTypeHeaderKey      = "Content-Type"
	_contentTypeEmpty          = contentType("")
	contentTypeTextPlain       = contentType("text/plain")
	contentTypeApplicationJSON = contentType("application/json")
)

type contentType string

func (c contentType) applyToRequest(r *http.Request) {
	if c == _contentTypeEmpty {
		return
	}
	r.Header.Set(_contentTypeHeaderKey, string(c))
}

func (c contentType) applyToResponse(w http.ResponseWriter) {
	if c == _contentTypeEmpty {
		return
	}
	w.Header().Set(_contentTypeHeaderKey, string(c))
}
