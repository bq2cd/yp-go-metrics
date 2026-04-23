package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service/servicetest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type faultyMetricJSONResponder struct{}

func (r *faultyMetricJSONResponder) WriteResponse(w http.ResponseWriter, m model.Metric) error {
	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(http.StatusOK)
	var invalid chan struct{}
	return json.NewEncoder(w).Encode(invalid)
}

type faultyMetricBatchJSONResponder struct{}

func (r *faultyMetricBatchJSONResponder) WriteResponse(w http.ResponseWriter, metrics []model.Metric) error {
	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(http.StatusOK)
	var invalid chan struct{}
	return json.NewEncoder(w).Encode(invalid)
}

func Test_defaultMetricJSONResponder_WriteResponse(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		m model.Metric
	}
	type want struct {
		data []byte
	}
	tests := []struct {
		name      string
		args      args
		want      want
		assertion func(*testing.T, want, []byte, error)
	}{
		{
			name: "empty metric",
			args: args{w: httptest.NewRecorder(), m: model.Metric{}},
			want: want{data: []byte(`{"id": "", "type": ""}`)},
			assertion: func(t *testing.T, want want, body []byte, err error) {
				require.NoError(t, err)
				assert.JSONEq(t, string(want.data), string(body))
			},
		},
		{
			name: "empty counter value",
			args: args{w: httptest.NewRecorder(), m: model.Metric{ID: "id1", Type: model.MetricTypeCounter}},
			want: want{data: []byte(`{"id": "id1", "type": "counter"}`)},
			assertion: func(t *testing.T, want want, body []byte, err error) {
				require.NoError(t, err)
				assert.JSONEq(t, string(want.data), string(body))
			},
		},
		{
			name: "some counter",
			args: args{w: httptest.NewRecorder(), m: model.NewCounterMetric("id1", -5)},
			want: want{data: []byte(`{"id": "id1", "type": "counter", "delta": -5}`)},
			assertion: func(t *testing.T, want want, body []byte, err error) {
				require.NoError(t, err)
				assert.JSONEq(t, string(want.data), string(body))
			},
		},
		{
			name: "some gauge",
			args: args{w: httptest.NewRecorder(), m: model.NewGaugeMetric("id1", -2.5)},
			want: want{data: []byte(`{"id": "id1", "type": "gauge", "value": -2.5}`)},
			assertion: func(t *testing.T, want want, body []byte, err error) {
				require.NoError(t, err)
				assert.JSONEq(t, string(want.data), string(body))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &defaultMetricJSONResponder{}
			errSend := r.WriteResponse(tt.args.w, tt.args.m)
			resp := tt.args.w.Result()
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			tt.assertion(t, tt.want, body, errSend)
		})
	}
}

func TestNewRegistry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("all handlers have logger", func(t *testing.T) {
		tests := map[string]struct {
			logger log.Logger
		}{
			"incoming logger is nil": {
				logger: nil,
			},
			"incoming logger is not nil": {
				logger: log.NewNoopLogger(),
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				got := NewRegistry(tt.logger, servicetest.NewMockMetricStorer(ctrl), servicetest.NewMockStoragePinger(ctrl), servicetest.NewMockMetricAuditor(ctrl))
				require.NotEmpty(t, got)
				for _, h := range got {
					logger := reflect.ValueOf(h).Elem().FieldByName("logger")
					assert.Truef(t, logger.IsValid(), "missing logger field")
					assert.Falsef(t, logger.IsNil(), "logger is nil")
				}
			})
		}
	})

	t.Run("logger contains handler name field", func(t *testing.T) {
		logger := log.NewTestLogger()

		got := NewRegistry(logger, servicetest.NewMockMetricStorer(ctrl), servicetest.NewMockStoragePinger(ctrl), servicetest.NewMockMetricAuditor(ctrl))
		require.NotEmpty(t, got)
		for _, h := range got {
			hl, ok := h.(handlerLogger)
			assert.Truef(t, ok, "must implement handlerLogger interface")
			hl.getLogger().Info().Send()
		}

		events := logger.RecordedEvents()
		for ident := range got {
			assert.NotEmpty(t, events.FindMatchingEvents(log.LevelInfo, "", log.Str("handler", string(ident))))
		}
	})
}

func Test_getHandlers(t *testing.T) {
	// covered by [TestNewRegistry]
	t.SkipNow()
}

func Test_baseHandler_setLogger(t *testing.T) {
	// covered by [TestNewRegistry]
	t.SkipNow()
}

func Test_baseHandler_getLogger(t *testing.T) {
	// covered by [TestNewRegistry]
	t.SkipNow()
}

func Test_defaultMetricBatchJSONResponder_WriteResponse(t *testing.T) {
	type args struct {
		w       *httptest.ResponseRecorder
		metrics []model.Metric
	}
	type want struct {
		data []byte
	}
	type testcase struct {
		r         metricBatchJSONResponder
		args      args
		want      want
		assertion func(testing.TB, want, []byte, error)
	}
	tests := map[string]testcase{
		"empty metrics": {
			r:    &defaultMetricBatchJSONResponder{},
			args: args{w: httptest.NewRecorder(), metrics: []model.Metric{}},
			want: want{data: []byte(`[]`)},
			assertion: func(t testing.TB, want want, body []byte, err error) {
				require.NoError(t, err)
				assert.JSONEq(t, string(want.data), string(body))
			},
		},
		"single metric": {
			r:    &defaultMetricBatchJSONResponder{},
			args: args{w: httptest.NewRecorder(), metrics: []model.Metric{model.NewCounterMetric("id1", -5)}},
			want: want{data: []byte(`[{"id": "id1", "type": "counter", "delta": -5}]`)},
			assertion: func(t testing.TB, want want, body []byte, err error) {
				require.NoError(t, err)
				assert.JSONEq(t, string(want.data), string(body))
			},
		},
		"multiple metrics": {
			r:    &defaultMetricBatchJSONResponder{},
			args: args{w: httptest.NewRecorder(), metrics: []model.Metric{model.NewCounterMetric("id1", -5), model.NewGaugeMetric("id2", -3.21)}},
			want: want{data: []byte(`[{"id": "id1", "type": "counter", "delta": -5}, {"id": "id2", "type": "gauge", "value": -3.21}]`)},
			assertion: func(t testing.TB, want want, body []byte, err error) {
				require.NoError(t, err)
				assert.JSONEq(t, string(want.data), string(body))
			},
		},
		"json encoder failure": {
			r:    &faultyMetricBatchJSONResponder{},
			args: args{w: httptest.NewRecorder(), metrics: []model.Metric{model.NewCounterMetric("id1", -5), model.NewGaugeMetric("id2", -3.21)}},
			want: want{data: []byte(``)},
			assertion: func(t testing.TB, want want, body []byte, err error) {
				require.Error(t, err)
				assert.Equal(t, string(want.data), string(body))
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			errSend := tt.r.WriteResponse(tt.args.w, tt.args.metrics)
			resp := tt.args.w.Result()
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			tt.assertion(t, tt.want, body, errSend)
		})
	}
}
