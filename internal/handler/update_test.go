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
	"github.com/bq2cd/yp-go-metrics/internal/service/servicetest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
		name             string
		fields           fields
		args             args
		want             want
		setupAuditorMock func(*servicetest.MockMetricAuditor)
	}{
		// Bad Request
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "", contentType: ""},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType/someID", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "", contentType: ""},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType/someID/1.23", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "", contentType: ""},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/someID/1.23", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "", contentType: ""},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/someID/123/bla", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "", contentType: ""},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter//456", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "", contentType: ""},
		},
		{
			name:   "POST %s BAD_REQUEST",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter//456/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusBadRequest, body: "", contentType: ""},
		},
		// Not Found
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "", contentType: ""},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/badType/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "", contentType: ""},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "", contentType: ""},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "", contentType: ""},
		},
		{
			name:   "POST %s NOT_FOUND",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/ /456", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusNotFound, body: "", contentType: ""},
		},
		// Internal error
		{
			name:   "POST %s INTERNAL_ERROR",
			fields: fields{metrics: &faultyMetricService{}},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/123", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusInternalServerError, body: "", contentType: ""},
		},
		// OK
		{
			name:   "POST %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/123", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			setupAuditorMock: func(m *servicetest.MockMetricAuditor) {
				m.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(model.NewCounterMetric("id1", 123)), gomock.Any())
			},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/123/", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			setupAuditorMock: func(m *servicetest.MockMetricAuditor) {
				m.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(model.NewCounterMetric("id1", 123)), gomock.Any())
			},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/counter/id1/-456", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			setupAuditorMock: func(m *servicetest.MockMetricAuditor) {
				m.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(model.NewCounterMetric("id1", -456)), gomock.Any())
			},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id2/1.05", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			setupAuditorMock: func(m *servicetest.MockMetricAuditor) {
				m.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(model.NewGaugeMetric("id2", 1.05)), gomock.Any())
			},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id2/-3.03", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			setupAuditorMock: func(m *servicetest.MockMetricAuditor) {
				m.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(model.NewGaugeMetric("id2", -3.03)), gomock.Any())
			},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id3/25", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			setupAuditorMock: func(m *servicetest.MockMetricAuditor) {
				m.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(model.NewGaugeMetric("id3", 25)), gomock.Any())
			},
		},
		{
			name:   "POST %s OK",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodPost, url: "/update/gauge/id3/-35", contentType: "text/plain", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: "text/plain; charset=utf-8"},
			setupAuditorMock: func(m *servicetest.MockMetricAuditor) {
				m.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(model.NewGaugeMetric("id3", -35)), gomock.Any())
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf(tt.name, tt.args.url), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			logger := log.NewTestLogger()
			auditorMock := servicetest.NewMockMetricAuditor(ctrl)
			if tt.setupAuditorMock != nil {
				tt.setupAuditorMock(auditorMock)
			}
			h := &updateHandler{
				baseHandler: baseHandler{logger: logger},
				metrics:     tt.fields.metrics,
				auditor:     auditorMock,
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
