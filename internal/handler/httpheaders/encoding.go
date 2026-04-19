package httpheaders

import (
	"net/http"
	"slices"
	"strings"
)

const (
	// HeaderKeyContentEncoding represents HTTP header key for content encoding (e.g. gzip).
	HeaderKeyContentEncoding = "Content-Encoding"
	// HeaderKeyAcceptEncoding represents HTTP header key for acceptable content encoding from the client side.
	HeaderKeyAcceptEncoding = "Accept-Encoding"
)

const (
	// ContentEncodingEmpty represents the absence of content encoding in HTTP headers.
	ContentEncodingEmpty = ContentEncoding("")
	// ContentEncodingGzip represents `gzip` content encoding.
	ContentEncodingGzip = ContentEncoding("gzip")
	// ContentEncodingDeflate represents `deflate` content encoding.
	ContentEncodingDeflate = ContentEncoding("deflate")
)

// ContentEncoding represents content encoding of an HTTP request/response.
type ContentEncoding string

// AcceptedContentEncodings returns a slice of [ContentEncoding] that are provided in HTTP headers
// of a client's request.
func AcceptedContentEncodings(header http.Header) []ContentEncoding {
	accepted := strings.Split(header.Get(HeaderKeyAcceptEncoding), ",")
	result := make([]ContentEncoding, 0, len(accepted))
	for _, enc := range accepted {
		parts := strings.SplitN(enc, ";", 2)
		ce := ContentEncoding(strings.TrimSpace(parts[0]))
		result = append(result, ce)
	}
	return result
}

// GetContentEncoding returns [ContentEncoding] from provided HTTP headers.
func GetContentEncoding(header http.Header) ContentEncoding {
	return ContentEncoding(header.Get(HeaderKeyContentEncoding))
}

// String converts [ContentEncoding] into a string.
func (c ContentEncoding) String() string {
	return string(c)
}

// Accepted returns `true` if current [ContentEncoding] was provided in HTTP headers
// of a client's request.
// Empty content encoding is always accepted.
func (c ContentEncoding) Accepted(header http.Header) bool {
	if c == ContentEncodingEmpty {
		return true
	}
	return slices.Contains(AcceptedContentEncodings(header), c)
}

// MakeAccepted applies current [ContentEncoding] to the provided HTTP headers.
// It parses existing `Accepted-Encoding` header value, adds current content encoding
// if it is not there, and sets HTTP header to the new value.
// The method will return current [ContentEncoding] object, making it suitable
// for chain calling.
func (c ContentEncoding) MakeAccepted(header http.Header) ContentEncoding {
	accepted := AcceptedContentEncodings(header)
	if !slices.Contains(accepted, c) {
		accepted = append(accepted, c)
	}
	var buf strings.Builder
	for i, ce := range accepted {
		if ce == ContentEncodingEmpty {
			continue
		}
		buf.WriteString(string(ce))
		if i < len(accepted)-1 {
			buf.WriteString(", ")
		}
	}
	header.Set(HeaderKeyAcceptEncoding, buf.String())
	return c
}

// Matches returns `true` if provided HTTP headers have current [ContentEncoding] in `Content-Encoding` header.
func (c ContentEncoding) Matches(header http.Header) bool {
	return c == GetContentEncoding(header)
}

// Apply adds current [ContentEncoding] to the provided HTTP headers.
// It will remove the header entirely if the current content encoding is empty,
// otherwise it will override all header values with the current encoding.
// The method will return current [ContentEncoding] object, making it suitable
// for chain calling.
func (c ContentEncoding) Apply(header http.Header) ContentEncoding {
	if c == ContentEncodingEmpty {
		header.Del(HeaderKeyContentEncoding)
	} else {
		header.Set(HeaderKeyContentEncoding, string(c))
	}
	return c
}
