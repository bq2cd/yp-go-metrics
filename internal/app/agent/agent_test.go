package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/agent"
	"github.com/bq2cd/yp-go-metrics/internal/app/errhelper"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/handler/handlertest"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
				timeout: 50 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   10 * time.Millisecond,
					ReportInterval: 30 * time.Millisecond,
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
				timeout: 50 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   10 * time.Millisecond,
					ReportInterval: 30 * time.Millisecond,
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
				timeout: 50 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   10 * time.Millisecond,
					ReportInterval: 30 * time.Millisecond,
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mh := handlertest.NewMockHandler(ctrl)
			m := mh.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).
				MinTimes(1).
				Do(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.args.serverStatusCode)
				})
			if !tt.want.calledServer {
				m.Times(0)
			}

			ts := httptest.NewServer(mh)
			defer ts.Close()

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
				if e == agent.ErrSenderRequestFailed {
					continue
				}
				errFinal = errors.Join(errFinal, e)
			}
			if tt.want.wantErr {
				require.Error(t, errFinal)
			} else {
				require.NoError(t, errFinal)
			}
			assert.NotEmpty(t, logger.RecordedEvents())
		})
	}
}
