package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBodyData struct {
	data        []byte
	contentType contentType
}

func newTestBodyDataFromMetric(t *testing.T, m model.Metric) testBodyData {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(m)
	require.NoError(t, err)
	return testBodyData{
		data:        buf.Bytes(),
		contentType: contentTypeApplicationJSON,
	}
}

func newTestBodyDataFromMetricKey(t *testing.T, k model.MetricKey) testBodyData {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(k)
	require.NoError(t, err)
	return testBodyData{
		data:        buf.Bytes(),
		contentType: contentTypeApplicationJSON,
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
	b.contentType.applyToRequest(req)
	return req, nil
}

type faultyMetricJSONResponder struct{}

func (r *faultyMetricJSONResponder) WriteResponse(w http.ResponseWriter, m model.Metric) error {
	contentTypeApplicationJSON.applyToResponse(w)
	w.WriteHeader(http.StatusOK)
	var invalid chan struct{}
	return json.NewEncoder(w).Encode(invalid)
}

func dummyHTTPRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)
	return req
}

func Test_contentType_applyToRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name      string
		args      args
		c         contentType
		want      contentType
		assertion func(*testing.T, contentType, *http.Request)
	}{
		{
			name: "empty content type not applied",
			args: args{r: dummyHTTPRequest(t)},
			want: _contentTypeEmpty,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Empty(t, r.Header.Values(_contentTypeHeaderKey))
			},
		},
		{
			name: "empty content type does not override existing",
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(_contentTypeHeaderKey, "already/exists")
				return r
			}()},
			want: contentType("already/exists"),
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type applied",
			args: args{r: dummyHTTPRequest(t)},
			c:    contentTypeTextPlain,
			want: contentTypeTextPlain,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type overrides existing",
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(_contentTypeHeaderKey, "already/exists")
				return r
			}()},
			c:    contentTypeApplicationJSON,
			want: contentTypeApplicationJSON,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.applyToRequest(tt.args.r)
			tt.assertion(t, tt.want, tt.args.r)
		})
	}
}

func Test_contentType_applyToResponse(t *testing.T) {
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name      string
		args      args
		c         contentType
		want      contentType
		assertion func(*testing.T, contentType, *http.Request)
	}{
		{
			name: "empty content type not applied",
			args: args{w: httptest.NewRecorder()},
			want: _contentTypeEmpty,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Empty(t, r.Header.Values(_contentTypeHeaderKey))
			},
		},
		{
			name: "empty content type does not override existing",
			args: args{w: func() http.ResponseWriter {
				w := httptest.NewRecorder()
				w.Header().Set(_contentTypeHeaderKey, "already/exists")
				return w
			}()},
			want: contentType("already/exists"),
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type applied",
			args: args{w: httptest.NewRecorder()},
			c:    contentTypeTextPlain,
			want: contentTypeTextPlain,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
		{
			name: "new content type overrides existing",
			args: args{w: func() http.ResponseWriter {
				w := httptest.NewRecorder()
				w.Header().Set(_contentTypeHeaderKey, "already/exists")
				return w
			}()},
			c:    contentTypeApplicationJSON,
			want: contentTypeApplicationJSON,
			assertion: func(t *testing.T, want contentType, r *http.Request) {
				assert.Len(t, r.Header.Values(_contentTypeHeaderKey), 1)
				assert.Equal(t, want, contentType(r.Header.Get(_contentTypeHeaderKey)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.applyToResponse(tt.args.w)
		})
	}
}

func Test_contentType_matchesRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		c    contentType
		args args
		want bool
	}{
		{
			name: "no content-type",
			c:    contentTypeTextPlain,
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				return r
			}()},
			want: false,
		},
		{
			name: "different content-type",
			c:    contentTypeTextPlain,
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(_contentTypeHeaderKey, string(contentTypeApplicationJSON))
				return r
			}()},
			want: false,
		},
		{
			name: "same content-type",
			c:    contentTypeApplicationJSON,
			args: args{r: func() *http.Request {
				r := dummyHTTPRequest(t)
				r.Header.Set(_contentTypeHeaderKey, string(contentTypeApplicationJSON))
				return r
			}()},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.matchesRequest(tt.args.r))
		})
	}
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
