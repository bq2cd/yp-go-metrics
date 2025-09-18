package agent

import (
	"context"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReporter struct {
	mock.Mock
	metrics []model.Metric
	timeout time.Duration
}

func (m *mockReporter) Report(metrics []model.Metric) error {
	m.Called(metrics)
	m.metrics = metrics
	if m.timeout > 0 {
		time.Sleep(m.timeout)
	}
	return nil
}

func Test_defaultReporter_Report(t *testing.T) {
	type fields struct {
		client *resty.Client
	}
	type args struct {
		method    string
		urlRegexp *regexp.Regexp
		responder func(assert.TestingT) httpmock.Responder
		metrics   []model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "send single counter",
			fields: fields{
				client: resty.New().SetBaseURL("http://localhost:1234"),
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metrics: []model.Metric{model.NewCounterMetric("id1", 5)},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.NoError(t, err)
			},
		},
		{
			name: "send multiple counters",
			fields: fields{
				client: resty.New().SetBaseURL("http://localhost:1234"),
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10)},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5", "POST http://localhost:1234/update/counter/id2/10"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.NoError(t, err)
			},
		},
		{
			name: "send multiple metrics",
			fields: fields{
				client: resty.New().SetBaseURL("http://localhost:1234"),
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{
					"POST http://localhost:1234/update/counter/id1/5",
					"POST http://localhost:1234/update/counter/id2/10",
					"POST http://localhost:1234/update/gauge/id1/-5",
					"POST http://localhost:1234/update/gauge/id2/-3.01",
				}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.NoError(t, err)
			},
		},
		{
			name: "server error on single metric",
			fields: fields{
				client: resty.New().SetBaseURL("http://localhost:1234"),
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						m, err := handler.NewMetricFromURLPath(r.URL.Path)
						assert.NoError(t, err)
						if m.Type == model.MetricTypeGauge && m.ID == "id1" {
							return httpmock.NewStringResponse(http.StatusInternalServerError, ""), nil
						}
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id2", 10), model.NewGaugeMetric("id1", -5), model.NewGaugeMetric("id2", -3.01)},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{
					"POST http://localhost:1234/update/counter/id1/5",
					"POST http://localhost:1234/update/counter/id2/10",
					"POST http://localhost:1234/update/gauge/id1/-5",
					"POST http://localhost:1234/update/gauge/id2/-3.01",
				}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.Errorf(t, err, "expected success")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &defaultReporter{
				context: t.Context(),
				client:  tt.fields.client,
			}
			httpmock.ActivateNonDefault(r.client.GetClient())
			defer httpmock.Reset()

			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, tt.args.responder(t))

			tt.assertion(t, r.Report(tt.args.metrics))
		})
	}
}

func TestNewDefaultReporter(t *testing.T) {
	type args struct {
		context context.Context
		client  *resty.Client
	}
	tests := []struct {
		name      string
		args      args
		assertion func(assert.TestingT, args, *defaultReporter)
	}{
		{
			name: "default initialisation",
			args: args{context: context.Background(), client: resty.New()},
			assertion: func(t assert.TestingT, want args, got *defaultReporter) {
				assert.Equal(t, want.context, got.context)
				assert.Equal(t, want.client, got.client)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewDefaultReporter(tt.args.context, tt.args.client))
		})
	}
}

func Test_defaultReporter_reportSingle(t *testing.T) {
	type fields struct {
		client   *resty.Client
		deadline time.Duration
	}
	type args struct {
		method    string
		urlRegexp *regexp.Regexp
		responder func(assert.TestingT) httpmock.Responder
		metric    model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "send counter",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metric: model.NewCounterMetric("id1", 5),
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.NoError(t, err)
			},
		},
		{
			name: "send gauge",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metric: model.NewGaugeMetric("id1", -5.5),
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/gauge/id1/-5.5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.NoError(t, err)
			},
		},
		{
			name: "server error",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						return httpmock.NewStringResponse(http.StatusBadGateway, ""), nil
					}
				},
				metric: model.NewCounterMetric("id1", 5),
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.Error(t, err)
			},
		},
		{
			name: "server timeout",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						time.Sleep(75 * time.Millisecond)
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metric: model.NewCounterMetric("id1", 5),
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.Errorf(t, err, "request cancelled")
			},
		},
		{
			name: "server deadline",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(200 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						time.Sleep(1 * time.Second)
						return httpmock.NewStringResponse(http.StatusOK, ""), nil
					}
				},
				metric: model.NewCounterMetric("id1", 5),
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				return assert.Errorf(t, err, "context deadline exceeded")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.fields.deadline)
			defer cancel()

			r := &defaultReporter{
				context: ctx,
				client:  tt.fields.client,
			}
			httpmock.ActivateNonDefault(r.client.GetClient())
			defer httpmock.Reset()

			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, tt.args.responder(t))

			tt.assertion(t, r.reportSingle(tt.args.metric))
		})
	}
}
