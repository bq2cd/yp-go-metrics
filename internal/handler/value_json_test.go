package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_valueJSONHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics   service.MetricStorer
		responder metricJSONResponder
	}
	type args struct {
		bodyData       testBodyData
		shouldCompress bool
	}
	type want struct {
		code        int
		body        string
		contentType httpheaders.ContentType
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
				metrics:   service.NewMetricStorer(storagetest.NewMockStorage()),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "id1")),
			},
			want: want{
				code:        http.StatusNotFound,
				body:        `metric not found`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		{
			name: "empty metric key",
			fields: fields{
				metrics: service.NewMetricStorer(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "")),
			},
			want: want{
				code:        http.StatusBadRequest,
				body:        `empty metric type or id`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		{
			name: "invalid content-type",
			fields: fields{
				metrics: service.NewMetricStorer(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: func() testBodyData {
					bd := newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "id1"))
					bd.contentType = httpheaders.ContentTypeTextPlain
					return bd
				}(),
			},
			want: want{
				code:        http.StatusBadRequest,
				body:        `invalid content type`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		{
			name: "invalid json",
			fields: fields{
				metrics: service.NewMetricStorer(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: testBodyData{
					data:        []byte(`{ id: 1 }`),
					contentType: httpheaders.ContentTypeApplicationJSON,
				},
			},
			want: want{
				code:        http.StatusUnprocessableEntity,
				body:        `cannot decode metric`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		{
			name: "metric not found",
			fields: fields{
				metrics: service.NewMetricStorer(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "id3")),
			},
			want: want{
				code:        http.StatusNotFound,
				body:        `metric not found`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		{
			name: "faulty storage",
			fields: fields{
				metrics: service.NewMetricStorer(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				).MakeFaulty()),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, storagetest.FaultyStorageErrorTrigger)),
			},
			want: want{
				code:        http.StatusInternalServerError,
				body:        `cannot retrieve metric`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		{
			name: "metric found",
			fields: fields{
				metrics: service.NewMetricStorer(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeGauge, "id2")),
			},
			want: want{
				code:        http.StatusOK,
				body:        `{ "id": "id2", "type": "gauge", "value": -3.7 }`,
				contentType: httpheaders.ContentTypeApplicationJSON,
			},
		},
		{
			name: "json encoder error",
			fields: fields{
				metrics: service.NewMetricStorer(storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &faultyMetricJSONResponder{},
			},
			args: args{
				bodyData: newTestBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeGauge, "id2")),
			},
			want: want{
				code:        http.StatusOK,
				body:        ``,
				contentType: httpheaders.ContentTypeApplicationJSON,
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
			h := &valueJSONHandler{
				logger:    logger,
				metrics:   tt.fields.metrics,
				responder: tt.fields.responder,
			}
			ts := httptest.NewServer(h)
			defer ts.Close()

			req, err := tt.args.bodyData.toRequest(http.MethodPost, ts.URL+"/value", tt.args.shouldCompress)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.True(t, tt.want.contentType.Matches(resp.Header))
			if tt.want.contentType == httpheaders.ContentTypeApplicationJSON && !tt.want.invalidJSON {
				assert.JSONEq(t, tt.want.body, string(body))
			} else {
				assert.Equal(t, tt.want.body, strings.TrimRight(string(body), "\n"))
			}
			if tt.assertLogEvents != nil {
				tt.assertLogEvents(t, logger.RecordedEvents())
			}
		})
	}
}
