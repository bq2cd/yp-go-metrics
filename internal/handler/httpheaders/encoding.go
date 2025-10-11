package httpheaders

import (
	"net/http"
	"slices"
	"strings"
)

const (
	HeaderKeyContentEncoding = "Content-Encoding"
	HeaderKeyAcceptEncoding  = "Accept-Encoding"
	ContentEncodingEmpty     = ContentEncoding("")
	ContentEncodingGzip      = ContentEncoding("gzip")
	ContentEncodingDeflate   = ContentEncoding("deflate")
)

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

type ContentEncoding string

func (c ContentEncoding) String() string {
	return string(c)
}

func (c ContentEncoding) Accepted(header http.Header) bool {
	if c == ContentEncodingEmpty {
		return true
	}
	return slices.Contains(AcceptedContentEncodings(header), c)
}

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

func (c ContentEncoding) Matches(header http.Header) bool {
	return c == ContentEncoding(header.Get(HeaderKeyContentEncoding))
}

func (c ContentEncoding) Apply(header http.Header) ContentEncoding {
	if c == ContentEncodingEmpty {
		header.Del(HeaderKeyContentEncoding)
	} else {
		header.Set(HeaderKeyContentEncoding, string(c))
	}
	return c
}
