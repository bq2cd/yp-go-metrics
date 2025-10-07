package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	type args struct {
		logger  log.Logger
		metrics service.Metrics
		mux     http.Handler
	}
	type want struct {
		routes map[string]string
		walkFn func(map[string]string, map[string]string) chi.WalkFunc
	}
	tests := []struct {
		name      string
		args      args
		want      want
		assertion func(*router, want) bool
	}{
		{
			name: "new router with mux=nil",
			args: args{logger: log.NewNoopLogger(), metrics: service.NewMetrics(repository.NewMemStorage()), mux: nil},
			want: want{
				routes: map[string]string{
					"/":         "GET",
					"/update":   "POST",
					"/update/*": "POST",
					"/value":    "POST",
					"/value/*":  "GET",
				},
				walkFn: func(want map[string]string, seen map[string]string) chi.WalkFunc {
					walk := func(method string, route string, hh http.Handler, middlewares ...func(http.Handler) http.Handler) error {
						if v, ok := want[route]; ok {
							if v == method {
								seen[route] = method
							}
						}
						return nil
					}
					return walk
				},
			},
			assertion: func(rt *router, want want) bool {
				assert.NotNil(t, rt.logger)
				assert.NotNil(t, rt.mux)
				assert.Implements(t, (*chi.Router)(nil), rt.mux)
				assert.NotNil(t, rt.metrics)
				assert.Implements(t, (*service.Metrics)(nil), rt.metrics)
				seen := make(map[string]string)
				err := chi.Walk(rt.mux.(chi.Router), want.walkFn(want.routes, seen))
				assert.NoError(t, err)
				assert.Equal(t, want.routes, seen)
				return true
			},
		},
		{
			name: "new router with mux=NewServeMux()",
			args: args{logger: log.NewNoopLogger(), metrics: service.NewMetrics(repository.NewMemStorage()), mux: http.NewServeMux()},
			assertion: func(rt *router, want want) bool {
				assert.NotNil(t, rt.logger)
				assert.NotNil(t, rt.mux)
				assert.Implements(t, (*http.Handler)(nil), rt.mux)
				assert.NotNil(t, rt.metrics)
				assert.Implements(t, (*service.Metrics)(nil), rt.metrics)
				return true
			},
		},
		{
			name: "new router with mux=chi.NewRouter()",
			args: args{logger: log.NewNoopLogger(), metrics: service.NewMetrics(repository.NewMemStorage()), mux: chi.NewRouter()},
			assertion: func(rt *router, want want) bool {
				assert.NotNil(t, rt.logger)
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
			assert.True(t, tt.assertion(NewRouter(tt.args.logger, tt.args.metrics, tt.args.mux), tt.want))
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
		method   string
		url      string
		bodyData testBodyData
		metrics  []model.Metric
	}
	type want struct {
		code int
		body string
	}
	tests := []struct {
		args      args
		want      want
		assertion func(assert.TestingT, want, string)
	}{
		{
			args: args{
				method: http.MethodGet,
				url:    "/",
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 123),
					model.NewGaugeMetric("id2", -1.23),
					model.NewGaugeMetric("id3", 0.01),
				},
			},
			want: want{code: http.StatusOK, body: "id1 123\nid2 -1.23\nid3 0.01"},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.ElementsMatch(t, strings.Split(want.body, "\n"), strings.Split(body, "\n"))
			},
		},
		{
			args: args{method: http.MethodGet, url: "/bla"},
			want: want{code: http.StatusMethodNotAllowed, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodGet, url: "/update"},
			want: want{code: http.StatusMethodNotAllowed, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value"},
			want: want{code: http.StatusMethodNotAllowed, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodPost, url: "/update"},
			want: want{code: http.StatusUnprocessableEntity, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodPost, url: "/update/"},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodPost, url: "/update/counter"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodPost, url: "/update/counter/id1/123"},
			want: want{code: http.StatusOK, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodPost, url: "/update/counter/id1/123/none"},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value/counter"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value/counter/"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value/counter/id1"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{
				method:  http.MethodGet,
				url:     "/value/counter/id1",
				metrics: []model.Metric{model.NewCounterMetric("id1", 456)},
			},
			want: want{code: http.StatusOK, body: "456"},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{
				method:  http.MethodGet,
				url:     "/value/gauge/id1",
				metrics: []model.Metric{model.NewGaugeMetric("id1", -4.56)},
			},
			want: want{code: http.StatusOK, body: "-4.56"},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value/counter/id1/123"},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s %d", tt.args.method, tt.args.url, tt.want.code), func(t *testing.T) {
			stor := repository.NewMemStorage()
			for _, m := range tt.args.metrics {
				require.NoError(t, stor.Set(m))
			}
			logger := log.NewTestLogger()
			rt := NewRouter(logger, service.NewMetrics(stor), nil)
			ts := httptest.NewServer(rt)
			defer ts.Close()

			req, err := tt.args.bodyData.toRequest(tt.args.method, ts.URL+tt.args.url)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.args.url, resp.Request.URL.Path)
			assert.Equal(t, tt.want.code, resp.StatusCode)
			tt.assertion(t, tt.want, strings.TrimRight(string(body), "\n"))

			events := logger.RecordedEvents()
			assert.GreaterOrEqual(t, len(events), 1)
			found := events.FindMatchingEvents(
				log.LevelInfo, "processed request",
				log.Str("uri", resp.Request.URL.Path),
				log.Str("method", tt.args.method),
				log.Int("status", resp.StatusCode),
			)
			require.Len(t, found, 1)
			e := found[0]
			fsize := e.Fields().GetFieldByKey("size")
			require.NotNil(t, fsize)
			assert.Equal(t, len(body), fsize.Value)
			assert.NotNil(t, e.Fields().GetFieldByKey("duration"))
			assert.NotNil(t, e.Fields().GetFieldByKey("request_id"))
		})
	}

}

func Test_router_configureChiRouter(t *testing.T) {
	type fields struct {
		logger  log.Logger
		mux     http.Handler
		metrics service.Metrics
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// other cases are covered by `TestNewRouter`
		{
			name: "ignore non-chi mux",
			fields: fields{
				logger:  log.NewNoopLogger(),
				mux:     http.NewServeMux(),
				metrics: service.NewMetrics(repository.NewMemStorage()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &router{
				logger:  tt.fields.logger,
				mux:     tt.fields.mux,
				metrics: tt.fields.metrics,
			}
			rt.configureChiRouter()
		})
	}
}
