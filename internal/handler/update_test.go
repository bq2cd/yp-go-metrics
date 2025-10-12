package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_updateHandler_ServeHTTP(t *testing.T) {
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
		// Bad Request
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType/someID", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType/someID/1.23", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/someID/1.23", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/someID/123/bla", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter//456", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter//456/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		// Not Found
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/ /456", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "\n", contentType: "text/plain; charset=utf-8"},
		},
		// Internal error
		{
			name:   "POST %s INTERNAL_ERROR",
			fields: fields{metrics: &faultyMetricService{}},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/123", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusInternalServerError, body: "failed to update metric\n", contentType: "text/plain; charset=utf-8"},
		},
		// OK
		{
			name:   "POST %s OK",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/123", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/123/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/-456", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id2/1.05", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id2/-3.03", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id3/25", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: service.NewMetricStorer(repository.NewMemStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id3/-35", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf(tt.name, tt.args.url), func(t *testing.T) {
			h := &updateHandler{
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
