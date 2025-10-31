package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_valueHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics service.MetricStorer
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
		name   string
		fields fields
		args   args
		want   want
	}{
		// Internal Error
		{
			name:   "GET %s INTERNAL_ERROR",
			fields: fields{metrics: &faultyMetricService{}},
			args:   args{method: http.MethodGet, url: "/value/counter/id1", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusInternalServerError, body: "cannot retrieve metric\n", contentType: "text/plain; charset=utf-8"},
		},
		// Not Found
		{
			name:   "GET %s NOT_FOUND",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodGet, url: "/value/badType", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "missing metric id\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "GET %s NOT_FOUND",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodGet, url: "/value/counter/id1", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "metric not found\n", contentType: "text/plain; charset=utf-8"},
		},
		// Bad Request
		{
			name:   "GET %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodGet, url: "/value/counter/id1/123", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "missing metric value\n", contentType: "text/plain; charset=utf-8"},
		},
		// OK
		{
			name:   "GET %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage(model.NewCounterMetric("id1", 123)))},
			args:   args{method: http.MethodGet, url: "/value/counter/id1", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "123", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "GET %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage(model.NewCounterMetric("id1", 123)))},
			args:   args{method: http.MethodGet, url: "/value/counter/id1/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "123", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "GET %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage(model.NewGaugeMetric("id1", -1.23)))},
			args:   args{method: http.MethodGet, url: "/value/gauge/id1", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "-1.23", contentType: "text/plain; charset=utf-8"},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf(tt.name, tt.args.url), func(t *testing.T) {
			h := &valueHandler{
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
			_ = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("content-type"))
			assert.Equal(t, tt.want.body, string(body))
		})
	}
}
