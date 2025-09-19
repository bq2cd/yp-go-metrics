package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	type args struct {
		metrics service.Metrics
		mux     http.Handler
	}
	type want struct {
		walkFn func(map[string]any) chi.WalkFunc
	}
	tests := []struct {
		name      string
		args      args
		want      want
		assertion func(*router, want) bool
	}{
		{
			name: "new router with mux=nil",
			args: args{metrics: service.NewMetrics(repository.NewMemStorage()), mux: nil},
			want: want{
				walkFn: func(seen map[string]any) chi.WalkFunc {
					walk := func(method string, route string, hh http.Handler, middlewares ...func(http.Handler) http.Handler) error {
						seen[fmt.Sprintf("%s %s", method, route)] = true
						switch route {
						case "/":
							h, ok := hh.(*defaultHandler)
							if !ok {
								return fmt.Errorf("invalid handler for /: %+v, %T", h, hh)
							}
							return nil
						case "/update/*":
							if method != http.MethodPost {
								return fmt.Errorf("invalid method for /update")
							}
							h, ok := hh.(*updateHandler)
							if !ok {
								return fmt.Errorf("invalid handler for POST /update: %+v %T", h, hh)
							}
							return nil
						}
						return fmt.Errorf("unknown route: %v %v", method, route)
					}
					return walk
				},
			},
			assertion: func(rt *router, want want) bool {
				assert.NotNil(t, rt.mux)
				assert.Implements(t, (*chi.Router)(nil), rt.mux)
				assert.NotNil(t, rt.metrics)
				assert.Implements(t, (*service.Metrics)(nil), rt.metrics)
				seen := make(map[string]any)
				err := chi.Walk(rt.mux.(chi.Router), want.walkFn(seen))
				assert.NoError(t, err)
				assert.Contains(t, seen, "POST /update/*")
				return true
			},
		},
		{
			name: "new router with mux=NewServeMux()",
			args: args{metrics: service.NewMetrics(repository.NewMemStorage()), mux: http.NewServeMux()},
			assertion: func(rt *router, want want) bool {
				assert.NotNil(t, rt.mux)
				assert.Implements(t, (*http.Handler)(nil), rt.mux)
				assert.NotNil(t, rt.metrics)
				assert.Implements(t, (*service.Metrics)(nil), rt.metrics)
				return true
			},
		},
		{
			name: "new router with mux=chi.NewRouter()",
			args: args{metrics: service.NewMetrics(repository.NewMemStorage()), mux: chi.NewRouter()},
			assertion: func(rt *router, want want) bool {
				assert.NotNil(t, rt.mux)
				assert.Implements(t, (*http.Handler)(nil), rt.mux)
				assert.NotNil(t, rt.metrics)
				assert.Implements(t, (*service.Metrics)(nil), rt.metrics)
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.assertion(NewRouter(tt.args.metrics, tt.args.mux), tt.want))
		})
	}
}

type MockServeMux struct {
	mock.Mock
	urlPath string
}

func (m *MockServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	m.urlPath = r.URL.Path
}

func Test_router_ServeHTTP(t *testing.T) {
	type args struct {
		method string
		url    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "%s %s",
			args: args{method: http.MethodGet, url: "/"},
		},
		{
			name: "%s %s",
			args: args{method: http.MethodGet, url: "/bla"},
		},
		{
			name: "%s %s",
			args: args{method: http.MethodGet, url: "/update"},
		},
		{
			name: "%s %s",
			args: args{method: http.MethodPost, url: "/update/"},
		},
		{
			name: "%s %s",
			args: args{method: http.MethodPost, url: "/update/counter"},
		},
		{
			name: "%s %s",
			args: args{method: http.MethodPost, url: "/update/counter/id1/123"},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf(tt.name, tt.args.method, tt.args.url), func(t *testing.T) {
			rt := NewRouter(service.NewMetrics(repository.NewMemStorage()), nil)
			ts := httptest.NewServer(rt)
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, http.NoBody)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.args.url, resp.Request.URL.Path)
		})
	}

}
