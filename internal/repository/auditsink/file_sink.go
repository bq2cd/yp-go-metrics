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
		jsonMarshaller: func(ctx context.Context, event model.AuditEvent) ([]byte, error) {
			return json.MarshalContext(ctx, event)
		},
	}

	return sink, nil
}

type fileSink struct {
	fp             *os.File
	jsonMarshaller func(context.Context, model.AuditEvent) ([]byte, error)
}

// WriteEvent writes given audit event to a file in JSONL format.
// Each audit event is JSON-encoded and written on a separate line.
func (s *fileSink) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	data, err := s.jsonMarshaller(ctx, event)
	if err != nil {
		return fmt.Errorf("cannot encode audit event to JSON: %w", err)
	}

	_, err = s.fp.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("cannot write encoded audit event to file: %w", err)
	}

	return nil
}

// Close closes underlying file, flushing all in-memory data to disk.
func (s *fileSink) Close() error {
	return s.fp.Close()
}
