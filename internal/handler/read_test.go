package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_readHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics service.MetricStorer
	}
	type args struct {
		method      string
		url         string
		contentType httpheaders.ContentType
		body        io.Reader
	}
	type want struct {
		code        int
		body        string
		contentType httpheaders.ContentType
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion func(*testing.T, want, string)
	}{
		// Internal Error
		{
			name:   "GET %s INTERNAL_ERROR",
			fields: fields{metrics: &faultyMetricService{}},
			args:   args{method: http.MethodGet, url: "/", body: http.NoBody},
			want:   want{code: http.StatusInternalServerError, body: "", contentType: httpheaders.ContentTypeEmpty},
			assertion: func(t *testing.T, want want, body string) {
				assert.Equal(t, want.body, strings.TrimRight(body, "\n"))
			},
		},
		// OK
		{
			name:   "GET %s OK (no metrics)",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage())},
			args:   args{method: http.MethodGet, url: "/", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "", contentType: httpheaders.ContentTypeTextHTML},
			assertion: func(t *testing.T, want want, body string) {
				assert.Equal(t, want.body, strings.TrimRight(body, "\n"))

			},
		},
		{
			name:   "GET %s OK (single metric)",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage(model.NewCounterMetric("id1", 123)))},
			args:   args{method: http.MethodGet, url: "/", body: http.NoBody},
			want:   want{code: http.StatusOK, body: "id1 123", contentType: httpheaders.ContentTypeTextHTML},
			assertion: func(t *testing.T, want want, body string) {
				assert.Equal(t, want.body, strings.TrimRight(body, "\n"))
			},
		},
		{
			name: "GET %s OK (multiple metrics)",
			fields: fields{metrics: newMetricStorer(t, storagetest.NewMockStorage(
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id1", 1.23),
				model.NewGaugeMetric("id2", -1.23),
			))},
			args: args{method: http.MethodGet, url: "/", contentType: "text/plain", body: http.NoBody},
			want: want{code: http.StatusOK, body: "id1 123\nid2 -123\nid1 1.23\nid2 -1.23", contentType: httpheaders.ContentTypeTextHTML},
			assertion: func(t *testing.T, want want, body string) {
				content := strings.TrimRight(body, "\n")
				assert.ElementsMatch(t, strings.Split(want.body, "\n"), strings.Split(content, "\n"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf(tt.name, tt.args.url), func(t *testing.T) {
			logger := log.NewTestLogger()
			h := &readHandler{
				baseHandler: baseHandler{logger: logger},
				metrics:     tt.fields.metrics,
			}
			ts := httptest.NewServer(h)
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, tt.args.body)
			require.NoError(t, err)
			tt.args.contentType.Apply(req.Header)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.True(t, tt.want.contentType.Matches(resp.Header))
			tt.assertion(t, tt.want, string(body))
		})
	}
}
