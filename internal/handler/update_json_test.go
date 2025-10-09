package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_updateJSONHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics   service.Metrics
		responder metricJSONResponder
	}
	type args struct {
		bodyData testBodyData
	}
	type want struct {
		code        int
		body        string
		contentType contentType
		invalidJSON bool
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		want            want
		assertLogEvents func(*testing.T, log.TestLogEventSet)
	}{
		{
			name: "empty storage",
			fields: fields{
				metrics:   service.NewMetrics(storagetest.NewMockStorage()),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.NewCounterMetric("id1", 7)),
			},
			want: want{
				code:        http.StatusOK,
				body:        `{ "id": "id1", "type": "counter", "delta": 7 }`,
				contentType: contentTypeApplicationJSON,
			},
		},
		{
			name: "add new counter",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.NewCounterMetric("id1", 7)),
			},
			want: want{
				code:        http.StatusOK,
				body:        `{ "id": "id1", "type": "counter", "delta": 7 }`,
				contentType: contentTypeApplicationJSON,
			},
		},
		{
			name: "update existing counter",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
					model.NewCounterMetric("id1", 10),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.NewCounterMetric("id1", 7)),
			},
			want: want{
				code:        http.StatusOK,
				body:        `{ "id": "id1", "type": "counter", "delta": 17 }`,
				contentType: contentTypeApplicationJSON,
			},
		},
		{
			name: "add new gauge",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id4", 77),
					model.NewGaugeMetric("id5", 8.8),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.NewGaugeMetric("id1", -7.8)),
			},
			want: want{
				code:        http.StatusOK,
				body:        `{ "id": "id1", "type": "gauge", "value": -7.8 }`,
				contentType: contentTypeApplicationJSON,
			},
		},
		{
			name: "update existing gauge",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id4", 77),
					model.NewGaugeMetric("id5", 8.8),
					model.NewGaugeMetric("id1", 0.8),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.NewGaugeMetric("id1", -7.8)),
			},
			want: want{
				code:        http.StatusOK,
				body:        `{ "id": "id1", "type": "gauge", "value": -7.8 }`,
				contentType: contentTypeApplicationJSON,
			},
		},
		{
			name: "skip empty metric",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.Metric{ID: "id1", Type: model.MetricTypeCounter}),
			},
			want: want{
				code:        http.StatusNotFound,
				body:        ``,
				contentType: contentTypeTextPlainUTF8,
			},
		},
		{
			name: "skip empty metric, return existing",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
					model.NewCounterMetric("id1", -33),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.Metric{ID: "id1", Type: model.MetricTypeCounter}),
			},
			want: want{
				code:        http.StatusOK,
				body:        `{ "id": "id1", "type": "counter", "delta": -33 }`,
				contentType: contentTypeApplicationJSON,
			},
		},
		{
			name: "invalid content-type",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: func() testBodyData {
					bd := newTestBodyDataFromMetric(t, model.NewCounterMetric("id1", 7))
					bd.contentType = contentTypeTextPlain
					return bd
				}(),
			},
			want: want{
				code:        http.StatusBadRequest,
				body:        ``,
				contentType: contentTypeTextPlainUTF8,
			},
		},
		{
			name: "invalid json",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: testBodyData{
					data:        []byte(`{ id: 1 }`),
					contentType: contentTypeApplicationJSON,
				},
			},
			want: want{
				code:        http.StatusUnprocessableEntity,
				body:        ``,
				contentType: contentTypeTextPlainUTF8,
			},
		},
		{
			name: "empty metric key",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "")),
			},
			want: want{
				code:        http.StatusBadRequest,
				body:        ``,
				contentType: contentTypeTextPlainUTF8,
			},
		},
		{
			name: "faulty storage",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id5", 88),
				).MakeFaulty()),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.NewCounterMetric(storagetest.FaultyStorageErrorTrigger, 7)),
			},
			want: want{
				code:        http.StatusInsufficientStorage,
				body:        ``,
				contentType: contentTypeTextPlainUTF8,
			},
		},
		{
			name: "json encoder error",
			fields: fields{
				metrics: service.NewMetrics(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &faultyMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetric(t, model.NewCounterMetric("id3", 7)),
			},
			want: want{
				code:        http.StatusOK,
				body:        ``,
				contentType: contentTypeApplicationJSON,
				invalidJSON: true,
			},
			assertLogEvents: func(t *testing.T, events log.TestLogEventSet) {
				require.Len(t, events, 1)
				e := events[0]
				assert.Equal(t, log.LevelError, e.Level())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewTestLogger()
			h := &updateJSONHandler{
				logger:    logger,
				metrics:   tt.fields.metrics,
				responder: tt.fields.responder,
			}
			ts := httptest.NewServer(h)
			defer ts.Close()

			req, err := tt.args.bodyData.toRequest(http.MethodPost, ts.URL+"/update")
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, contentType(resp.Header.Get(_contentTypeHeaderKey)))
			if tt.want.contentType == contentTypeApplicationJSON && !tt.want.invalidJSON {
				assert.JSONEq(t, tt.want.body, string(body))
			} else {
				assert.Equal(t, tt.want.body, strings.TrimRight(string(body), "\n"))
			}
		})
	}
}
