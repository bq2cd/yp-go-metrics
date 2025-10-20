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

func GetContentType(header http.Header) ContentType {
	return ContentType(header.Get(HeaderKeyContentType))
}

func (c ContentType) String() string {
	return string(c)
}

func (c ContentType) UTF8() ContentType {
	return ContentType(fmt.Sprintf("%s; charset=utf-8", c))
}

func (c ContentType) Matches(header http.Header) bool {
	return c == GetContentType(header)
}

func (c ContentType) Apply(header http.Header) ContentType {
	if c == ContentTypeEmpty {
		header.Del(HeaderKeyContentType)
	} else {
		header.Set(HeaderKeyContentType, string(c))
	}
	return c
}
