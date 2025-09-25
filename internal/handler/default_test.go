package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultHandler_ServeHTTP(t *testing.T) {
	type args struct {
		method      string
		url         string
		contentType string
		body        io.Reader
	}
	type want struct {
		code        int
		body        string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		// Not Allowed
		{
			name: "GET %s NOT_ALLOWED",
			args: args{method: http.MethodGet, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusMethodNotAllowed, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "HEAD %s NOT_ALLOWED",
			args: args{method: http.MethodHead, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusMethodNotAllowed, body: "", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "PUT %s NOT_ALLOWED",
			args: args{method: http.MethodPut, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusMethodNotAllowed, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "DELETE %s NOT_ALLOWED",
			args: args{method: http.MethodDelete, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusMethodNotAllowed, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		// Bad Request
		{
			name: "POST %s BAD_REQUEST",
			args: args{method: http.MethodPost, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "POST %s BAD_REQUEST",
			args: args{method: http.MethodPost, url: "/bla", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf(tt.name, tt.args.url), func(t *testing.T) {
			h := &defaultHandler{}
			ts := httptest.NewServer(h)
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, tt.args.body)
			require.NoError(t, err)
			req.Header.Set("content-type", tt.args.contentType)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("content-type"))
			assert.Equal(t, tt.want.body, string(body))
		})
	}
}
