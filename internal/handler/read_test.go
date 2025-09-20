package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_readHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics service.Metrics
	}
	type args struct {
		method      string
		url         string
		contentType string
		body        io.Reader
	}
	type want struct {
		code        int
		body        string
		contentType string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion func(assert.TestingT, want, string)
	}{
		// OK
		{
			name: "GET %s OK (no metrics)",
			fields: fields{metrics: service.NewMetrics(
				func() repository.Storage {
					s := repository.NewMemStorage()
					return s
				}(),
			)},
			args: args{method: http.MethodGet, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			name: "GET %s OK (single metric)",
			fields: fields{metrics: service.NewMetrics(
				func() repository.Storage {
					s := repository.NewMemStorage()
					err := s.Set(model.NewCounterMetric("id1", 123))
					require.NoError(t, err)
					return s
				}(),
			)},
			args: args{method: http.MethodGet, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusOK, body: "id1 123\n", contentType: "text/plain; charset=utf-8"},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.Equal(t, want.body, body)
			},
		},
		{
			name: "GET %s OK (multiple metrics)",
			fields: fields{metrics: service.NewMetrics(
				func() repository.Storage {
					s := repository.NewMemStorage()
					for _, m := range []model.Metric{
						model.NewCounterMetric("id1", 123),
						model.NewCounterMetric("id2", -123),
						model.NewGaugeMetric("id1", 1.23),
						model.NewGaugeMetric("id2", -1.23),
					} {
						require.NoError(t, s.Set(m))
					}
					return s
				}(),
			)},
			args: args{method: http.MethodGet, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusOK, body: "id1 123\nid2 -123\nid1 1.23\nid2 -1.23\n", contentType: "text/plain; charset=utf-8"},
			assertion: func(t assert.TestingT, want want, body string) {
				assert.ElementsMatch(t, strings.Split(want.body, "\n"), strings.Split(body, "\n"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf(tt.name, tt.args.url), func(t *testing.T) {
			h := &readHandler{
				metrics: tt.fields.metrics,
			}
			ts := httptest.NewServer(h)
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, tt.args.body)
			require.NoError(t, err)
			req.Header.Set("content-type", tt.args.contentType)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("content-type"))
			tt.assertion(t, tt.want, string(body))
		})
	}
}
