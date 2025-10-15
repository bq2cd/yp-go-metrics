package httpheaders

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentEncoding_Accepted(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		name string
		c    ContentEncoding
		args args
		want bool
	}{
		{
			name: "no header, empty encoding",
			args: args{h: buildHTTPHeader().Header()},
			c:    ContentEncodingEmpty,
			want: true,
		},
		{
			name: "no header, no gzip",
			args: args{h: buildHTTPHeader().Header()},
			c:    ContentEncodingGzip,
			want: false,
		},
		{
			name: "header gzip, want deflate",
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip").Header()},
			c:    ContentEncodingDeflate,
			want: false,
		},
		{
			name: "header gzip, want gzip",
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip").Header()},
			c:    ContentEncodingGzip,
			want: true,
		},
		{
			name: "header gzip, deflate, br; want gzip",
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip, deflate, br").Header()},
			c:    ContentEncodingGzip,
			want: true,
		},
		{
			name: "header with weighted gzip, deflate, br; want gzip",
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip;q=0.5, deflate;q=1.0, br").Header()},
			c:    ContentEncodingGzip,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.Accepted(tt.args.h))
		})
	}
}

func TestContentEncoding_Matches(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		name string
		c    ContentEncoding
		args args
		want bool
	}{
		{
			name: "no header, want empty",
			args: args{h: buildHTTPHeader().Header()},
			c:    ContentEncodingEmpty,
			want: true,
		},
		{
			name: "no header, want deflate",
			args: args{h: buildHTTPHeader().Header()},
			c:    ContentEncodingDeflate,
			want: false,
		},
		{
			name: "header gzip, want deflate",
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentEncoding, "gzip").Header()},
			c:    ContentEncodingDeflate,
			want: false,
		},
		{
			name: "header deflate, want deflate",
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentEncoding, "deflate").Header()},
			c:    ContentEncodingDeflate,
			want: true,
		},
		{
			name: "header deflate, want empty",
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentEncoding, "deflate").Header()},
			c:    ContentEncodingEmpty,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.Matches(tt.args.h))
		})
	}
}

func TestContentEncoding_String(t *testing.T) {
	tests := []struct {
		name string
		c    ContentEncoding
		want string
	}{
		{
			name: "empty",
			c:    ContentEncodingEmpty,
			want: "",
		},
		{
			name: "gzip",
			c:    ContentEncodingGzip,
			want: "gzip",
		},
		{
			name: "deflate",
			c:    ContentEncodingDeflate,
			want: "deflate",
		},
		{
			name: "custom",
			c:    ContentEncoding("zstd1"),
			want: "zstd1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.String())
		})
	}
}

func TestAcceptedContentEncodings(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		name string
		args args
		want []ContentEncoding
	}{
		{
			name: "no header",
			args: args{h: buildHTTPHeader().Header()},
			want: []ContentEncoding{ContentEncodingEmpty},
		},
		{
			name: "single encoding",
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip").Header()},
			want: []ContentEncoding{ContentEncodingGzip},
		},
		{
			name: "multiple encodings",
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip, deflate, zstd").Header()},
			want: []ContentEncoding{ContentEncodingGzip, ContentEncodingDeflate, ContentEncoding("zstd")},
		},
		{
			name: "multiple encodings with weights",
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip;q=0.5, deflate;q=0.8, zstd;q=1.0").Header()},
			want: []ContentEncoding{ContentEncodingGzip, ContentEncodingDeflate, ContentEncoding("zstd")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AcceptedContentEncodings(tt.args.h))
		})
	}
}

func TestContentEncoding_MakeAccepted(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		name string
		c    ContentEncoding
		args args
		want string
	}{
		{
			name: "no header, none requested",
			c:    ContentEncodingEmpty,
			args: args{h: buildHTTPHeader().Header()},
			want: "",
		},
		{
			name: "no header, gzip requested",
			c:    ContentEncodingGzip,
			args: args{h: buildHTTPHeader().Header()},
			want: "gzip",
		},
		{
			name: "header contains deflate, zstd; gzip requested",
			c:    ContentEncodingGzip,
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "deflate, zstd").Header()},
			want: "deflate, zstd, gzip",
		},
		{
			name: "header contains deflate, zstd with weigths; gzip requested",
			c:    ContentEncodingGzip,
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "deflate;q=0.5, zstd;q=0.3").Header()},
			want: "deflate, zstd, gzip",
		},
		{
			name: "header contains gzip, zstd; gzip requested",
			c:    ContentEncodingGzip,
			args: args{h: buildHTTPHeader().Set(HeaderKeyAcceptEncoding, "gzip, zstd").Header()},
			want: "gzip, zstd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.MakeAccepted(tt.args.h)
			assert.Equal(t, tt.want, tt.args.h.Get(HeaderKeyAcceptEncoding))
		})
	}
}

func TestContentEncoding_Apply(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		name      string
		c         ContentEncoding
		args      args
		want      ContentEncoding
		assertion func(*testing.T, ContentEncoding, http.Header)
	}{
		{
			name: "empty encoding not applied",
			args: args{h: buildHTTPHeader().Header()},
			want: ContentEncodingEmpty,
			assertion: func(t *testing.T, want ContentEncoding, h http.Header) {
				assert.Empty(t, h.Values(HeaderKeyContentEncoding))
			},
		},
		{
			name: "empty encoding clears existing",
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentEncoding, "deflate").Header()},
			want: ContentEncodingEmpty,
			assertion: func(t *testing.T, want ContentEncoding, h http.Header) {
				assert.Empty(t, h.Values(HeaderKeyContentEncoding))
			},
		},
		{
			name: "new encoding applied",
			args: args{h: buildHTTPHeader().Header()},
			c:    ContentEncodingGzip,
			want: ContentEncodingGzip,
			assertion: func(t *testing.T, want ContentEncoding, h http.Header) {
				assert.Len(t, h.Values(HeaderKeyContentEncoding), 1)
				assert.Equal(t, want, ContentEncoding(h.Get(HeaderKeyContentEncoding)))
			},
		},
		{
			name: "new encoding overrides existing",
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentEncoding, "deflate").Header()},
			c:    ContentEncodingGzip,
			want: ContentEncodingGzip,
			assertion: func(t *testing.T, want ContentEncoding, h http.Header) {
				assert.Len(t, h.Values(HeaderKeyContentEncoding), 1)
				assert.Equal(t, want, ContentEncoding(h.Get(HeaderKeyContentEncoding)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Apply(tt.args.h)
			tt.assertion(t, tt.want, tt.args.h)
		})
	}
}
