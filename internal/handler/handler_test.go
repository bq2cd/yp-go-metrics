package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBodyData struct {
	data        []byte
	contentType contentType
}

func (b *testBodyData) toRequest(method, url string) (*http.Request, error) {
	var body io.ReadCloser = http.NoBody
	if b.data != nil {
		body = io.NopCloser(bytes.NewReader(b.data))
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	b.contentType.applyToRequest(req)
	return req, nil
}

func dummyHTTPRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)
	return req
}

func Test_contentType_applyToRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name      string
		args      args
		c         contentType
		want      contentType
		assertion func(*testing.T, contentType, *http.Request)
	}{
		{
			name: "empty content type not applied",
			args: args{r: dummyHTTPRequest(t)},
			want: _contentTypeEmpty,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Empty(t, r.Header.Values(_contentTypeHeaderKey))
			},
		},
		{
			name: "empty content type does not override existing",
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(_contentTypeHeaderKey, "already/exists")
				return r
			}()},
			want: contentType("already/exists"),
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type applied",
			args: args{r: dummyHTTPRequest(t)},
			c:    contentTypeTextPlain,
			want: contentTypeTextPlain,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type overrides existing",
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(_contentTypeHeaderKey, "already/exists")
				return r
			}()},
			c:    contentTypeApplicationJSON,
			want: contentTypeApplicationJSON,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.applyToRequest(tt.args.r)
			tt.assertion(t, tt.want, tt.args.r)
		})
	}
}

func Test_contentType_applyToResponse(t *testing.T) {
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name      string
		args      args
		c         contentType
		want      contentType
		assertion func(*testing.T, contentType, *http.Request)
	}{
		{
			name: "empty content type not applied",
			args: args{w: httptest.NewRecorder()},
			want: _contentTypeEmpty,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Empty(t, r.Header.Values(_contentTypeHeaderKey))
			},
		},
		{
			name: "empty content type does not override existing",
			args: args{w: func() http.ResponseWriter {
				w := httptest.NewRecorder()
				w.Header().Set(_contentTypeHeaderKey, "already/exists")
				return w
			}()},
			want: contentType("already/exists"),
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type applied",
			args: args{w: httptest.NewRecorder()},
			c:    contentTypeTextPlain,
			want: contentTypeTextPlain,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type overrides existing",
			args: args{w: func() http.ResponseWriter {
				w := httptest.NewRecorder()
				w.Header().Set(_contentTypeHeaderKey, "already/exists")
				return w
			}()},
			c:    contentTypeApplicationJSON,
			want: contentTypeApplicationJSON,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.applyToResponse(tt.args.w)
		})
	}
}
