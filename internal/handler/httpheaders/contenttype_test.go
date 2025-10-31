package httpheaders

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentType_Apply(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		name      string
		args      args
		c         ContentType
		want      ContentType
		assertion func(*testing.T, ContentType, http.Header)
	}{
		{
			name: "empty content type not applied",
			args: args{h: buildHTTPHeader().Header()},
			want: ContentTypeEmpty,
			assertion: func(t *testing.T, want ContentType, h http.Header) {
				assert.Empty(t, h.Values(HeaderKeyContentType))
			},
		},
		{
			name: "empty content type clears existing",
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentType, "already/existing").Header()},
			want: ContentTypeEmpty,
			assertion: func(t *testing.T, want ContentType, h http.Header) {
				assert.Empty(t, h.Values(HeaderKeyContentType))
			},
		},
		{
			name: "new content type applied",
			args: args{h: buildHTTPHeader().Header()},
			c:    ContentTypeTextPlain,
			want: ContentTypeTextPlain,
			assertion: func(t *testing.T, want ContentType, h http.Header) {
				assert.Len(t, h.Values(HeaderKeyContentType), 1)
				assert.Equal(t, want, ContentType(h.Get(HeaderKeyContentType)))
			},
		},
		{
			name: "new content type overrides existing",
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentType, "already/existing").Header()},
			c:    ContentTypeApplicationJSON,
			want: ContentTypeApplicationJSON,
			assertion: func(t *testing.T, want ContentType, h http.Header) {
				assert.Len(t, h.Values(HeaderKeyContentType), 1)
				assert.Equal(t, want, ContentType(h.Get(HeaderKeyContentType)))
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

func TestContentType_Matches(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		name string
		c    ContentType
		args args
		want bool
	}{
		{
			name: "no content-type",
			c:    ContentTypeTextPlain,
			args: args{h: buildHTTPHeader().Header()},
			want: false,
		},
		{
			name: "different content-type",
			c:    ContentTypeTextPlain,
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentType, ContentTypeApplicationJSON.String()).Header()},
			want: false,
		},
		{
			name: "same content-type",
			c:    ContentTypeApplicationJSON,
			args: args{h: buildHTTPHeader().Set(HeaderKeyContentType, ContentTypeApplicationJSON.String()).Header()},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.Matches(tt.args.h))
		})
	}
}

func TestContentType_String(t *testing.T) {
	tests := []struct {
		name string
		c    ContentType
		want string
	}{
		{
			name: "empty",
			c:    ContentTypeEmpty,
			want: "",
		},
		{
			name: "text",
			c:    ContentTypeTextPlain,
			want: "text/plain",
		},
		{
			name: "json",
			c:    ContentTypeApplicationJSON,
			want: "application/json",
		},
		{
			name: "custom",
			c:    ContentType("application/x.vnd-custom"),
			want: "application/x.vnd-custom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.String())
		})
	}
}

func TestContentType_UTF8(t *testing.T) {
	tests := []struct {
		name string
		c    ContentType
		want ContentType
	}{
		{
			name: "text/plain",
			c:    ContentTypeTextPlain,
			want: ContentType("text/plain; charset=utf-8"),
		},
		{
			name: "text/html",
			c:    ContentTypeTextHTML,
			want: ContentType("text/html; charset=utf-8"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.UTF8())
		})
	}
}

func TestGetContentType(t *testing.T) {
	type args struct {
		header http.Header
	}
	type want struct {
		got ContentType
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := GetContentType(tt.args.header)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
