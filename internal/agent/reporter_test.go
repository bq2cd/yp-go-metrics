package agent

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReporter struct {
	mock.Mock
	metrics []model.Metric
}

func (m *mockReporter) Report(metrics []model.Metric) error {
	m.Called(metrics)
	m.metrics = metrics
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
			name:   "send single counter",
			fields: fields{client: resty.New().SetBaseURL("http://localhost:1234")},
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
			name:   "send multiple counters",
			fields: fields{client: resty.New().SetBaseURL("http://localhost:1234")},
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
			name:   "send multiple metrics",
			fields: fields{client: resty.New().SetBaseURL("http://localhost:1234")},
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
			name:   "send error",
			fields: fields{client: resty.New().SetBaseURL("http://localhost:1234")},
			args: args{
				method:    http.MethodPost,
				urlRegexp: regexp.MustCompile("^http://localhost:1234/update/([^/]+)/([^/]+)/([^/]+)/?$"),
				responder: func(t assert.TestingT) httpmock.Responder {
					return func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, "text/plain", r.Header.Get("content-type"))
						return httpmock.NewStringResponse(http.StatusBadGateway, ""), nil
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
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &defaultReporter{
				client: tt.fields.client,
			}
			httpmock.ActivateNonDefault(r.client.GetClient())
			defer httpmock.Reset()

			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, tt.args.responder(t))

			tt.assertion(t, r.Report(tt.args.metrics))

		})
	}
}
