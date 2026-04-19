package httpheaders

import (
	"fmt"
	"net/http"
)

// HeaderKeyContentType represents header key name for content type.
const HeaderKeyContentType = "Content-Type"

const (
	// ContentTypeEmpty represents absence of `Content-Type` HTTP header.
	ContentTypeEmpty = ContentType("")
	// ContentTypeTextPlain represents `text/plain` content type.
	ContentTypeTextPlain = ContentType("text/plain")
	// ContentTypeTextHTML represents `text/html` content type.
	ContentTypeTextHTML = ContentType("text/html")
	// ContentTypeApplicationJSON represents `application/json` content type.
	ContentTypeApplicationJSON = ContentType("application/json")
)

// ContentType represents HTTP content type.
type ContentType string

// GetContentType extracts [ContentType] from HTTP headers.
func GetContentType(header http.Header) ContentType {
	return ContentType(header.Get(HeaderKeyContentType))
}

// String converts [ContentType] to a string.
func (c ContentType) String() string {
	return string(c)
}

// UTF8 returns a copy of [ContentType] with added `charset=utf-8` piece.
func (c ContentType) UTF8() ContentType {
	return ContentType(fmt.Sprintf("%s; charset=utf-8", c))
}

// Matches returns `true` if current [ContentType] is contained in HTTP headers.
func (c ContentType) Matches(header http.Header) bool {
	return c == GetContentType(header)
}

// Apply add current [ContentType] to the provided HTTP headers.
// If the current content type is empty, `Content-Type` header is removed,
// otherwise it is set with the value of the current content type.
// Any extra `Content-Type` header values are removed.
func (c ContentType) Apply(header http.Header) ContentType {
	if c == ContentTypeEmpty {
		header.Del(HeaderKeyContentType)
	} else {
		header.Set(HeaderKeyContentType, string(c))
	}
	return c
}
