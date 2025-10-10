package httpheaders

import (
	"fmt"
	"net/http"
)

const (
	HeaderKeyContentType       = "Content-Type"
	ContentTypeEmpty           = ContentType("")
	ContentTypeTextPlain       = ContentType("text/plain")
	ContentTypeTextHTML        = ContentType("text/html")
	ContentTypeApplicationJSON = ContentType("application/json")
)

type ContentType string

func (c ContentType) String() string {
	return string(c)
}

func (c ContentType) UTF8() ContentType {
	return ContentType(fmt.Sprintf("%s; charset=utf-8", c))
}

func (c ContentType) Matches(header http.Header) bool {
	return c == ContentType(header.Get(HeaderKeyContentType))
}

func (c ContentType) Apply(header http.Header) {
	if c == ContentTypeEmpty {
		header.Del(HeaderKeyContentType)
		return
	}
	header.Set(HeaderKeyContentType, string(c))
}
