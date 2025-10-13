package servertest

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func filterNonEmptyStrings(in []string) []string {
	out := make([]string, 0, cap(in))
	for _, e := range in {
		if e != "" {
			out = append(out, e)
		}
	}
	return out
}

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

func TestNewTempFileFactory(t *testing.T) {
	assert.Equal(t, &tempFileFactory{
		t:       t,
		created: make([]string, 0),
	}, NewTempFileFactory(t))
}

func Test_tempFileFactory_create(t *testing.T) {
	type args struct {
		dir     string
		pattern string
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, string)
	}{
		{
			name: "empty dir, empty pattern",
			args: args{},
			assertion: func(t *testing.T, path string) {
				assert.FileExists(t, path)
			},
		},
		{
			name: "existing dir, empty pattern",
			args: args{dir: "/tmp"},
			assertion: func(t *testing.T, path string) {
				assert.FileExists(t, path)
			},
		},
		{
			name: "empty dir, some pattern",
			args: args{pattern: "temp-factory-*-create"},
			assertion: func(t *testing.T, path string) {
				assert.FileExists(t, path)
			},
		},
		{
			name: "existing dir, some pattern",
			args: args{dir: "/tmp", pattern: "temp-factory-*-create"},
			assertion: func(t *testing.T, path string) {
				assert.FileExists(t, path)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := &tempFileFactory{
				t:       t,
				created: make([]string, 0),
			}
			got := ff.create(tt.args.dir, tt.args.pattern)
			tt.assertion(t, got)
			_ = os.Remove(got)
		})
	}
}

func Test_tempFileFactory_Create(t *testing.T) {
	type fields struct {
		created []string
	}
	type args struct {
		pattern string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion func(*testing.T, []string)
	}{
		{
			name:   "nothing created yet",
			fields: fields{created: []string{}},
			args:   args{pattern: "some-*-thing"},
			assertion: func(t *testing.T, created []string) {
				require.Len(t, created, 1)
			},
		},
		{
			name: "some previously created",
			fields: fields{created: func() []string {
				out := make([]string, 2)
				out[0] = "/tmp/123"
				out[1] = "/tmp/456"
				return out
			}()},
			args: args{pattern: "some-*-thing"},
			assertion: func(t *testing.T, created []string) {
				require.Len(t, created, 3)
				assert.Equal(t, created[0], "/tmp/123")
				assert.Equal(t, created[1], "/tmp/456")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := &tempFileFactory{
				t:       t,
				created: tt.fields.created,
			}
			defer ff.RemoveAll()
			got := ff.Create(tt.args.pattern)
			assert.FileExists(t, got)
			tt.assertion(t, ff.created)
		})
	}
}

func Test_tempFileFactory_RemoveAll(t *testing.T) {
	type fields struct {
		created []string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "nothing created yet",
			fields: fields{created: []string{}},
		},
		{
			name: "some non-existent files",
			fields: fields{created: func() []string {
				out := make([]string, 2)
				out[0] = "/tmp/123"
				out[1] = "/tmp/456"
				return out
			}()},
		},
		{
			name: "some existing files",
			fields: fields{created: func() []string {
				f := NewTempFileFactory(t)
				out := make([]string, 2)
				out[0] = f.create("", "dummy-*")
				out[1] = f.create("", "dummy-*")
				return out
			}()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := &tempFileFactory{
				t:       t,
				created: tt.fields.created,
			}
			ff.RemoveAll()
			for _, f := range ff.created {
				assert.NoFileExists(t, f)
			}
		})
	}
}

func TestNewListenAddressFactory(t *testing.T) {
	assert.Equal(t, &listenAddressFactory{
		t:         t,
		allocated: make([]string, 0),
	}, NewListenAddressFactory(t))
}

func Test_listenAddressFactory_New(t *testing.T) {
	type fields struct {
		allocated []string
	}
	tests := []struct {
		name      string
		fields    fields
		assertion func(*testing.T, []string)
	}{
		{
			name:   "allocated empty",
			fields: fields{allocated: []string{}},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 1)
				require.Len(t, allocated, 1)
				assert.NotEmpty(t, allocated[0])
			},
		},
		{
			name: "allocated non-empty, not sparse",
			fields: fields{allocated: func() []string {
				n := 5
				out := make([]string, n)
				for i := range n {
					out[i] = fmt.Sprintf("addr%d", i)
				}
				return out
			}()},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 6)
				require.Len(t, allocated, 6)
				assert.NotEmpty(t, allocated[5])
				for i := range 5 {
					assert.NotEmpty(t, allocated[i])
				}
			},
		},
		{
			name: "allocated non-empty, sparse",
			fields: fields{allocated: func() []string {
				n := 5
				out := make([]string, n)
				out[1] = "addr1"
				out[3] = "addr3"
				return out
			}()},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 3)
				require.Len(t, allocated, 6)
				assert.NotEmpty(t, allocated[5])
				assert.NotEmpty(t, allocated[1])
				assert.NotEmpty(t, allocated[3])
				assert.Empty(t, allocated[0])
				assert.Empty(t, allocated[2])
				assert.Empty(t, allocated[4])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &listenAddressFactory{
				t:         t,
				allocated: tt.fields.allocated,
			}
			got := f.New()
			assert.NotEmpty(t, got)
			tt.assertion(t, f.allocated)
		})
	}
}

func Test_listenAddressFactory_Get(t *testing.T) {
	type fields struct {
		allocated []string
	}
	type args struct {
		idx int
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion func(*testing.T, []string)
	}{
		{
			name:   "allocated empty, idx 0",
			fields: fields{allocated: []string{}},
			args:   args{idx: 0},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 1)
				require.Len(t, allocated, 1)
				assert.NotEmpty(t, allocated[0])
			},
		},
		{
			name:   "allocated empty, idx 5",
			fields: fields{allocated: []string{}},
			args:   args{idx: 5},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 1)
				require.Len(t, allocated, 6)
				assert.NotEmpty(t, allocated[5])
			},
		},
		{
			name: "allocated in the middle, idx 5",
			fields: fields{allocated: func() []string {
				out := make([]string, 10)
				out[3] = "something"
				return out
			}()},
			args: args{idx: 5},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 2)
				require.Len(t, allocated, 10)
				assert.NotEmpty(t, allocated[5])
				assert.NotEmpty(t, allocated[3])
			},
		},
		{
			name: "allocated in the end, idx 15",
			fields: fields{allocated: func() []string {
				out := make([]string, 10)
				out[8] = "addr"
				out[9] = "something"
				return out
			}()},
			args: args{idx: 15},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 3)
				require.Len(t, allocated, 16)
				assert.NotEmpty(t, allocated[15])
				assert.NotEmpty(t, allocated[8])
				assert.NotEmpty(t, allocated[9])
			},
		},
		{
			name: "allocated in the end, idx 0",
			fields: fields{allocated: func() []string {
				out := make([]string, 10)
				out[8] = "addr"
				out[9] = "something"
				return out
			}()},
			args: args{idx: 0},
			assertion: func(t *testing.T, allocated []string) {
				require.Len(t, filterNonEmptyStrings(allocated), 3)
				require.Len(t, allocated, 10)
				assert.NotEmpty(t, allocated[0])
				assert.NotEmpty(t, allocated[8])
				assert.NotEmpty(t, allocated[9])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &listenAddressFactory{
				t:         t,
				allocated: tt.fields.allocated,
			}
			got := f.Get(tt.args.idx)
			assert.NotEmpty(t, got)
			tt.assertion(t, f.allocated)
		})
	}
}

func TestGetCwd(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, cwd, GetCwd(t))
}
