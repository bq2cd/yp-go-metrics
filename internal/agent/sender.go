package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

var (
	ErrSenderRequestFailed = errors.New("http request error")
	ErrSenderResponseNotOK = errors.New("http response not OK")
)

// Sender abstracts a protocol to encode and send a single metric.
type Sender interface {
	Send(metric model.Metric) error
}

// NewSenderPlain creates an instance of a sender that reports metrics in plain text.
func NewSenderPlain(ctx context.Context, client *resty.Client) *senderPlain {
	return &senderPlain{context: ctx, client: client}
}

type senderPlain struct {
	context context.Context
	client  *resty.Client
}

func (s *senderPlain) Send(metric model.Metric) error {
	metricOp := urlpath.NewOperationFromMetric(urlpath.OperationTypeUpdate, metric)
	urlPath, err := metricOp.ToURLPath()
	if err != nil {
		return fmt.Errorf("cannot convert metric to url path: %w", err)
	}

	req := s.client.R().SetContext(s.context)

	resp, err := req.SetHeader("content-type", "text/plain").Post(urlPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSenderRequestFailed, err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("expected success, got status %v: %w", resp.Status(), ErrSenderResponseNotOK)
	}

	return nil
}
