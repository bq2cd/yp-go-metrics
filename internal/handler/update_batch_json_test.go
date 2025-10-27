package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/handlertest"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_updateBatchJSONHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics   service.MetricStorer
		responder metricBatchJSONResponder
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
	type testcase struct {
		fields          fields
		args            args
		want            want
		assertLogEvents func(*testing.T, log.TestLogEventSet)
	}
	tests := map[string]testcase{
		"add counters to empty storage": {
			fields: fields{
				metrics:   newMetricStorer(t, storagetest.NewMockStorage()),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{
					model.NewCounterMetric("id1", 7),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -5),
				}),
			},
			want: want{
				code: http.StatusOK,
				body: `[
					{ "id": "id1", "type": "counter", "delta": 7 },
					{ "id": "id2", "type": "counter", "delta": 10 },
					{ "id": "id3", "type": "counter", "delta": -5 }
				]`,
				contentType: httpheaders.ContentTypeApplicationJSON,
			},
		},
		"add counters to non-empty storage with overlapping ids": {
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 10),
					model.NewCounterMetric("id2", -32),
					model.NewCounterMetric("id3", 11),
				)),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{
					model.NewCounterMetric("id1", 7),
					model.NewCounterMetric("id3", -5),
					model.NewCounterMetric("id5", 12),
					model.NewCounterMetric("id7", 8),
					model.NewCounterMetric("id5", -19),
				}),
			},
			want: want{
				code: http.StatusOK,
				body: `[
					{ "id": "id1", "type": "counter", "delta": 17 },
					{ "id": "id3", "type": "counter", "delta": 6 },
					{ "id": "id5", "type": "counter", "delta": -7 },
					{ "id": "id7", "type": "counter", "delta": 8 }
				]`,
				contentType: httpheaders.ContentTypeApplicationJSON,
			},
		},
		"add mixed metrics to non-empty storage with overlapping ids": {
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 10),
					model.NewCounterMetric("id2", -32),
					model.NewCounterMetric("id3", 11),
					model.NewGaugeMetric("id10", 8.3),
					model.NewGaugeMetric("id11", -5.6),
					model.NewGaugeMetric("id12", 0.032),
				)),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{
					model.NewCounterMetric("id1", -3),
					model.NewGaugeMetric("id10", -0.8),
					model.NewCounterMetric("id3", -5),
					model.NewGaugeMetric("id11", 7.11),
					model.NewCounterMetric("id5", 12),
					model.NewGaugeMetric("id10", 9.3),
					model.NewCounterMetric("id1", 5),
					model.NewGaugeMetric("id12", 99),
					model.NewCounterMetric("id7", 8),
					model.NewGaugeMetric("id13", -0.21),
					model.NewCounterMetric("id5", -19),
					model.NewGaugeMetric("id15", 8.345),
					model.NewCounterMetric("id1", 7),
				}),
			},
			want: want{
				code: http.StatusOK,
				body: `[
					{ "id": "id1", "type": "counter", "delta": 19 },
					{ "id": "id3", "type": "counter", "delta": 6 },
					{ "id": "id5", "type": "counter", "delta": -7 },
					{ "id": "id7", "type": "counter", "delta": 8 },
					{ "id": "id10", "type": "gauge", "value": 9.3 },
					{ "id": "id11", "type": "gauge", "value": 7.11 },
					{ "id": "id12", "type": "gauge", "value": 99 },
					{ "id": "id13", "type": "gauge", "value": -0.21 },
					{ "id": "id15", "type": "gauge", "value": 8.345 }
				]`,
				contentType: httpheaders.ContentTypeApplicationJSON,
			},
		},
		"empty request returns empty response": {
			fields: fields{
				metrics:   newMetricStorer(t, storagetest.NewMockStorage()),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{})},
			want: want{
				code:        http.StatusOK,
				body:        `[]`,
				contentType: httpheaders.ContentTypeApplicationJSON,
			},
		},
		"empty metrics are skipped": {
			fields: fields{
				metrics:   newMetricStorer(t, storagetest.NewMockStorage()),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{
					model.NewCounterMetric("id1", 7),
					{},
					model.NewCounterMetric("id2", 10),
					{Type: model.MetricTypeCounter},
					model.NewCounterMetric("id3", -5),
					{Type: model.MetricTypeCounter, ID: "id4"},
				}),
			},
			want: want{
				code: http.StatusOK,
				body: `[
					{ "id": "id1", "type": "counter", "delta": 7 },
					{ "id": "id2", "type": "counter", "delta": 10 },
					{ "id": "id3", "type": "counter", "delta": -5 }
				]`,
				contentType: httpheaders.ContentTypeApplicationJSON,
			},
		},
		"empty metrics are skipped but pre-existing are returned": {
			fields: fields{
				metrics: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id4", 35),
				)),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{
					model.NewCounterMetric("id1", 7),
					{},
					model.NewCounterMetric("id2", 10),
					{Type: model.MetricTypeCounter},
					model.NewCounterMetric("id3", -5),
					{Type: model.MetricTypeCounter, ID: "id4"},
				}),
			},
			want: want{
				code: http.StatusOK,
				body: `[
					{ "id": "id1", "type": "counter", "delta": 7 },
					{ "id": "id2", "type": "counter", "delta": 10 },
					{ "id": "id3", "type": "counter", "delta": -5 },
					{ "id": "id4", "type": "counter", "delta": 35 }
				]`,
				contentType: httpheaders.ContentTypeApplicationJSON,
			},
		},
		"invalid content type": {
			fields: fields{
				metrics:   newMetricStorer(t, storagetest.NewMockStorage()),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{
					model.NewCounterMetric("id1", 7),
					{},
					model.NewCounterMetric("id2", 10),
					{Type: model.MetricTypeCounter},
					model.NewCounterMetric("id3", -5),
					{Type: model.MetricTypeCounter, ID: "id4"},
				}).AsType(httpheaders.ContentTypeTextPlain),
			},
			want: want{
				code:        http.StatusBadRequest,
				body:        `invalid content type`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		"invalid json": {
			fields: fields{
				metrics:   newMetricStorer(t, storagetest.NewMockStorage()),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataOfType(t, []byte(`{"id": "id1", "type": "counter"}`), httpheaders.ContentTypeApplicationJSON),
			},
			want: want{
				code:        http.StatusUnprocessableEntity,
				body:        `cannot decode metrics`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		"faulty storage": {
			fields: fields{
				metrics:   newMetricStorer(t, storagetest.NewMockStorage().MakeFaulty()),
				responder: &defaultMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{model.NewGaugeMetric(storagetest.FaultyStorageErrorTrigger, 0.05)}),
			},
			want: want{
				code:        http.StatusInsufficientStorage,
				body:        `cannot store metrics`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		"json encoder error": {
			fields: fields{
				metrics:   newMetricStorer(t, storagetest.NewMockStorage()),
				responder: &faultyMetricBatchJSONResponder{},
			},
			args: args{
				bodyData: handlertest.NewBodyDataFromMetrics(t, []model.Metric{
					model.NewCounterMetric("id1", 7),
					{},
					model.NewCounterMetric("id2", 10),
					{Type: model.MetricTypeCounter},
					model.NewCounterMetric("id3", -5),
					{Type: model.MetricTypeCounter, ID: "id4"},
				}),
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			logger := log.NewTestLogger()
			h := &updateBatchJSONHandler{
				baseHandler: baseHandler{logger: logger},
				metrics:     tt.fields.metrics,
				responder:   tt.fields.responder,
			}
			ts := httptest.NewServer(h)
			defer ts.Close()

			req := tt.args.bodyData.NewRequest(http.MethodPost, ts.URL+"/updates", tt.args.shouldCompress)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.True(t, tt.want.contentType.Matches(resp.Header))
			if tt.want.contentType == httpheaders.ContentTypeApplicationJSON && !tt.want.invalidJSON {
				var err error
				var gotMetrics, wantMetrics []model.Metric
				err = json.Unmarshal(body, &gotMetrics)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(tt.want.body), &wantMetrics)
				require.NoError(t, err)
				assert.ElementsMatch(t, wantMetrics, gotMetrics)
			} else {
				assert.Equal(t, tt.want.body, strings.TrimRight(string(body), "\n"))
			}
			if tt.assertLogEvents != nil {
				tt.assertLogEvents(t, logger.RecordedEvents())
			}
		})
	}
}
