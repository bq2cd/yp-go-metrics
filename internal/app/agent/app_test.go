package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/errhelper"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

var (
	errTestFinished = errors.New("test finished")
)

type mockHandler struct {
	mu                 sync.Mutex
	numCalls           int
	responseStatusCode int
	lastRequestBody    *bytes.Buffer
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.numCalls++
	m.mu.Unlock()

	rgz, err := gzip.NewReader(r.Body)
	if err != nil {
		panic(err) // cannot use `testify` in HTTP handlers; better panic here instead.
	}

	defer rgz.Close()

	m.mu.Lock()
	m.lastRequestBody.Reset()
	m.lastRequestBody.ReadFrom(rgz)
	m.mu.Unlock()

	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(m.responseStatusCode)
	_ = json.NewEncoder(w).Encode([]model.Metric{})
}

func TestRun(t *testing.T) {
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
				timeout: 300 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   50 * time.Millisecond,
					ReportInterval: 50 * time.Millisecond,
				},
				overrideURL:      true,
				serverStatusCode: http.StatusOK,
			},
			want: want{
				calledServer: true,
			},
		},
		"agent collects metrics and reports encrypted metrics to server successfully": {
			args: args{
				timeout: 300 * time.Millisecond,
				cfg: config.Config{
					PollInterval:   50 * time.Millisecond,
					ReportInterval: 50 * time.Millisecond,
					ServerPublicKey: []byte(`
-----BEGIN PUBLIC KEY-----
MCowBQYDK2VuAyEAkZEdmg2VjtMlDU5mDWH76QagkM22DkDqVxt0W7NjqFM=
-----END PUBLIC KEY-----
						`),
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
				timeout: 300 * time.Millisecond,
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
				timeout: 300 * time.Millisecond,
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
			handler := &mockHandler{
				responseStatusCode: tt.args.serverStatusCode,
				lastRequestBody:    bytes.NewBuffer(nil),
			}

			ts := httptest.NewServer(handler)
			defer ts.Close()

			time.Sleep(20 * time.Millisecond)

			if tt.args.overrideURL {
				upstreamURL, err := url.Parse(ts.URL)
				require.NoError(t, err)
				tt.args.cfg.UpstreamURL = *upstreamURL
			}

			ctx, cancel := context.WithTimeoutCause(t.Context(), tt.args.timeout, errTestFinished)
			defer cancel()

			logger := log.NewTestLogger()

			// Act
			errRun := Run(ctx, logger, tt.args.cfg)

			// Assert
			errFinal := extractFinalError(errRun)
			if tt.want.wantErr {
				require.Error(t, errFinal)
			} else {
				require.NoError(t, errFinal)
			}
			if tt.want.calledServer {
				assert.GreaterOrEqual(t, handler.numCalls, 1)
				assert.NotEmpty(t, handler.lastRequestBody.Bytes())

				var lastRequestData any
				err := json.Unmarshal(handler.lastRequestBody.Bytes(), &lastRequestData)
				if len(tt.args.cfg.ServerPublicKey) == 0 {
					require.NoErrorf(t, err, "server should've received cleartext JSON data")
				} else {
					require.Errorf(t, err, "server should've received encrypted data (not cleartext JSON)")
				}
			} else {
				assert.Equalf(t, 0, handler.numCalls, "no HTTP requests to server should've happened")
			}
			assert.NotEmpty(t, logger.RecordedEvents())
		})
	}
}

func extractFinalError(errRun error) error {
	var errFinal error

	flattenedErrors := errhelper.UnwrapJoined(errRun)

	for i := 0; i < len(flattenedErrors); i++ {
		var curr, next error

		curr = flattenedErrors[i]
		if i < len(flattenedErrors)-1 {
			next = flattenedErrors[i+1]
		}

		if curr == ErrMetricCollectionFailed && errors.Is(next, context.DeadlineExceeded) {
			i++

			continue
		}

		if curr == ErrSenderRequestFailed && errors.Is(next, errTestFinished) {
			i++

			continue
		}

		// a bit hacky way to ignore canceled [retrymgr.Sleeper].
		if curr != nil && strings.HasPrefix(curr.Error(), "sleeper error ") {
			inner := errors.Unwrap(curr)
			if errors.Is(inner, context.DeadlineExceeded) {
				i++

				continue
			}
		}

		errFinal = errors.Join(errFinal, curr)
	}

	return errFinal
}
