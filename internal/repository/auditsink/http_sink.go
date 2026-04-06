package auditsink

import (
	"context"
	"net/url"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
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

func (s *httpSink) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	return nil
}
