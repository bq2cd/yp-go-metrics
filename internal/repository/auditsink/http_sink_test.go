package auditsink

import (
	"context"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

func TestHTTPSink_WriteEvent(t *testing.T) {
	testEvent := model.AuditEvent{
		Timestamp:   time.Now().Unix(),
		MetricNames: []string{"metric1", "metric2"},
		IPAddress:   "127.0.0.1",
	}

	tests := map[string]struct {
		remoteURL neturl.URL
		timeout   time.Duration
		wantErr   bool
	}{
		"unreachable url": {
			remoteURL: neturl.URL{
				Scheme: "http",
				Host:   "localhost:0",
			},
			timeout: 100 * time.Millisecond,
			wantErr: true,
		},
		"slow upstream server": {
			remoteURL: func() neturl.URL {
				ts := httptest.NewServer(&mockHTTPSinkServer{delay: 100 * time.Millisecond})
				url, _ := neturl.Parse(ts.URL)

				return *url
			}(),
			timeout: 20 * time.Millisecond,
			wantErr: true,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			sink, err := NewHTTPSink(tc.remoteURL)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(t.Context(), tc.timeout)
			defer cancel()

			// Act
			err = sink.WriteEvent(ctx, testEvent)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

type mockHTTPSinkServer struct {
	delay time.Duration
}

func (m *mockHTTPSinkServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(m.delay)

	w.WriteHeader(http.StatusOK)
}
