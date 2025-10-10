package contenttype

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dummyHTTPRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)
	return req
}

func dummyHTTPResponse(_ *testing.T) *http.Response {
	var buf bytes.Buffer
	return &http.Response{Header: make(http.Header), Body: io.NopCloser(&buf)}
}

func TestContentType_ApplyToRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name      string
		args      args
		c         ContentType
		want      ContentType
		assertion func(*testing.T, ContentType, *http.Request)
	}{
		{
			name: "empty content type not applied",
			args: args{r: dummyHTTPRequest(t)},
			want: contentTypeEmpty,
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Empty(t, r.Header.Values(contentTypeHeaderKey))
			},
		},
		{
			name: "empty content type does not override existing",
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(contentTypeHeaderKey, "already/exists")
				return r
			}()},
			want: ContentType("already/exists"),
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Len(t, r.Header.Values(contentTypeHeaderKey), 1)
				assert.Equal(t, want, ContentType(r.Header.Get(contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type applied",
			args: args{r: dummyHTTPRequest(t)},
			c:    TextPlain,
			want: TextPlain,
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Len(t, r.Header.Values(contentTypeHeaderKey), 1)
				assert.Equal(t, want, ContentType(r.Header.Get(contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type overrides existing",
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(contentTypeHeaderKey, "already/exists")
				return r
			}()},
			c:    ApplicationJSON,
			want: ApplicationJSON,
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Len(t, r.Header.Values(contentTypeHeaderKey), 1)
				assert.Equal(t, want, ContentType(r.Header.Get(contentTypeHeaderKey)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.ApplyToRequest(tt.args.r)
			tt.assertion(t, tt.want, tt.args.r)
		})
	}
}

func TestContentType_ApplyToResponse(t *testing.T) {
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name      string
		args      args
		c         ContentType
		want      ContentType
		assertion func(*testing.T, ContentType, *http.Request)
	}{
		{
			name: "empty content type not applied",
			args: args{w: httptest.NewRecorder()},
			want: contentTypeEmpty,
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Empty(t, r.Header.Values(contentTypeHeaderKey))
			},
		},
		{
			name: "empty content type does not override existing",
			args: args{w: func() http.ResponseWriter {
				w := httptest.NewRecorder()
				w.Header().Set(contentTypeHeaderKey, "already/exists")
				return w
			}()},
			want: ContentType("already/exists"),
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Len(t, r.Header.Values(contentTypeHeaderKey), 1)
				assert.Equal(t, want, ContentType(r.Header.Get(contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type applied",
			args: args{w: httptest.NewRecorder()},
			c:    TextPlain,
			want: TextPlain,
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Len(t, r.Header.Values(contentTypeHeaderKey), 1)
				assert.Equal(t, want, ContentType(r.Header.Get(contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type overrides existing",
			args: args{w: func() http.ResponseWriter {
				w := httptest.NewRecorder()
				w.Header().Set(contentTypeHeaderKey, "already/exists")
				return w
			}()},
			c:    ApplicationJSON,
			want: ApplicationJSON,
			assertion: func(t *testing.T, want ContentType, r *http.Request) {
				assert.Len(t, r.Header.Values(contentTypeHeaderKey), 1)
				assert.Equal(t, want, ContentType(r.Header.Get(contentTypeHeaderKey)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.ApplyToResponse(tt.args.w)
		})
	}
}

func TestContentType_MatchesRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		c    ContentType
		args args
		want bool
	}{
		{
			name: "no content-type",
			c:    TextPlain,
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				return r
			}()},
			want: false,
		},
		{
			name: "different content-type",
			c:    TextPlain,
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(contentTypeHeaderKey, string(ApplicationJSON))
				return r
			}()},
			want: false,
		},
		{
			name: "same content-type",
			c:    ApplicationJSON,
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(contentTypeHeaderKey, string(ApplicationJSON))
				return r
			}()},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.MatchesRequest(tt.args.r))
		})
	}
}

func TestContentType_MatchesResponse(t *testing.T) {
	type args struct {
		r func() *http.Response
	}
	tests := []struct {
		name string
		c    ContentType
		args args
		want bool
	}{
		{
			name: "no content-type",
			c:    TextPlain,
			args: args{r: func() *http.Response {
				r := dummyHTTPResponse(t)
				return r
			}},
			want: false,
		},
		{
			name: "different content-type",
			c:    TextPlain,
			args: args{r: func() *http.Response {
				r := dummyHTTPResponse(t)
				r.Header.Set(contentTypeHeaderKey, string(ApplicationJSON))
				return r
			}},
			want: false,
		},
		{
			name: "same content-type",
			c:    ApplicationJSON,
			args: args{r: func() *http.Response {
				r := dummyHTTPResponse(t)
				r.Header.Set(contentTypeHeaderKey, string(ApplicationJSON))
				return r
			}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.args.r()
			defer func() { _ = r.Body.Close() }()
			assert.Equal(t, tt.want, tt.c.MatchesResponse(r))
		})
	}
}
