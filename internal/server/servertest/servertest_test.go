package servertest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRandomListenAddress(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(*testing.T, string)
	}{
		{
			name: "localhost",
			assertion: func(t *testing.T, got string) {
				assert.True(t, got != "")
				parts := strings.Split(got, ":")
				assert.Len(t, parts, 2)
				assert.Equal(t, "127.0.0.1", parts[0])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, GetRandomListenAddress(t))
		})
	}
}

func TestMakeRequestDiscardResponse(t *testing.T) {
	type setup struct {
		method      string
		url         string
		body        io.ReadCloser
		responder   httpmock.Responder
		hookRequest func(r *http.Request)
	}
	type args struct {
		c *http.Client
	}
	type want struct {
		numCalls int
	}
	tests := []struct {
		name      string
		setup     setup
		args      args
		want      want
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "simple request with explicit default client",
			setup: setup{
				method: http.MethodGet,
				url:    "http://localhost:91/ping",
				body:   http.NoBody,
				responder: func(r *http.Request) (*http.Response, error) {
					return httpmock.NewStringResponse(http.StatusOK, ""), nil
				},
				hookRequest: func(r *http.Request) {},
			},
			args: args{
				c: http.DefaultClient,
			},
			want: want{numCalls: 1},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "simple request with implicit default client",
			setup: setup{
				method: http.MethodGet,
				url:    "http://localhost:91/ping",
				body:   http.NoBody,
				responder: func(r *http.Request) (*http.Response, error) {
					return httpmock.NewStringResponse(http.StatusOK, ""), nil
				},
				hookRequest: func(r *http.Request) {},
			},
			args: args{
				c: nil,
			},
			want: want{numCalls: 1},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "invalid request",
			setup: setup{
				method: http.MethodGet,
				url:    "http://localhost:91/ping",
				body:   http.NoBody,
				responder: func(r *http.Request) (*http.Response, error) {
					return httpmock.NewStringResponse(http.StatusOK, ""), nil
				},
				hookRequest: func(r *http.Request) {
					r.URL = nil
				},
			},
			args: args{
				c: http.DefaultClient,
			},
			want: want{numCalls: 0},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.c == nil {
				httpmock.Activate(t)
			} else {
				httpmock.ActivateNonDefault(tt.args.c)
			}
			defer httpmock.Reset()

			httpmock.RegisterResponder(tt.setup.method, tt.setup.url, tt.setup.responder)

			req, err := http.NewRequest(tt.setup.method, tt.setup.url, tt.setup.body)
			require.NoError(t, err)

			tt.setup.hookRequest(req)

			tt.assertion(t, MakeRequestDiscardResponse(tt.args.c, req))

			calls := httpmock.GetCallCountInfo()
			key := fmt.Sprintf("%s %s", tt.setup.method, tt.setup.url)
			assert.Contains(t, calls, key)
			assert.Equal(t, tt.want.numCalls, calls[key])
		})
	}
}
