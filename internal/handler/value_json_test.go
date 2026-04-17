package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/handler/handlertest"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func Test_valueJSONHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics   service.MetricStorer
		responder metricJSONResponder
	}
	type args struct {
		bodyData       handlertest.BodyData
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
				metrics:   newMetricStorer(t, storagetest.NewMockStorage()),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "id1")),
			},
			want: want{
				code:        http.StatusNotFound,
				body:        ``,
				contentType: httpheaders.ContentTypeEmpty,
			},
		},
		{
			name: "empty metric key",
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "")),
			},
			want: want{
				code:        http.StatusBadRequest,
				body:        ``,
				contentType: httpheaders.ContentTypeEmpty,
			},
		},
		{
			name: "invalid content-type",
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "id1")).AsType(httpheaders.ContentTypeTextPlain),
			},
			want: want{
				code:        http.StatusBadRequest,
				body:        ``,
				contentType: httpheaders.ContentTypeEmpty,
			},
		},
		{
			name: "invalid json",
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataOfType(t, []byte(`{ id: 1 }`), httpheaders.ContentTypeApplicationJSON),
			},
			want: want{
				code:        http.StatusUnprocessableEntity,
				body:        ``,
				contentType: httpheaders.ContentTypeEmpty,
			},
		},
		{
			name: "metric not found",
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "id3")),
			},
			want: want{
				code:        http.StatusNotFound,
				body:        ``,
				contentType: httpheaders.ContentTypeEmpty,
			},
		},
		{
			name: "faulty storage",
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				).MakeFaulty()),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, storagetest.FaultyStorageErrorTrigger)),
			},
			want: want{
				code:        http.StatusInternalServerError,
				body:        ``,
				contentType: httpheaders.ContentTypeEmpty,
			},
		},
		{
			name: "metric found",
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &defaultMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeGauge, "id2")),
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
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 12),
					model.NewGaugeMetric("id2", -3.7),
				)),
				responder: &faultyMetricJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeGauge, "id2")),
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
				baseHandler: baseHandler{logger: logger},
				metrics:     tt.fields.metrics,
				responder:   tt.fields.responder,
			}
			ts := httptest.NewServer(h)
			defer ts.Close()

			req := tt.args.bodyData.NewRequest(http.MethodPost, ts.URL+"/value", tt.args.shouldCompress)

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
