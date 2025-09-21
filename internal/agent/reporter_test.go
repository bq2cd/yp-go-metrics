package agent

import (
	"context"
	"net/http"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReporter struct {
	mock.Mock
	metrics []model.Metric
	timeout time.Duration
	mu      sync.Mutex
}

func (m *mockReporter) Report(metrics []model.Metric) error {
	m.Called(metrics)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = metrics
	if m.timeout > 0 {
		time.Sleep(m.timeout)
	}
	return nil
}

func Test_defaultReporter_Report(t *testing.T) {
	type fields struct {
		client   *resty.Client
		reported repository.Storage
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
			name: "invalid metric",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234"),
				reported: repository.NewMemStorage(),
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
				metrics: []model.Metric{{Type: model.MetricTypeCounter, ID: "id1"}},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.NotContains(t, calls, key)
				}
				return assert.ErrorIs(t, err, urlpath.ErrMissingMetricValue)
			},
		},
		{
			name: "send single counter",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234"),
				reported: repository.NewMemStorage(),
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
				client:   resty.New().SetBaseURL("http://localhost:1234"),
				reported: repository.NewMemStorage(),
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
			name: "send multiple counters with the same id",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234"),
				reported: repository.NewMemStorage(),
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
				metrics: []model.Metric{model.NewCounterMetric("id1", 5), model.NewCounterMetric("id1", -10), model.NewCounterMetric("id1", 7)},
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				calls := httpmock.GetCallCountInfo()
				keys := []string{
					"POST http://localhost:1234/update/counter/id1/5",
					"POST http://localhost:1234/update/counter/id1/-15",
					"POST http://localhost:1234/update/counter/id1/17",
				}
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
				client:   resty.New().SetBaseURL("http://localhost:1234"),
				reported: repository.NewMemStorage(),
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
				client:   resty.New().SetBaseURL("http://localhost:1234"),
				reported: repository.NewMemStorage(),
			},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						m, err := urlpath.NewOperationFromURLPath(r.URL.Path).ToMetric()
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
				context:  t.Context(),
				client:   tt.fields.client,
				reported: tt.fields.reported,
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
		context  context.Context
		client   *resty.Client
		reported repository.Storage
	}
	tests := []struct {
		name      string
		args      args
		assertion func(assert.TestingT, args, *defaultReporter)
	}{
		{
			name: "default initialisation",
			args: args{context: context.Background(), client: resty.New(), reported: repository.NewMemStorage()},
			assertion: func(t assert.TestingT, want args, got *defaultReporter) {
				assert.Equal(t, want.context, got.context)
				assert.Equal(t, want.client, got.client)
				assert.Equal(t, want.reported, got.reported)
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
		reported repository.Storage
		deadline time.Duration
	}
	type args struct {
		method    string
		urlRegexp *regexp.Regexp
		responder func(assert.TestingT) httpmock.Responder
		metric    model.Metric
	}
	type want struct {
		metric model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion func(assert.TestingT, error, model.Metric, repository.Storage)
	}{
		{
			name: "send counter without value",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				reported: repository.NewMemStorage(),
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
				metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			},
			want: want{
				metric: model.Metric{},
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.NotContains(t, calls, key)
				}
				assert.Errorf(t, err, "empty counter")
				got, err := reported.Get(want.Key())
				assert.ErrorIs(t, err, repository.ErrMetricNotFound)
				assert.Equal(t, want, got)
			},
		},
		{
			name: "send counter without value 2",
			fields: fields{
				client: resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				reported: func() repository.Storage {
					s := repository.NewMemStorage()
					err := s.Set(model.NewCounterMetric("id1", -5))
					assert.NoError(t, err)
					return s
				}(),
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
				metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			},
			want: want{
				metric: model.NewCounterMetric("id1", -5),
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				for _, num := range calls {
					assert.Zero(t, num)
				}
				assert.Errorf(t, err, "empty counter")
				got, err := reported.Get(want.Key())
				assert.NoError(t, err)
				assert.Equal(t, want, got)
			},
		},
		{
			name: "send counter",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				reported: repository.NewMemStorage(),
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
			want: want{
				metric: model.NewCounterMetric("id1", 5),
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				assert.NoError(t, err)
				got, err := reported.Get(want.Key())
				assert.NoError(t, err)
				assert.Equal(t, want, got)
			},
		},
		{
			name: "send counter 2",
			fields: fields{
				client: resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				reported: func() repository.Storage {
					s := repository.NewMemStorage()
					err := s.Set(model.NewCounterMetric("id1", -5))
					assert.NoError(t, err)
					return s
				}(),
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
			want: want{
				metric: model.NewCounterMetric("id1", 5),
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/10"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				assert.NoError(t, err)
				got, err := reported.Get(want.Key())
				assert.NoError(t, err)
				assert.Equal(t, want, got)
			},
		},
		{
			name: "send gauge",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				reported: repository.NewMemStorage(),
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
			want: want{
				metric: model.NewGaugeMetric("id1", -5.5),
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/gauge/id1/-5.5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				assert.NoError(t, err)
				got, err := reported.Get(want.Key())
				assert.NoError(t, err)
				assert.Equal(t, want, got)
			},
		},
		{
			name: "server error",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				reported: repository.NewMemStorage(),
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
			want: want{
				metric: model.Metric{},
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				assert.Error(t, err)
				got, err := reported.Get(want.Key())
				assert.ErrorIs(t, err, repository.ErrMetricNotFound)
				assert.Equal(t, want, got)
			},
		},
		{
			name: "server timeout",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(50 * time.Millisecond),
				reported: repository.NewMemStorage(),
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
			want: want{
				metric: model.Metric{},
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				assert.Errorf(t, err, "request cancelled")
				got, err := reported.Get(want.Key())
				assert.ErrorIs(t, err, repository.ErrMetricNotFound)
				assert.Equal(t, want, got)
			},
		},
		{
			name: "server deadline",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(200 * time.Millisecond),
				reported: repository.NewMemStorage(),
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
			want: want{
				metric: model.Metric{},
			},
			assertion: func(t assert.TestingT, err error, want model.Metric, reported repository.Storage) {
				calls := httpmock.GetCallCountInfo()
				keys := []string{"POST http://localhost:1234/update/counter/id1/5"}
				for _, key := range keys {
					assert.Contains(t, calls, key)
					assert.Equal(t, 1, calls[key])
				}
				assert.Errorf(t, err, "context deadline exceeded")
				got, err := reported.Get(want.Key())
				assert.ErrorIs(t, err, repository.ErrMetricNotFound)
				assert.Equal(t, want, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.fields.deadline)
			defer cancel()

			r := &defaultReporter{
				context:  ctx,
				client:   tt.fields.client,
				reported: tt.fields.reported,
			}
			httpmock.ActivateNonDefault(r.client.GetClient())
			defer httpmock.Reset()

			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, tt.args.responder(t))

			metric := tt.args.metric.Copy()
			tt.assertion(t, r.reportSingle(metric), tt.want.metric, r.reported)
			assert.Equal(t, tt.args.metric, metric)
		})
	}
}
