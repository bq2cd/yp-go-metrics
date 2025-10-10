package agent

import (
	"context"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/handler/contenttype"
	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSender struct {
	mock.Mock
	wantErr func(model.Metric) error
}

func (m *mockSender) Send(metric model.Metric) error {
	m.Called(metric)
	if m.wantErr != nil {
		return m.wantErr(metric)
	}
	return nil
}

func Test_senderPlain_Send(t *testing.T) {
	type fields struct {
		client   *resty.Client
		deadline time.Duration
	}
	type responder struct {
		contentType contenttype.ContentType
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
				contentType: contenttype.TextPlain,
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
				contentType: contenttype.TextPlain,
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
				contentType: contenttype.TextPlain,
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
				contentType: contenttype.TextPlain,
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
				contentType: contenttype.TextPlain,
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
				contentType: contenttype.TextPlain,
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
			name: "server deadline",
			fields: fields{
				client:   resty.New().SetBaseURL("http://localhost:1234").SetTimeout(200 * time.Millisecond),
				deadline: 100 * time.Millisecond,
			},
			responder: responder{
				contentType: contenttype.TextPlain,
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
				context: ctx,
				client:  tt.fields.client,
			}
			httpmock.ActivateNonDefault(s.client.GetClient())
			defer httpmock.Reset()
			httpmock.RegisterRegexpResponder(tt.args.method, tt.args.urlRegexp, func(r *http.Request) (*http.Response, error) {
				require.True(t, tt.responder.contentType.MatchesRequest(r))
				time.Sleep(tt.responder.timeout)
				return httpmock.NewStringResponse(tt.responder.status, ""), nil
			})

			metric := tt.args.metric.Copy()

			err := s.Send(metric)

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
		ctx    context.Context
		client *resty.Client
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, *senderPlain)
	}{
		{
			name: "default",
			args: args{ctx: context.Background(), client: resty.New()},
			assertion: func(t *testing.T, args args, got *senderPlain) {
				assert.Equal(t, args.ctx, got.context)
				assert.Equal(t, args.client, got.client)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewSenderPlain(tt.args.ctx, tt.args.client))
		})
	}
}
