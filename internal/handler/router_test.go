package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	type args struct {
		metrics service.Metrics
		mux     *http.ServeMux
	}
	tests := []struct {
		name string
		args args
		want func(*router) bool
	}{
		{
			name: "new router with mux=nil",
			args: args{metrics: service.NewMetrics(repository.NewMemStorage()), mux: nil},
			want: func(rt *router) bool {
				assert.NotNil(t, rt.mux)
				assert.Implements(t, (*http.Handler)(nil), rt.mux)
				assert.NotNil(t, rt.metrics)
				assert.Implements(t, (*service.Metrics)(nil), rt.metrics)
				return true
			},
		},
		{
			name: "new router with mux=NewServeMux()",
			args: args{metrics: service.NewMetrics(repository.NewMemStorage()), mux: http.NewServeMux()},
			want: func(rt *router) bool {
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
			assert.True(t, tt.want(NewRouter(tt.args.metrics, tt.args.mux)))
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
			mux := new(MockServeMux)

			mux.On("ServeHTTP", mock.Anything, mock.Anything).Return()

			rt := &router{
				mux:     mux,
				metrics: service.NewMetrics(repository.NewMemStorage()),
			}
			ts := httptest.NewServer(rt)
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, http.NoBody)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			mux.AssertExpectations(t)
			mux.AssertNumberOfCalls(t, "ServeHTTP", 1)

			assert.Equal(t, tt.args.url, mux.urlPath)
		})
	}

}
