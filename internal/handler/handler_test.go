package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/contenttype"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBodyData struct {
	data        []byte
	contentType contenttype.ContentType
}

func newTestBodyDataFromMetric(t *testing.T, m model.Metric) testBodyData {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(m)
	require.NoError(t, err)
	return testBodyData{
		data:        buf.Bytes(),
		contentType: contenttype.ApplicationJSON,
	}
}

func newTestBodyDataFromMetricKey(t *testing.T, k model.MetricKey) testBodyData {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(k)
	require.NoError(t, err)
	return testBodyData{
		data:        buf.Bytes(),
		contentType: contenttype.ApplicationJSON,
	}
}

func (b *testBodyData) toRequest(method, url string) (*http.Request, error) {
	var body io.ReadCloser = http.NoBody
	if b.data != nil {
		body = io.NopCloser(bytes.NewReader(b.data))
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	b.contentType.ApplyToRequest(req)
	return req, nil
}

type faultyMetricJSONResponder struct{}

func (r *faultyMetricJSONResponder) WriteResponse(w http.ResponseWriter, m model.Metric) error {
	contenttype.ApplicationJSON.ApplyToResponse(w)
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
