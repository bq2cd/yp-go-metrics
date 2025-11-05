package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/handler/handlertest"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/middleware"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service/servicetest"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRouter_ServeHTTP(t *testing.T) {
	t.Run("handlers", testRouterServeHTTPHandlers)
	t.Run("middleware", testRouterServeHTTPMiddleware)
}

type testHandlerArgs struct {
	method              string
	url                 string
	bodyData            handlertest.BodyData
	shouldCompress      bool
	acceptedEncodings   []httpheaders.ContentEncoding
	shouldSign          bool
	expectNoHandlerCall bool
	overrideHash        httpheaders.HashSHA256
}

type testHandlerWant struct {
	status          int
	contentType     httpheaders.ContentType
	contentEncoding httpheaders.ContentEncoding
	body            []byte
	hash            httpheaders.HashSHA256
}

type testHandlerCase struct {
	args      testHandlerArgs
	want      testHandlerWant
	secretKey []byte
}

func testRouterServeHTTPHandlers(t *testing.T) {
	type testcase struct {
		method string
		url    string
	}
	tests := map[handler.Ident]struct {
		status      int
		body        []byte
		contentType httpheaders.ContentType
		cases       []testcase
	}{
		handler.IdentDefault: {
			status:      http.StatusOK,
			body:        []byte(`default`),
			contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			cases: []testcase{
				{method: http.MethodGet, url: "/some-url"},
				{method: http.MethodGet, url: "/123/456"},
				{method: http.MethodGet, url: "/value"},
				{method: http.MethodPost, url: "/value/counter"},
				{method: http.MethodPost, url: "/value/counter/id1"},
				{method: http.MethodGet, url: "/update"},
				{method: http.MethodGet, url: "/update/"},
				{method: http.MethodGet, url: "/update/counter/id1"},
				{method: http.MethodGet, url: "/update/counter/id1/123"},
			},
		},
		handler.IdentRead: {
			status:      http.StatusOK,
			body:        []byte(`<html>metrics</html>`),
			contentType: httpheaders.ContentTypeTextHTML,
			cases: []testcase{
				{method: http.MethodGet, url: "/"},
			},
		},
		handler.IdentUpdate: {
			status:      http.StatusOK,
			body:        []byte{},
			contentType: httpheaders.ContentTypeTextPlain,
			cases: []testcase{
				{method: http.MethodPost, url: "/update/counter"},
				{method: http.MethodPost, url: "/update/counter/"},
				{method: http.MethodPost, url: "/update/counter/id1"},
				{method: http.MethodPost, url: "/update/counter/id1/"},
				{method: http.MethodPost, url: "/update/counter/id1/123"},
				{method: http.MethodPost, url: "/update/counter/id1/123/"},
				{method: http.MethodPost, url: "/update/counter/id1/123/something"},
				{method: http.MethodPost, url: "/update/gauge/id1"},
				{method: http.MethodPost, url: "/update/gauge/id1/123"},
			},
		},
		handler.IdentUpdateJSON: {
			status:      http.StatusOK,
			body:        []byte(`{}`),
			contentType: httpheaders.ContentTypeApplicationJSON,
			cases: []testcase{
				{method: http.MethodPost, url: "/update"},
				{method: http.MethodPost, url: "/update/"},
			},
		},
		handler.IdentUpdateBatchJSON: {
			status:      http.StatusOK,
			body:        []byte(`{}`),
			contentType: httpheaders.ContentTypeApplicationJSON,
			cases: []testcase{
				{method: http.MethodPost, url: "/updates"},
				{method: http.MethodPost, url: "/updates/"},
			},
		},
		handler.IdentValue: {
			status:      http.StatusOK,
			body:        []byte{},
			contentType: httpheaders.ContentTypeTextPlain,
			cases: []testcase{
				{method: http.MethodGet, url: "/value/counter"},
				{method: http.MethodGet, url: "/value/counter/"},
				{method: http.MethodGet, url: "/value/counter/id1"},
				{method: http.MethodGet, url: "/value/counter/id1/"},
				{method: http.MethodGet, url: "/value/counter/id1/123"},
				{method: http.MethodGet, url: "/value/counter/id1/123/"},
				{method: http.MethodGet, url: "/value/counter/id1/123/something"},
				{method: http.MethodGet, url: "/value/gauge"},
				{method: http.MethodGet, url: "/value/gauge/id1"},
			},
		},
		handler.IdentValueJSON: {
			status:      http.StatusOK,
			body:        []byte(`{}`),
			contentType: httpheaders.ContentTypeApplicationJSON,
			cases: []testcase{
				{method: http.MethodPost, url: "/value"},
				{method: http.MethodPost, url: "/value/"},
			},
		},
		handler.IdentPing: {
			status:      http.StatusOK,
			body:        []byte(`OK`),
			contentType: httpheaders.ContentTypeTextPlain,
			cases: []testcase{
				{method: http.MethodGet, url: "/ping"},
			},
		},
	}
	for ident, tt := range tests {
		cases := make([]testHandlerCase, 0, len(tt.cases))
		for _, tc := range tt.cases {
			cases = append(cases, testHandlerCase{
				args: testHandlerArgs{
					method:   tc.method,
					url:      tc.url,
					bodyData: handlertest.NewBodyData(t, nil),
				},
				want: testHandlerWant{
					status:      tt.status,
					contentType: tt.contentType,
					body:        tt.body,
				},
			})
		}
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tt.contentType.Apply(w.Header())
			w.WriteHeader(tt.status)
			w.Write(tt.body)
		})
		testHandlerTable(t, ident, fn, cases)
	}
}

func testHandlerFuncMirrorMetric(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		var m model.Metric
		err := json.NewDecoder(r.Body).Decode(&m)
		assert.NoError(t, err)
		httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(m)
		assert.NoError(t, err)
	}
}

func testRouterServeHTTPMiddleware(t *testing.T) {
	tests := map[string]struct {
		ident handler.Ident
		fn    http.HandlerFunc
		cases []testHandlerCase
	}{
		"recovers from panic": {
			ident: handler.IdentUpdateJSON,
			fn: func(w http.ResponseWriter, r *http.Request) {
				panic("oops")
			},
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:   http.MethodPost,
						url:      "/update",
						bodyData: handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
					},
					want: testHandlerWant{
						status:      http.StatusInternalServerError,
						contentType: httpheaders.ContentTypeTextPlain.UTF8(),
						body:        []byte("recovered from panic\n"),
					},
				},
			},
		},
		"decompresses request": {
			ident: handler.IdentUpdateJSON,
			fn:    testHandlerFuncMirrorMetric(t),
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:         http.MethodPost,
						url:            "/update",
						bodyData:       handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
						shouldCompress: true,
					},
					want: testHandlerWant{
						status:      http.StatusOK,
						contentType: httpheaders.ContentTypeApplicationJSON,
						body:        []byte(`{"id": "id1", "type": "counter", "delta": 123}`),
					},
				},
			},
		},
		"compresses response": {
			ident: handler.IdentUpdateJSON,
			fn:    testHandlerFuncMirrorMetric(t),
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:            http.MethodPost,
						url:               "/update",
						bodyData:          handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
						shouldCompress:    false,
						acceptedEncodings: []httpheaders.ContentEncoding{httpheaders.ContentEncodingGzip},
					},
					want: testHandlerWant{
						status:          http.StatusOK,
						contentType:     httpheaders.ContentTypeApplicationJSON,
						contentEncoding: httpheaders.ContentEncodingGzip,
						body:            []byte(`{"id": "id1", "type": "counter", "delta": 123}`),
					},
				},
			},
		},
		"not compresses response if not accepted by client": {
			ident: handler.IdentUpdateJSON,
			fn:    testHandlerFuncMirrorMetric(t),
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:            http.MethodPost,
						url:               "/update",
						bodyData:          handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
						shouldCompress:    false,
						acceptedEncodings: []httpheaders.ContentEncoding{httpheaders.ContentEncodingDeflate},
					},
					want: testHandlerWant{
						status:          http.StatusOK,
						contentType:     httpheaders.ContentTypeApplicationJSON,
						contentEncoding: httpheaders.ContentEncodingEmpty,
						body:            []byte(`{"id": "id1", "type": "counter", "delta": 123}`),
					},
				},
			},
		},
		"decompresses request, compresses response": {
			ident: handler.IdentUpdateJSON,
			fn:    testHandlerFuncMirrorMetric(t),
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:            http.MethodPost,
						url:               "/update",
						bodyData:          handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
						shouldCompress:    true,
						acceptedEncodings: []httpheaders.ContentEncoding{httpheaders.ContentEncodingGzip},
					},
					want: testHandlerWant{
						status:          http.StatusOK,
						contentType:     httpheaders.ContentTypeApplicationJSON,
						contentEncoding: httpheaders.ContentEncodingGzip,
						body:            []byte(`{"id": "id1", "type": "counter", "delta": 123}`),
					},
				},
			},
		},
		"not compresses binary response": {
			ident: handler.IdentDefault,
			fn: func(w http.ResponseWriter, r *http.Request) {
				httpheaders.ContentType("image/png").Apply(w.Header())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`binary data`))

			},
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:   http.MethodGet,
						url:      "/get/some/binary/data",
						bodyData: handlertest.NewBodyData(t, nil),
					},
					want: testHandlerWant{
						status:          http.StatusOK,
						contentType:     httpheaders.ContentType("image/png"),
						contentEncoding: httpheaders.ContentEncodingEmpty,
						body:            []byte(`binary data`),
					},
				},
			},
		},
		// FIXME:
		// Such requests should not be accepted, but we are forced to accept them
		// because of `go-autotests` which do not sign their requests; see
		// https://github.com/Yandex-Practicum/go-autotests/blob/0591b1dbbcbcf741c41c8eca0718bf676ed7307f/cmd/metricstest_v2/iteration14_test.go#L462
		"when secret key present, accepts request without signature and signs response": {
			ident: handler.IdentUpdateJSON,
			fn:    testHandlerFuncMirrorMetric(t),
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:              http.MethodPost,
						url:                 "/update",
						bodyData:            handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
						shouldCompress:      true,
						acceptedEncodings:   []httpheaders.ContentEncoding{httpheaders.ContentEncodingGzip},
						shouldSign:          false,
						expectNoHandlerCall: false,
					},
					want: testHandlerWant{
						status:          http.StatusOK,
						contentType:     httpheaders.ContentTypeApplicationJSON,
						contentEncoding: httpheaders.ContentEncodingGzip,
						body:            []byte(`{"id": "id1", "type": "counter", "delta": 123}`),
						hash:            httpheaders.HashSHA256("6159e85f20e3dc1f908997d0150102acf21c52efe580ff4b3c24c2076801dc4e"), // https://tools.onecompiler.com/hmac-sha256
					},
					secretKey: []byte(`super-secret-key`),
				},
			},
		},
		"when secret key present, rejects request with incorrect signature": {
			ident: handler.IdentUpdateJSON,
			fn:    testHandlerFuncMirrorMetric(t),
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:              http.MethodPost,
						url:                 "/update",
						bodyData:            handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
						shouldCompress:      true,
						acceptedEncodings:   []httpheaders.ContentEncoding{httpheaders.ContentEncodingGzip},
						shouldSign:          false,
						overrideHash:        httpheaders.GetHashSHA256FromBytes([]byte(`incorrect signature`)),
						expectNoHandlerCall: true,
					},
					want: testHandlerWant{
						status:          http.StatusBadRequest,
						contentType:     httpheaders.ContentTypeTextPlain.UTF8(),
						contentEncoding: httpheaders.ContentEncodingEmpty,
						body:            []byte(`signature mismatch` + "\n"),
						hash:            httpheaders.HashSHA256Empty,
					},
					secretKey: []byte(`super-secret-key`),
				},
			},
		},
		"when secret key present, validates request signature and signs response": {
			ident: handler.IdentUpdateJSON,
			fn:    testHandlerFuncMirrorMetric(t),
			cases: []testHandlerCase{
				{
					args: testHandlerArgs{
						method:            http.MethodPost,
						url:               "/update",
						bodyData:          handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 123)),
						shouldCompress:    true,
						acceptedEncodings: []httpheaders.ContentEncoding{httpheaders.ContentEncodingGzip},
						shouldSign:        true,
					},
					want: testHandlerWant{
						status:          http.StatusOK,
						contentType:     httpheaders.ContentTypeApplicationJSON,
						contentEncoding: httpheaders.ContentEncodingGzip,
						body:            []byte(`{"id": "id1", "type": "counter", "delta": 123}`),
						hash:            httpheaders.HashSHA256("6159e85f20e3dc1f908997d0150102acf21c52efe580ff4b3c24c2076801dc4e"), // https://tools.onecompiler.com/hmac-sha256
					},
					secretKey: []byte(`super-secret-key`),
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			testHandlerTable(t, tt.ident, tt.fn, tt.cases)
		})
	}
}

func testHandlerTable(t *testing.T, handlerID handler.Ident, handlerFn http.Handler, cases []testHandlerCase) {
	for _, tc := range cases {
		name := fmt.Sprintf("%s %s %s", handlerID, tc.args.method, tc.args.url)
		t.Run(name, func(t *testing.T) {
			testHandlerRun(t, handlerID, handlerFn, tc.args, tc.want, tc.secretKey)
		})
	}
}

func testHandlerRun(t *testing.T, handlerID handler.Ident, handlerFn http.Handler, args testHandlerArgs, want testHandlerWant, secretKey []byte) {
	t.Helper()

	// Arrange

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewTestLogger()

	handlers := handler.NewRegistry(log.NewNoopLogger(), servicetest.NewMockMetricStorer(ctrl), servicetest.NewMockStoragePinger(ctrl))
	for id := range handlers {
		m := handlertest.NewMockHandler(ctrl)
		handlers[id] = m
		if id == handlerID {
			if args.expectNoHandlerCall {
				m.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(0)
			} else {
				m.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Do(handlerFn)
			}
		} else {
			// We do not expect other handlers to get called, but if
			// a wrong handler did get called, we would like to know its name to facilitate tests debugging.
			m.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).AnyTimes().Do(
				func(w http.ResponseWriter, r *http.Request) {
					ctrl.T.Fatalf("expected call to handler %v, got handler %v", handlerID, id)
				},
			)
		}
	}

	signer := hmacsigner.NewHMACSigner(secretKey)
	rtr, err := New(logger, handlers, signer)
	require.NoError(t, err)

	ts := httptest.NewServer(rtr)
	defer ts.Close()

	req := args.bodyData.NewRequest(args.method, ts.URL+args.url, args.shouldCompress)
	for _, enc := range args.acceptedEncodings {
		enc.MakeAccepted(req.Header)
	}
	if args.shouldSign {
		args.bodyData.GetDataSignature(signer).Apply(req.Header)
	}
	if args.overrideHash != httpheaders.HashSHA256Empty {
		args.overrideHash.Apply(req.Header)
	}

	// Act

	resp, err := ts.Client().Do(req)
	defer func() { _ = resp.Body.Close() }()
	require.NoError(t, err)
	bodyData := handlertest.NewBodyDataFromResponse(t, resp)

	// Assert

	assert.Equal(t, args.url, resp.Request.URL.Path)
	assert.Equal(t, want.status, resp.StatusCode)
	assert.Truef(t, want.contentEncoding.Matches(resp.Header), "expected %v encoding, got %v", want.contentEncoding, httpheaders.GetContentEncoding(resp.Header))
	assert.Truef(t, want.hash.Matches(resp.Header), "expected hash %v, got %v", want.hash, httpheaders.GetHashSHA256(resp.Header))

	bodyData.AssertType(want.contentType)
	bodyData.AssertData(want.body)

	events := logger.RecordedEvents()
	assert.GreaterOrEqual(t, len(events), 1)
	found := events.FindMatchingEvents(
		log.LevelInfo, "processed request",
		log.Str("uri", resp.Request.URL.Path),
		log.Str("method", args.method),
		log.Int("status", resp.StatusCode),
	)
	require.Len(t, found, 1)
	e := found[0]
	fsize := e.Fields().GetFieldByKey("size")
	require.NotNil(t, fsize)
	assert.Equal(t, bodyData.Len(), fsize.Value)
	assert.NotNil(t, e.Fields().GetFieldByKey("duration"))
	assert.NotNil(t, e.Fields().GetFieldByKey("request_id"))
}

func TestRoute_Validate(t *testing.T) {
	type fields struct {
		patterns []string
		handler  http.Handler
	}
	type want struct {
		err error
	}
	type testcase struct {
		fields fields
		want   want
	}
	exampleHandler := http.NewServeMux()
	tests := map[string]testcase{
		"empty patterns fails": {
			fields: fields{patterns: []string{}, handler: exampleHandler},
			want:   want{err: ErrRouteEmptyPatterns},
		},
		"empty handler fails": {
			fields: fields{patterns: []string{"GET /"}, handler: nil},
			want:   want{err: ErrRouteEmptyHandler},
		},
		"duplicate patterns fail": {
			fields: fields{patterns: []string{"GET /", "POST /123", "GET /"}, handler: exampleHandler},
			want:   want{err: ErrRouteDuplicatePatterns},
		},
		"unique patterns pass": {
			fields: fields{patterns: []string{"GET /", "POST /123"}, handler: http.NewServeMux()},
			want:   want{err: nil},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := &Route{
				patterns: tt.fields.patterns,
				handler:  tt.fields.handler,
			}
			err := r.Validate()
			require.Equal(t, tt.want.err, err)
		})
	}
}

func TestNewRoute(t *testing.T) {
	type args struct {
		handler  http.Handler
		patterns []string
	}
	type want struct {
		got Route
		err error
	}
	type testcase struct {
		args args
		want want
	}
	exampleHandler := http.NewServeMux()
	tests := map[string]testcase{
		"empty patterns fail": {
			args: args{handler: exampleHandler, patterns: []string{}},
			want: want{got: Route{}, err: ErrRouteEmptyPatterns},
		},
		"nil handler fails": {
			args: args{handler: nil, patterns: []string{"GET /"}},
			want: want{got: Route{}, err: ErrRouteEmptyHandler},
		},
		"unique patterns pass": {
			args: args{handler: exampleHandler, patterns: []string{"GET /123", "POST /456"}},
			want: want{got: Route{patterns: []string{"GET /123", "POST /456"}, handler: exampleHandler}, err: nil},
		},
		"duplicate patterns removed": {
			args: args{handler: exampleHandler, patterns: []string{"GET /123", "POST /456", "GET /123"}},
			want: want{got: Route{patterns: []string{"GET /123", "POST /456"}, handler: exampleHandler}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewRoute(tt.args.handler, tt.args.patterns...)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type args struct {
		logger   log.Logger
		handlers handler.Registry
		signer   hmacsigner.HMACSigner
	}
	type want struct {
		wantErr bool
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty handlers fail": {
			args: args{logger: log.NewNoopLogger(), handlers: handler.Registry{}, signer: hmacsigner.NewHMACSigner(nil)},
			want: want{wantErr: true},
		},
		"proper handlers pass": {
			args: args{logger: log.NewNoopLogger(), handlers: handler.NewRegistry(log.NewNoopLogger(), servicetest.NewMockMetricStorer(ctrl), servicetest.NewMockStoragePinger(ctrl)), signer: hmacsigner.NewHMACSigner(nil)},
		},
		"nil logger replaced by noop": {
			args: args{logger: nil, handlers: handler.NewRegistry(log.NewNoopLogger(), servicetest.NewMockMetricStorer(ctrl), servicetest.NewMockStoragePinger(ctrl)), signer: hmacsigner.NewHMACSigner(nil)},
		},
		"nil signer fails": {
			args: args{logger: log.NewNoopLogger(), handlers: handler.NewRegistry(log.NewNoopLogger(), servicetest.NewMockMetricStorer(ctrl), servicetest.NewMockStoragePinger(ctrl))},
			want: want{wantErr: true},
		},
		"proper signer passes": {
			args: args{logger: log.NewNoopLogger(), handlers: handler.NewRegistry(log.NewNoopLogger(), servicetest.NewMockMetricStorer(ctrl), servicetest.NewMockStoragePinger(ctrl)), signer: hmacsigner.NewHMACSigner(nil)},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.logger, tt.args.handlers, tt.args.signer)
			if tt.want.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
			testNewChiRoutes(t, got)
		})
	}
}

func testNewChiRoutes(t *testing.T, rt *Router) {
	wantRoutes := map[string]bool{
		"GET /":          true,
		"POST /update":   true,
		"POST /update/":  true,
		"POST /update/*": true,
		"POST /value":    true,
		"POST /value/":   true,
		"GET /value/*":   true,
		"GET /ping":      true,
	}
	testChiRoutes(t, rt.mux, wantRoutes)
}

func testChiRoutes(t *testing.T, mux chi.Router, wantRoutes map[string]bool) {
	t.Helper()

	// Arrange
	require.NotNil(t, mux)
	require.NotEmpty(t, wantRoutes)

	seenRoutes := map[string]bool{}
	walkFn := func(method string, route string, hh http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		key := fmt.Sprintf("%s %s", method, route)
		if wantRoutes[key] {
			seenRoutes[key] = true
		}
		return nil
	}

	// Act
	err := chi.Walk(mux, walkFn)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, wantRoutes, seenRoutes)
}

func Test_configureChiRouter(t *testing.T) {
	type args struct {
		mux         *chi.Mux
		middlewares []middleware.Middleware
		routes      []Route
	}
	type want struct {
		chiRoutes map[string]bool
		wantErr   bool
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty middlewares and routes do nothing": {
			args: args{mux: chi.NewRouter()},
		},
		"nil middleware fails": {
			args: args{mux: chi.NewRouter(), middlewares: []middleware.Middleware{middleware.RequestID(), nil}},
			want: want{wantErr: true},
		},
		"invalid route fails": {
			args: args{mux: chi.NewRouter(), middlewares: []middleware.Middleware{middleware.RequestID()}, routes: []Route{{handler: http.NewServeMux()}}},
			want: want{wantErr: true},
		},
		"given routes configured": {
			args: args{
				mux:         chi.NewRouter(),
				middlewares: []middleware.Middleware{middleware.RequestID()},
				routes: []Route{
					{handler: http.NewServeMux(), patterns: []string{"GET /"}},
					{handler: http.NewServeMux(), patterns: []string{"POST /"}},
					{handler: http.NewServeMux(), patterns: []string{"/123/456", "PUT /update"}},
				},
			},
			want: want{
				chiRoutes: map[string]bool{
					"GET /":           true,
					"POST /":          true,
					"GET /123/456":    true,
					"POST /123/456":   true,
					"PUT /123/456":    true,
					"PATCH /123/456":  true,
					"HEAD /123/456":   true,
					"DELETE /123/456": true,
					"PUT /update":     true,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := configureChiRouter(tt.args.mux, tt.args.middlewares, tt.args.routes)
			if tt.want.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if len(tt.want.chiRoutes) > 0 {
				testChiRoutes(t, tt.args.mux, tt.want.chiRoutes)
			}
		})
	}
}

func TestMiddlewares(t *testing.T) {
	// covered by [TestRouter_ServeHTTP]
	t.SkipNow()
}

func TestRouteDefinitions(t *testing.T) {
	// covered by [TestRouter_ServeHTTP]
	t.SkipNow()
}

func TestRoutes(t *testing.T) {
	type args struct {
		handlers handler.Registry
	}
	type want struct {
		wantErr bool
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty registry fails": {
			args: args{handlers: handler.Registry{}},
			want: want{wantErr: true},
		},
		"partial registry fails": {
			args: args{handlers: handler.Registry{
				handler.IdentDefault: http.NewServeMux(),
			}},
			want: want{wantErr: true},
		},
		"full registry with nil handlers fails": {
			args: args{handlers: func() handler.Registry {
				reg := handler.Registry{}
				for _, rd := range RouteDefinitions() {
					reg[rd.ident] = nil
				}
				return reg
			}()},
			want: want{wantErr: true},
		},
		"full registry succeeds": {
			args: args{handlers: func() handler.Registry {
				reg := handler.Registry{}
				for _, rd := range RouteDefinitions() {
					reg[rd.ident] = http.NewServeMux()
				}
				return reg
			}()},
			want: want{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Routes(tt.args.handlers)
			if tt.want.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, got)
		})
	}
}
