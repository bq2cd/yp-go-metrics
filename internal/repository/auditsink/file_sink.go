package auditsink

import (
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// NewFileSink creates an instance of file-based audit event sink.
func NewFileSink(path string) (*fileSink, error) {
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open audit file sink: %w", err)
	}

	sink := &fileSink{
		fp: fp,
	}

	return sink, nil
}

type fileSink struct {
	fp *os.File
}

func (s *fileSink) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	data, err := json.MarshalContext(ctx, event)
	if err != nil {
		return fmt.Errorf("cannot encode audit event to JSON: %w", err)
	}

	_, err = s.fp.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("cannot write encoded audit event to file: %w", err)
	}

	return nil
}

func (s *fileSink) Close() error {
	return s.fp.Close()
}
