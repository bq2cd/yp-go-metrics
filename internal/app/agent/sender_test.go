package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner/hmacsignertest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mockSender struct {
	mock.Mock
	wantErr      func(model.Metric) error
	wantBatchErr func(model.MetricSet) (model.MetricSet, error)
}

func (m *mockSender) Send(ctx context.Context, metric model.Metric) error {
	m.Called(ctx, metric)
	if m.wantErr != nil {
		return m.wantErr(metric)
	}
	return nil
}

func (m *mockSender) SendBatch(ctx context.Context, metrics model.MetricSet) (model.MetricSet, error) {
	m.Called(ctx, metrics)
	if m.wantBatchErr != nil {
		return m.wantBatchErr(metrics)
	}
	return metrics, nil
}

func Test_senderPlain_Send(t *testing.T) {
	type fields struct {
		client   *resty.Client
		deadline time.Duration
	}
	type responder struct {
		contentType httpheaders.ContentType
		status      int
		timeout     time.Duration
	}
	type args struct {
		method    string
		urlRegexp *regexp.Regexp
		metric    model.Metric
	}
	type want struct {
		httpCall string
		checkErr func(*testing.T, error)
	}
	tests := []struct {
		name      string
		fields    fields
		responder responder
		args      args
		want      want
	}{
		{
			name: "send counter without value",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				metric:    model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			},
			want: want{
				checkErr: func(t *testing.T, err error) { assert.ErrorIs(t, err, urlpath.ErrMissingMetricValue) },
			},
		},
		{
			name: "send gauge without value",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				metric:    model.Metric{Type: model.MetricTypeGauge, ID: "id1"},
			},
			want: want{
				checkErr: func(t *testing.T, err error) { assert.ErrorIs(t, err, urlpath.ErrMissingMetricValue) },
			},
		},
		{
			name: "send counter",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall: "POST http://localhost:1234/update/counter/id1/5",
				checkErr: func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "send gauge",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				metric:    model.NewGaugeMetric("id1", -5.5),
			},
			want: want{
				httpCall: "POST http://localhost:1234/update/gauge/id1/-5.5",
				checkErr: func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "server error",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusBadGateway,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall: "POST http://localhost:1234/update/counter/id1/5",
				checkErr: func(t *testing.T, err error) { assert.ErrorIs(t, err, ErrSenderResponseNotOK) },
			},
		},
		{
			name: "server timeout",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
				timeout:     75 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall: "POST http://localhost:1234/update/counter/id1/5",
				checkErr: func(t *testing.T, err error) { assert.Errorf(t, err, "request cancelled") },
			},
		},
		{
			name: "client context cancelled",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(200 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
				timeout:     500 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall: "POST http://localhost:1234/update/counter/id1/5",
				checkErr: func(t *testing.T, err error) { assert.Errorf(t, err, "context deadline exceeded") },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.fields.deadline)
			defer cancel()

			s := &senderPlain{
				client: tt.fields.client,
			}
			httpmock.ActivateNonDefault(s.client.GetClient())
			defer httpmock.Reset()
			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, func(r *http.Request) (*http.Response, error) {
				require.True(t, tt.responder.contentType.Matches(r.Header))
				time.Sleep(tt.responder.timeout)
				return httpmock.NewStringResponse(tt.responder.status, ""), nil
			})

			metric := tt.args.metric.Copy()

			err := s.Send(ctx, metric)

			defer func() {
				assert.Equal(t, tt.args.metric, metric)
			}()

			if tt.want.httpCall != "" {
				calls := httpmock.GetCallCountInfo()
				assert.Contains(t, calls, tt.want.httpCall)
				assert.Equal(t, 1, calls[tt.want.httpCall])
			}
			tt.want.checkErr(t, err)
		})
	}
}

func TestNewSenderPlain(t *testing.T) {
	type args struct {
		client *resty.Client
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, *senderPlain)
	}{
		{
			name: "default",
			args: args{client: resty.New()},
			assertion: func(t *testing.T, args args, got *senderPlain) {
				assert.Equal(t, args.client, got.client)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewSenderPlain(tt.args.client))
		})
	}
}

func TestNewSenderJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		client         *resty.Client
		retrierFactory retrymgr.RetrierFactory
		hmacSigner     hmacsigner.HMACSigner
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, *senderJSON)
	}{
		{
			name: "default",
			args: args{
				client: resty.New(),
				retrierFactory: func() retrymgr.RetrierFactory {
					return retrymgr.NewRetrierFactory(log.NewNoopLogger(), retrymgrtest.NewMockSleeper(ctrl), func() retrymgr.Strategy {
						return retrymgrtest.NewMockStrategy(ctrl)
					})
				}(),
				hmacSigner: hmacsigner.NewHMACSigner(nil),
			},
			assertion: func(t *testing.T, args args, got *senderJSON) {
				assert.Equal(t, args.client, got.client)
				assert.Equal(t, args.retrierFactory, got.retrierFactory)
				assert.Equal(t, args.hmacSigner, got.hmacSigner)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewSenderJSON(tt.args.client, tt.args.retrierFactory, tt.args.hmacSigner))
		})
	}
}

func Test_senderJSON_Send(t *testing.T) {
	t.Run("nothing signed without secret key", func(t *testing.T) {
		testSenderJSONSendHelper(t, func(ctrl *gomock.Controller, numCalls int) *hmacsignertest.MockHMACSigner {
			m := hmacsignertest.NewMockHMACSigner(ctrl)
			m.EXPECT().Sign(gomock.Any()).Return(nil, hmacsigner.ErrMissingSecretKey).Times(numCalls)
			return m
		}, httpheaders.HashSHA256Empty)
	})

	t.Run("request body signed with secret key", func(t *testing.T) {
		testSenderJSONSendHelper(t, func(ctrl *gomock.Controller, numCalls int) *hmacsignertest.MockHMACSigner {
			m := hmacsignertest.NewMockHMACSigner(ctrl)
			m.EXPECT().Sign(gomock.Any()).Return([]byte(`dummy-hash-bytes`), nil).Times(numCalls)
			return m
		}, httpheaders.HashSHA256("64756d6d792d686173682d6279746573"))
	})
}

func testSenderJSONSendHelper(t *testing.T, setupSigner func(*gomock.Controller, int) *hmacsignertest.MockHMACSigner, wantHashHeader httpheaders.HashSHA256) {
	t.Helper()

	type fields struct {
		client   *resty.Client
		deadline time.Duration
	}
	type responder struct {
		contentType     httpheaders.ContentType
		contentEncoding httpheaders.ContentEncoding
		status          int
		data            any
		timeout         time.Duration
	}
	type args struct {
		method    string
		urlRegexp *regexp.Regexp
		metric    model.Metric
	}
	type want struct {
		httpCall    string
		requestBody string
		checkErr    func(*testing.T, error)
		numRetries  int
	}
	tests := []struct {
		name      string
		fields    fields
		responder responder
		args      args
		want      want
	}{
		{
			name: "send counter without value",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			},
			want: want{
				checkErr: func(t *testing.T, err error) { assert.ErrorIs(t, err, ErrSenderEmptyMetric) },
			},
		},
		{
			name: "send gauge without value",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeTextPlain,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.Metric{Type: model.MetricTypeGauge, ID: "id1"},
			},
			want: want{
				checkErr: func(t *testing.T, err error) { assert.ErrorIs(t, err, ErrSenderEmptyMetric) },
			},
		},
		{
			name: "send counter",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/update/",
				requestBody: `{ "id": "id1", "type": "counter", "delta": 5}`,
				checkErr:    func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "send gauge",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.NewGaugeMetric("id1", -5.5),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/update/",
				requestBody: `{ "id": "id1", "type": "gauge", "value": -5.5}`,
				checkErr:    func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "server error",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusBadGateway,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/update/",
				requestBody: `{ "id": "id1", "type": "counter", "delta": 5}`,
				checkErr:    func(t *testing.T, err error) { assert.ErrorIs(t, err, ErrSenderResponseNotOK) },
				numRetries:  3,
			},
		},
		{
			name: "server timeout",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusBadGateway,
				timeout:     75 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/update/",
				requestBody: `{ "id": "id1", "type": "counter", "delta": 5}`,
				checkErr:    func(t *testing.T, err error) { assert.Errorf(t, err, "request cancelled") },
				numRetries:  3,
			},
		},
		{
			name: "client context cancelled",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusBadGateway,
				timeout:     500 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/update/",
				requestBody: `{ "id": "id1", "type": "counter", "delta": 5}`,
				checkErr:    func(t *testing.T, err error) { assert.Errorf(t, err, "context deadline exceeded") },
				numRetries:  1,
			},
		},
		{
			name: "send counter with gzip compression",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentEncoding: httpheaders.ContentEncodingGzip,
				contentType:     httpheaders.ContentTypeApplicationJSON,
				status:          http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.NewCounterMetric("id1", 5),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/update/",
				requestBody: `{ "id": "id1", "type": "counter", "delta": 5}`,
				checkErr:    func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
		{
			name: "send gauge with gzip compression",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentEncoding: httpheaders.ContentEncodingGzip,
				contentType:     httpheaders.ContentTypeApplicationJSON,
				status:          http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/?$"),
				metric:    model.NewGaugeMetric("id1", -5.5),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/update/",
				requestBody: `{ "id": "id1", "type": "gauge", "value": -5.5}`,
				checkErr:    func(t *testing.T, err error) { assert.NoError(t, err) },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			retrierFactory := retrymgr.NewRetrierFactory(log.NewNoopLogger(),
				func() retrymgr.Sleeper {
					s := retrymgrtest.NewMockSleeper(ctrl)
					s.EXPECT().Sleep(gomock.Any(), 1*time.Millisecond).Return(nil).Times(tt.want.numRetries)
					return s
				}(),
				func() retrymgr.Strategy {
					s := retrymgrtest.NewMockStrategy(ctrl)
					s.EXPECT().Name().Return("mock_strategy")
					if tt.want.numRetries > 0 {
						s.EXPECT().NextDelay().Return(1*time.Millisecond, true).Times(tt.want.numRetries)
						s.EXPECT().NextDelay().Return(time.Duration(0), false).Times(1)
					}
					return s
				},
			)

			var hmacSigner *hmacsignertest.MockHMACSigner
			if tt.want.httpCall != "" {
				hmacSigner = setupSigner(ctrl, tt.want.numRetries+1)
			} else {
				hmacSigner = setupSigner(ctrl, 0)
			}

			ctx, cancel := context.WithTimeout(t.Context(), tt.fields.deadline)
			defer cancel()

			shouldCompress := tt.responder.contentEncoding != httpheaders.ContentEncodingEmpty
			s := &senderJSON{
				client:         tt.fields.client,
				hmacSigner:     hmacSigner,
				retrierFactory: retrierFactory,
				shouldCompress: shouldCompress,
			}
			httpmock.ActivateNonDefault(s.client.GetClient())
			defer httpmock.Reset()

			rbody := bytes.NewBuffer(nil)
			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, func(r *http.Request) (*http.Response, error) {
				require.True(t, tt.responder.contentType.Matches(r.Header))
				require.True(t, tt.responder.contentEncoding.Matches(r.Header))
				require.Truef(t, wantHashHeader.Matches(r.Header), "hash header mismatch")
				rbody.Reset()
				_, err := io.Copy(rbody, r.Body)
				require.NoError(t, err)
				time.Sleep(tt.responder.timeout)
				return httpmock.NewJsonResponse(tt.responder.status, tt.responder.data)
			})

			metric := tt.args.metric.Copy()
			err := s.Send(ctx, metric)

			defer func() {
				assert.Equal(t, tt.args.metric, metric)
			}()

			if tt.want.httpCall != "" {
				calls := httpmock.GetCallCountInfo()
				assert.Contains(t, calls, tt.want.httpCall)
				assert.Equal(t, tt.want.numRetries+1, calls[tt.want.httpCall])
			}
			tt.want.checkErr(t, err)
			if tt.want.requestBody == "" {
				return
			}
			var body string
			if shouldCompress {
				rgz, err := gzip.NewReader(rbody)
				require.NoError(t, err)
				b, err := io.ReadAll(rgz)
				require.NoError(t, err)
				body = string(b)
			} else {
				body = rbody.String()
			}
			assert.JSONEq(t, tt.want.requestBody, body)
		})
	}
}

func Test_senderJSON_setBody(t *testing.T) {
	type fields struct {
		client         *resty.Client
		shouldCompress bool
	}
	type args struct {
		req *resty.Request
		r   io.Reader
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &senderJSON{
				client:         tt.fields.client,
				shouldCompress: tt.fields.shouldCompress,
			}
			tt.assertion(t, s.setBody(tt.args.req, tt.args.r))
		})
	}
}

func Test_senderJSON_SendBatch(t *testing.T) {
	t.Run("nothing signed without secret key", func(t *testing.T) {
		testSenderJSONSendBatchHelper(t, func(ctrl *gomock.Controller, numCalls int) *hmacsignertest.MockHMACSigner {
			m := hmacsignertest.NewMockHMACSigner(ctrl)
			m.EXPECT().Sign(gomock.Any()).Return(nil, hmacsigner.ErrMissingSecretKey).Times(numCalls)
			return m
		}, httpheaders.HashSHA256Empty)
	})
	t.Run("request body signed with secret key", func(t *testing.T) {
		testSenderJSONSendBatchHelper(t, func(ctrl *gomock.Controller, numCalls int) *hmacsignertest.MockHMACSigner {
			m := hmacsignertest.NewMockHMACSigner(ctrl)
			m.EXPECT().Sign(gomock.Any()).Return([]byte(`dummy-hash-bytes`), nil).Times(numCalls)
			return m
		}, httpheaders.HashSHA256("64756d6d792d686173682d6279746573"))
	})
}

func testSenderJSONSendBatchHelper(t *testing.T, setupSigner func(*gomock.Controller, int) *hmacsignertest.MockHMACSigner, wantHashHeader httpheaders.HashSHA256) {
	t.Helper()

	type fields struct {
		client   *resty.Client
		deadline time.Duration
	}
	type responder struct {
		contentType     httpheaders.ContentType
		contentEncoding httpheaders.ContentEncoding
		status          int
		data            any
		timeout         time.Duration
	}
	type args struct {
		method    string
		urlRegexp *regexp.Regexp
		metrics   model.MetricSet
	}
	type want struct {
		httpCall    string
		requestBody string
		got         model.MetricSet
		checkErr    func(*testing.T, error)
		numRetries  int
	}
	type testcase struct {
		fields    fields
		responder responder
		args      args
		want      want
	}
	tests := map[string]testcase{
		"empty metrics are not sent": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.Metric{},
					model.Metric{Type: model.MetricTypeCounter},
					model.Metric{Type: model.MetricTypeGauge},
					model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
					model.Metric{Type: model.MetricTypeGauge, ID: "id2"},
				),
			},
			want: want{
				got:      model.NewMetricSet(),
				checkErr: func(t *testing.T, err error) { require.NoError(t, err) },
			},
		},
		"single metric is sent OK": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
				data: []model.Metric{
					model.NewCounterMetric("id1", 5),
				},
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
				),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/updates/",
				requestBody: `[{ "id": "id1", "type": "counter", "delta": 5}]`,
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
				),
				checkErr: func(t *testing.T, err error) { require.NoError(t, err) },
			},
		},
		"multiple new metrics are sent OK": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
				data: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				},
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				),
			},
			want: want{
				httpCall: "POST http://localhost:1234/updates/",
				requestBody: `[
					{ "id": "id1", "type": "counter", "delta": 5},
					{ "id": "id2", "type": "counter", "delta": -5},
					{ "id": "id3", "type": "gauge", "value": 1.5},
					{ "id": "id4", "type": "gauge", "value": -1.5}
				]`,
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				),
				checkErr: func(t *testing.T, err error) { require.NoError(t, err) },
			},
		},
		"multiple existing metrics are sent OK, server returns updated": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
				data: []model.Metric{
					model.NewCounterMetric("id1", 15),
					model.NewCounterMetric("id2", 0),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				},
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				),
			},
			want: want{
				httpCall: "POST http://localhost:1234/updates/",
				requestBody: `[
					{ "id": "id1", "type": "counter", "delta": 5},
					{ "id": "id2", "type": "counter", "delta": -5},
					{ "id": "id3", "type": "gauge", "value": 1.5},
					{ "id": "id4", "type": "gauge", "value": -1.5}
				]`,
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 15),
					model.NewCounterMetric("id2", 0),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				),
				checkErr: func(t *testing.T, err error) { require.NoError(t, err) },
			},
		},
		"multiple metrics are compressed and sent OK": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentEncoding: httpheaders.ContentEncodingGzip,
				contentType:     httpheaders.ContentTypeApplicationJSON,
				status:          http.StatusOK,
				data: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				},
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				),
			},
			want: want{
				httpCall: "POST http://localhost:1234/updates/",
				requestBody: `[
					{ "id": "id1", "type": "counter", "delta": 5},
					{ "id": "id2", "type": "counter", "delta": -5},
					{ "id": "id3", "type": "gauge", "value": 1.5},
					{ "id": "id4", "type": "gauge", "value": -1.5}
				]`,
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -5),
					model.NewGaugeMetric("id3", 1.5),
					model.NewGaugeMetric("id4", -1.5),
				),
				checkErr: func(t *testing.T, err error) { require.NoError(t, err) },
			},
		},
		"client context cancelled": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
				data: []model.Metric{
					model.NewCounterMetric("id1", 5),
				},
				timeout: 500 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
				),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/updates/",
				requestBody: `[{ "id": "id1", "type": "counter", "delta": 5}]`,
				got:         model.NewMetricSet(),
				checkErr:    func(t *testing.T, err error) { require.Errorf(t, err, "context deadline exceeded") },
				numRetries:  1,
			},
		},
		"server timeout": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusOK,
				data: []model.Metric{
					model.NewCounterMetric("id1", 5),
				},
				timeout: 75 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
				),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/updates/",
				requestBody: `[{ "id": "id1", "type": "counter", "delta": 5}]`,
				got:         model.NewMetricSet(),
				checkErr:    func(t *testing.T, err error) { require.Errorf(t, err, "request canceled") },
				numRetries:  3,
			},
		},
		"server error": {
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: httpheaders.ContentTypeApplicationJSON,
				status:      http.StatusBadGateway,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/updates/?$"),
				metrics: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
				),
			},
			want: want{
				httpCall:    "POST http://localhost:1234/updates/",
				requestBody: `[{ "id": "id1", "type": "counter", "delta": 5}]`,
				got:         model.NewMetricSet(),
				checkErr:    func(t *testing.T, err error) { require.ErrorIs(t, err, ErrSenderResponseNotOK) },
				numRetries:  3,
			},
		},
	}
	for name, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		retrierFactory := retrymgr.NewRetrierFactory(log.NewNoopLogger(),
			func() retrymgr.Sleeper {
				s := retrymgrtest.NewMockSleeper(ctrl)
				s.EXPECT().Sleep(gomock.Any(), 1*time.Millisecond).Return(nil).Times(tt.want.numRetries)
				return s
			}(),
			func() retrymgr.Strategy {
				s := retrymgrtest.NewMockStrategy(ctrl)
				s.EXPECT().Name().Return("mock_strategy")
				if tt.want.numRetries > 0 {
					s.EXPECT().NextDelay().Return(1*time.Millisecond, true).Times(tt.want.numRetries)
					s.EXPECT().NextDelay().Return(time.Duration(0), false).Times(1)
				}
				return s
			},
		)

		var hmacSigner *hmacsignertest.MockHMACSigner
		if tt.want.httpCall != "" {
			hmacSigner = setupSigner(ctrl, tt.want.numRetries+1)
		} else {
			hmacSigner = setupSigner(ctrl, 0)
		}

		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.fields.deadline)
			defer cancel()

			shouldCompress := tt.responder.contentEncoding != httpheaders.ContentEncodingEmpty
			s := &senderJSON{
				client:         tt.fields.client,
				retrierFactory: retrierFactory,
				hmacSigner:     hmacSigner,
				shouldCompress: shouldCompress,
			}
			httpmock.ActivateNonDefault(s.client.GetClient())
			defer httpmock.Reset()

			rbody := bytes.NewBuffer(nil)
			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, func(r *http.Request) (*http.Response, error) {
				require.True(t, tt.responder.contentType.Matches(r.Header))
				require.True(t, tt.responder.contentEncoding.Matches(r.Header))
				require.Truef(t, wantHashHeader.Matches(r.Header), "hash header mismatch")
				rbody.Reset()
				_, err := io.Copy(rbody, r.Body)
				require.NoError(t, err)
				time.Sleep(tt.responder.timeout)
				return httpmock.NewJsonResponse(tt.responder.status, tt.responder.data)
			})

			// Act
			orig := tt.args.metrics
			got, err := s.SendBatch(ctx, tt.args.metrics)

			// Assert
			tt.want.checkErr(t, err)
			assert.Equal(t, tt.want.got, got)
			assert.Equal(t, orig, tt.args.metrics)

			if tt.want.httpCall != "" {
				calls := httpmock.GetCallCountInfo()
				assert.Contains(t, calls, tt.want.httpCall)
				assert.Equal(t, tt.want.numRetries+1, calls[tt.want.httpCall])
			}
			if tt.want.requestBody == "" {
				return
			}
			var body string
			if shouldCompress {
				rgz, err := gzip.NewReader(rbody)
				require.NoError(t, err)
				b, err := io.ReadAll(rgz)
				require.NoError(t, err)
				body = string(b)
			} else {
				body = rbody.String()
			}
			var wantReqMetrics, gotReqMetrics []model.Metric
			err = json.Unmarshal([]byte(tt.want.requestBody), &wantReqMetrics)
			require.NoError(t, err)
			err = json.Unmarshal([]byte(body), &gotReqMetrics)
			require.NoError(t, err)
			assert.ElementsMatch(t, wantReqMetrics, gotReqMetrics)
		})
	}
}
