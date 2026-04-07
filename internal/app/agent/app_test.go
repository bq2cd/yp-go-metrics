package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/app/errhelper"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockHandler struct {
	mock.Mock
	numCalls   int
	statusCode int
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	m.numCalls++
	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(m.statusCode)
	_ = json.NewEncoder(w).Encode([]model.Metric{})

}

func TestRun(t *testing.T) {
	errTestFinished := errors.New("test finished")
	type args struct {
		timeout          time.Duration
		cfg              config.Config
		overrideURL      bool
		serverStatusCode int
	}
	type want struct {
		calledServer bool
		wantErr      bool
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"agent collects metrics and reports to server successfully": {
			args: args{
				timeout: 200 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   100 * time.Millisecond,
					ReportInterval: 100 * time.Millisecond,
				},
				overrideURL:      true,
				serverStatusCode: http.StatusOK,
			},
			want: want{
				calledServer: true,
			},
		},
		"agent collects metrics but server responds with error": {
			args: args{
				timeout: 100 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   50 * time.Millisecond,
					ReportInterval: 50 * time.Millisecond,
				},
				overrideURL:      true,
				serverStatusCode: http.StatusInternalServerError,
			},
			want: want{
				calledServer: true,
				wantErr:      true,
			},
		},
		"agent collects metrics but server unreachable": {
			args: args{
				timeout: 100 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   50 * time.Millisecond,
					ReportInterval: 50 * time.Millisecond,
					UpstreamURL:    url.URL{Host: "localhost"},
				},
			},
			want: want{
				calledServer: false,
				wantErr:      true,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			h := &mockHandler{statusCode: tt.args.serverStatusCode}
			if tt.want.calledServer {
				h.On("ServeHTTP", mock.Anything, mock.Anything).Return()
			}

			ts := httptest.NewServer(h)
			defer ts.Close()

			time.Sleep(50 * time.Millisecond)

			if tt.args.overrideURL {
				upstreamURL, err := url.Parse(ts.URL)
				require.NoError(t, err)
				tt.args.cfg.UpstreamURL = *upstreamURL
			}

			ctx, cancel := context.WithTimeoutCause(t.Context(), tt.args.timeout, errTestFinished)
			defer cancel()

			logger := log.NewTestLogger()

			// Act
			err := Run(ctx, logger, tt.args.cfg)

			// Assert
			var errFinal error
			for _, e := range errhelper.UnwrapJoined(err) {
				if errors.Is(e, errTestFinished) {
					continue
				}
				if e == ErrSenderRequestFailed {
					continue
				}
				errFinal = errors.Join(errFinal, e)
			}
			if tt.want.wantErr {
				require.Error(t, errFinal)
			} else {
				require.NoError(t, errFinal)
			}
			h.AssertExpectations(t)
			if tt.want.calledServer {
				assert.GreaterOrEqual(t, h.numCalls, 1)
			}
			assert.NotEmpty(t, logger.RecordedEvents())
		})
	}
}
