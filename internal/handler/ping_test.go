package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/handler/handlertest"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/service/servicetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_pingHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		pingerErr   error
		pingerDelay time.Duration
		timeout     time.Duration
	}
	type want struct {
		code        int
		body        string
		contentType httpheaders.ContentType
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		"database is reachable within timeout": {
			fields: fields{pingerDelay: 5 * time.Millisecond, pingerErr: nil, timeout: 20 * time.Millisecond},
			want: want{
				code:        http.StatusOK,
				body:        `OK`,
				contentType: httpheaders.ContentTypeTextPlain,
			},
		},
		"database is slow to respond": {
			fields: fields{pingerDelay: 50 * time.Millisecond, pingerErr: nil, timeout: 20 * time.Millisecond},
			want: want{
				code:        http.StatusInternalServerError,
				body:        `database timed out`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
		"database is outright unreachable": {
			fields: fields{pingerDelay: 2 * time.Millisecond, pingerErr: errors.New("connect failed"), timeout: 20 * time.Millisecond},
			want: want{
				code:        http.StatusInternalServerError,
				body:        `database unreachable`,
				contentType: httpheaders.ContentTypeTextPlain.UTF8(),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pinger := servicetest.NewMockStoragePinger(ctrl)
			pinger.EXPECT().Ping(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
				time.Sleep(tt.fields.pingerDelay)
				return tt.fields.pingerErr
			})

			logger := log.NewTestLogger()
			h := &pingHandler{
				baseHandler: baseHandler{logger: logger},
				pinger:      pinger,
				timeout:     tt.fields.timeout,
			}

			ts := httptest.NewServer(h)
			defer ts.Close()

			bodyData := handlertest.NewBodyData(t, nil)
			req := bodyData.NewRequest(http.MethodGet, ts.URL+"/ping", false)

			// Act
			resp, err := ts.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.NoError(t, err)

			// Assert
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.True(t, tt.want.contentType.Matches(resp.Header))
			assert.Equal(t, tt.want.body, strings.TrimRight(string(body), "\n"))
			if tt.fields.pingerErr != nil {
				assert.NotEmpty(t, logger.RecordedEvents())
			}

		})
	}
}
