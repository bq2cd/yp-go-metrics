package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/errhelper"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

var (
	errTestFinished = errors.New("test finished")
)

func TestRun(t *testing.T) {
	type args struct {
		timeout               time.Duration
		cfg                   config.Config
		overrideListenAddress bool
		method                string
		url                   string
		body                  []byte
	}
	type want struct {
		status  int
		wantErr bool
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"server is up and responds with OK": {
			args: args{
				timeout: 100 * time.Millisecond,
				cfg: config.Config{
					ShutdownTimeout: 1 * time.Millisecond,
				},
				overrideListenAddress: true,
				method:                http.MethodGet,
				url:                   "/",
			},
			want: want{
				status: http.StatusOK,
			},
		},
		"server fails to start due to invalid address": {
			args: args{
				timeout: 50 * time.Millisecond,
				cfg: config.Config{
					ListenAddress:   "12;34",
					ShutdownTimeout: 1 * time.Millisecond,
				},
				method: http.MethodGet,
				url:    "/",
			},
			want: want{
				wantErr: true,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			if tt.args.overrideListenAddress {
				tt.args.cfg.ListenAddress = servertest.GetRandomListenAddress(t)
			}

			req, err := http.NewRequest(tt.args.method, fmt.Sprintf("http://%s/%s", tt.args.cfg.ListenAddress, strings.TrimLeft(tt.args.url, "/")), bytes.NewReader(tt.args.body))
			require.NoError(t, err)

			ctx, cancel := context.WithTimeoutCause(t.Context(), tt.args.timeout, errTestFinished)
			defer cancel()

			logger := log.NewTestLogger()

			// Act
			errCh := make(chan error, 1)
			go func() {
				errCh <- Run(ctx, logger, tt.args.cfg)
			}()
			time.Sleep(5 * time.Millisecond)

			resp, errResp := http.DefaultClient.Do(req)

			// Assert
			var errFinal error
			for _, e := range errhelper.UnwrapJoined(<-errCh) {
				if errors.Is(e, errTestFinished) {
					continue
				}
				errFinal = errors.Join(errFinal, e)
			}
			assert.NotEmpty(t, logger.RecordedEvents())
			if tt.want.wantErr {
				require.Error(t, errFinal)
				return
			}
			require.NoError(t, errFinal)
			require.NoError(t, errResp)
			assert.Equal(t, tt.want.status, resp.StatusCode)
			_ = resp.Body.Close()
		})
	}
}
