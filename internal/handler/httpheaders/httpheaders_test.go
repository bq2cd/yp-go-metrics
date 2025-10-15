package httpheaders

import (
	"net/http"
)

type headerBuilder struct {
	header http.Header
}

func (h *headerBuilder) Set(key, value string) *headerBuilder {
	h.header.Set(key, value)
	return h
}

func (h *headerBuilder) Header() http.Header {
	return h.header
}

func buildHTTPHeader() *headerBuilder {
	return &headerBuilder{header: make(http.Header)}
}
