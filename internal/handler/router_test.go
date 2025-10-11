package handler

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/middleware"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
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
					"/update/":  "POST",
					"/update/*": "POST",
					"/value":    "POST",
					"/value/":   "POST",
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

func Test_newRoute(t *testing.T) {
	type args struct {
		handler  http.Handler
		patterns []string
	}
	type want struct {
		patterns []string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "no patterns",
			args: args{handler: http.NewServeMux(), patterns: []string{}},
			want: want{patterns: []string{}},
		},
		{
			name: "single pattern",
			args: args{handler: http.NewServeMux(), patterns: []string{"/"}},
			want: want{patterns: []string{"/"}},
		},
		{
			name: "multiple patterns",
			args: args{handler: http.NewServeMux(), patterns: []string{"/", "GET /bla", "POST /update"}},
			want: want{patterns: []string{"GET /bla", "/", "POST /update"}},
		},
		{
			name: "duplicate patterns",
			args: args{handler: http.NewServeMux(), patterns: []string{"/", "GET /bla", "POST /update", "GET /bla"}},
			want: want{patterns: []string{"GET /bla", "/", "POST /update"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := newRoute(tt.args.handler, tt.args.patterns...)
			assert.ElementsMatch(t, tt.want.patterns, rr.patterns)
			assert.Equal(t, tt.args.handler, rr.handler)
		})
	}
}

func Test_router_getRoutes(t *testing.T) {
	tests := []struct {
		name string
		want map[string]reflect.Type
	}{
		{
			name: "default",
			want: map[string]reflect.Type{
				"/*":             reflect.ValueOf(new(defaultHandler)).Type(),
				"GET /":          reflect.ValueOf(new(readHandler)).Type(),
				"POST /update":   reflect.ValueOf(new(updateJSONHandler)).Type(),
				"POST /update/":  reflect.ValueOf(new(updateJSONHandler)).Type(),
				"POST /update/*": reflect.ValueOf(new(updateHandler)).Type(),
				"POST /value":    reflect.ValueOf(new(valueJSONHandler)).Type(),
				"POST /value/":   reflect.ValueOf(new(valueJSONHandler)).Type(),
				"GET /value/*":   reflect.ValueOf(new(valueHandler)).Type(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &router{}
			got := make(map[string]reflect.Type)
			for _, rr := range rt.getRoutes() {
				for _, p := range rr.patterns {
					got[p] = reflect.ValueOf(rr.handler).Type()
				}
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_router_ServeHTTP(t *testing.T) {
	type args struct {
		method         string
		url            string
		bodyData       testBodyData
		shouldCompress bool
		metrics        []model.Metric
	}
	type want struct {
		code            int
		body            string
		contentType     httpheaders.ContentType
		contentEncoding httpheaders.ContentEncoding
	}
	type innerTest struct {
		name      string
		args      args
		want      want
		assertion func(*testing.T, want, []byte, http.Header)
	}

	runInnerTest := func(t *testing.T, url string, tt innerTest) {
		storage := storagetest.NewMockStorage(tt.args.metrics...)
		logger := log.NewTestLogger()
		rt := NewRouter(logger, service.NewMetrics(storage), nil)
		ts := httptest.NewServer(rt)
		defer ts.Close()

		req, err := tt.args.bodyData.toRequest(tt.args.method, ts.URL+url, tt.args.shouldCompress)
		require.NoError(t, err)
		tt.want.contentEncoding.MakeAccepted(req.Header)

		resp, err := ts.Client().Do(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = resp.Body.Close()
		require.NoError(t, err)
		if tt.want.contentEncoding == httpheaders.ContentEncodingGzip {
			r := bytes.NewReader(body)
			rgz, err := gzip.NewReader(r)
			require.NoError(t, err)
			body, err = io.ReadAll(rgz)
			require.NoError(t, err)
			err = rgz.Close()
			require.NoError(t, err)
		}

		assert.Equal(t, url, resp.Request.URL.Path)
		assert.Equal(t, tt.want.code, resp.StatusCode)
		tt.assertion(t, tt.want, body, resp.Header)

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
	}

	casesDefault := []innerTest{
		{
			args: args{method: http.MethodGet},
			want: want{code: http.StatusMethodNotAllowed, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodPost},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodPut},
			want: want{code: http.StatusMethodNotAllowed, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodDelete},
			want: want{code: http.StatusMethodNotAllowed, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
	}

	casesRead := []innerTest{
		{
			args: args{
				method: http.MethodGet,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 123),
					model.NewGaugeMetric("id2", -1.23),
					model.NewGaugeMetric("id3", 0.01),
				},
			},
			want: want{
				code:        http.StatusOK,
				body:        "id1 123\nid2 -1.23\nid3 0.01",
				contentType: httpheaders.ContentTypeTextHTML,
			},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				content := strings.TrimRight(string(body), "\n")
				assert.ElementsMatch(t, strings.Split(want.body, "\n"), strings.Split(content, "\n"))
				assert.True(t, want.contentType.Matches(h))
			},
		},
	}

	casesUpdate := []innerTest{
		{
			args: args{method: http.MethodPost, url: "/update/counter"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodPost, url: "/update/counter/id1/123"},
			want: want{code: http.StatusOK, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodPost, url: "/update/counter/id1/123/none"},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
	}

	casesUpdateJSON := []innerTest{
		{
			name: "GET not allowed",
			args: args{method: http.MethodGet},
			want: want{code: http.StatusMethodNotAllowed, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "empty body, missing content-type",
			args: args{method: http.MethodPost},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "bad json",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": 1 }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
			},
			want: want{code: http.StatusUnprocessableEntity, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "invalid content type",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "counter" }`),
					contentType: httpheaders.ContentTypeTextPlain,
				},
			},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "new counter",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "counter", "delta": -35 }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
			},
			want: want{code: http.StatusOK, body: `{ "id": "id1", "type": "counter", "delta": -35 }`, contentType: httpheaders.ContentTypeApplicationJSON},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
			},
		},
		{
			name: "new gauge",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "gauge", "value": -0.325 }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
			},
			want: want{code: http.StatusOK, body: `{ "id": "id1", "type": "gauge", "value": -0.325 }`, contentType: httpheaders.ContentTypeApplicationJSON},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
			},
		},
		{
			name: "new counter, gzip compression",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "counter", "delta": -35 }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
				shouldCompress: true,
			},
			want: want{
				code:            http.StatusOK,
				body:            `{ "id": "id1", "type": "counter", "delta": -35 }`,
				contentType:     httpheaders.ContentTypeApplicationJSON,
				contentEncoding: httpheaders.ContentEncodingGzip,
			},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
				assert.True(t, want.contentEncoding.Matches(h))
			},
		},
		{
			name: "new gauge, gzip compression, response only",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "gauge", "value": -0.325 }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
				shouldCompress: true,
			},
			want: want{
				code:            http.StatusOK,
				body:            `{ "id": "id1", "type": "gauge", "value": -0.325 }`,
				contentType:     httpheaders.ContentTypeApplicationJSON,
				contentEncoding: httpheaders.ContentEncodingGzip,
			},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
				assert.True(t, want.contentEncoding.Matches(h))
			},
		},
	}

	casesValue := []innerTest{
		{
			args: args{method: http.MethodGet, url: "/value/counter"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value/counter/"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value/counter/id1"},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{
				method:  http.MethodGet,
				url:     "/value/counter/id1",
				metrics: []model.Metric{model.NewCounterMetric("id1", 456)},
			},
			want: want{code: http.StatusOK, body: "456"},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{
				method:  http.MethodGet,
				url:     "/value/gauge/id1",
				metrics: []model.Metric{model.NewGaugeMetric("id1", -4.56)},
			},
			want: want{code: http.StatusOK, body: "-4.56"},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			args: args{method: http.MethodGet, url: "/value/counter/id1/123"},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
	}

	casesValueJSON := []innerTest{
		{
			name: "bad json",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": 1 }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
			},
			want: want{code: http.StatusUnprocessableEntity, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "invalid content type",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "counter" }`),
					contentType: httpheaders.ContentTypeTextPlain,
				},
			},
			want: want{code: http.StatusBadRequest, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "missing counter",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "counter" }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
			},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "existing counter",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "counter" }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
				metrics: []model.Metric{
					model.NewCounterMetric("id1", -33),
					model.NewGaugeMetric("id1", -0.33),
				},
			},
			want: want{code: http.StatusOK, body: `{ "id": "id1", "type": "counter", "delta": -33 }`, contentType: httpheaders.ContentTypeApplicationJSON},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
			},
		},
		{
			name: "missing gauge",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "gauge" }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
			},
			want: want{code: http.StatusNotFound, body: ""},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.Equal(t, want.body, strings.TrimRight(string(body), "\n"))
			},
		},
		{
			name: "existing gauge",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "gauge" }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
				metrics: []model.Metric{
					model.NewCounterMetric("id1", -33),
					model.NewGaugeMetric("id1", -0.33),
				},
			},
			want: want{code: http.StatusOK, body: `{ "id": "id1", "type": "gauge", "value": -0.33 }`, contentType: httpheaders.ContentTypeApplicationJSON},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
			},
		},
		{
			name: "existing counter, gzip compression",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "counter" }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
				shouldCompress: true,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", -33),
					model.NewGaugeMetric("id1", -0.33),
				},
			},
			want: want{
				code:            http.StatusOK,
				body:            `{ "id": "id1", "type": "counter", "delta": -33 }`,
				contentType:     httpheaders.ContentTypeApplicationJSON,
				contentEncoding: httpheaders.ContentEncodingGzip,
			},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
				assert.True(t, want.contentEncoding.Matches(h))
			},
		},
		{
			name: "existing gauge, gzip compression, response only",
			args: args{
				method: http.MethodPost,
				bodyData: testBodyData{
					data:        []byte(`{ "id": "id1", "type": "gauge" }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
				metrics: []model.Metric{
					model.NewCounterMetric("id1", -33),
					model.NewGaugeMetric("id1", -0.33),
				},
			},
			want: want{
				code:            http.StatusOK,
				body:            `{ "id": "id1", "type": "gauge", "value": -0.33 }`,
				contentType:     httpheaders.ContentTypeApplicationJSON,
				contentEncoding: httpheaders.ContentEncodingGzip,
			},
			assertion: func(t *testing.T, want want, body []byte, h http.Header) {
				assert.JSONEq(t, want.body, string(body))
				assert.True(t, want.contentType.Matches(h))
				assert.True(t, want.contentEncoding.Matches(h))
			},
		},
	}

	outerTests := []struct {
		name  string
		url   string
		cases []innerTest
	}{
		{
			name:  "default handler",
			url:   "/some-weird-url",
			cases: casesDefault,
		},
		{
			name:  "read handler",
			url:   "/",
			cases: casesRead,
		},
		{
			name:  "update handler plain",
			cases: casesUpdate,
		},
		{
			name:  "update handler json",
			url:   "/update",
			cases: casesUpdateJSON,
		},
		{
			name:  "update handler json slash",
			url:   "/update/",
			cases: casesUpdateJSON,
		},
		{
			name:  "value handler plain",
			cases: casesValue,
		},
		{
			name:  "value handler json",
			url:   "/value",
			cases: casesValueJSON,
		},
		{
			name:  "value handler json slash",
			url:   "/value/",
			cases: casesValueJSON,
		},
	}

	for _, outer := range outerTests {
		t.Run(outer.name, func(t *testing.T) {
			for _, inner := range outer.cases {
				url := inner.args.url
				if url == "" {
					url = outer.url
				}
				require.NotEmpty(t, url)
				t.Run(fmt.Sprintf("%s %s %s", inner.name, inner.args.method, url), func(t *testing.T) {
					runInnerTest(t, url, inner)
				})
			}
		})
	}
}

func Test_router_getMiddlewares(t *testing.T) {
	type fields struct {
		logger  log.Logger
		mux     http.Handler
		metrics service.Metrics
	}
	tests := []struct {
		name   string
		fields fields
		want   []middleware.Middleware
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &router{
				logger:  tt.fields.logger,
				mux:     tt.fields.mux,
				metrics: tt.fields.metrics,
			}
			assert.Equal(t, tt.want, rt.getMiddlewares())
		})
	}
}
