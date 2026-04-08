package auditsink

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

func TestNewFileSink(t *testing.T) {
	tests := map[string]struct {
		filePath string
		wantErr  bool
	}{
		"not a file": {
			filePath: "/tmp",
			wantErr:  true,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			_, err := NewFileSink(tc.filePath)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFileSink_WriteEvent(t *testing.T) {
	tempFactory := servertest.NewTempFileFactory(t)
	defer tempFactory.RemoveAll()

	testEvent := model.AuditEvent{
		Timestamp:   time.Now().Unix(),
		MetricNames: []string{"metric1", "metric2"},
		IPAddress:   "127.0.0.1",
	}

	tests := map[string]struct {
		setupFileSink func(*fileSink)
		wantErr       bool
	}{
		"write error": {
			setupFileSink: func(fs *fileSink) {
				fs.fp, _ = os.Open(tempFactory.Create("read-only-*"))
			},
			wantErr: true,
		},
		"json encoding error": {
			setupFileSink: func(fs *fileSink) {
				fs.jsonMarshaller = func(_ context.Context, _ model.AuditEvent) ([]byte, error) {
					return nil, errors.New("unlikely json failure")
				}
			},
			wantErr: true,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			sink, err := NewFileSink(tempFactory.Create("file-sink-*"))
			require.NoError(t, err)

			tc.setupFileSink(sink)

			// Act
			err = sink.WriteEvent(t.Context(), testEvent)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
