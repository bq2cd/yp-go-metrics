package auditsink

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

var (
	// ErrHTTPSinkResponseError is returned by [WriteEvent] if remote HTTP server responds with non-2xx status code.
	ErrHTTPSinkResponseError = errors.New("HTTP sink responded with error")
)

// NewHTTPSink creates an instance of HTTP-based audit event sink.
func NewHTTPSink(remote url.URL) (*httpSink, error) {
	client := resty.New().SetBaseURL(remote.String())

	sink := &httpSink{
		client: client,
	}

	return sink, nil
}

type httpSink struct {
	client *resty.Client
}

// WriteEvent sends JSON-encoded audit event to a remote HTTP server.
func (s *httpSink) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	resp, err := s.client.R().SetContext(ctx).SetBody(event).Send()
	if err != nil {
		return fmt.Errorf("cannot send audit event to HTTP server: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("%w: %s", ErrHTTPSinkResponseError, resp.Status())
	}

	return nil
}

// Close does nothing and always returns `nil`.
func (s *httpSink) Close() error {
	return nil
}
