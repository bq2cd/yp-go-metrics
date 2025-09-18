package agent

import (
	"context"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

// Reporter abstracts a way to send metrics to an upstream.
type Reporter interface {
	Report(metrics []model.Metric) error
}

type defaultReporter struct {
	context context.Context
	client  *resty.Client
}

// NewDefaultReporter creates an instance of the default reporter.
func NewDefaultReporter(ctx context.Context, client *resty.Client) *defaultReporter {
	return &defaultReporter{context: ctx, client: client}
}

// Report sends incoming metrics to an upstream.
func (r *defaultReporter) Report(metrics []model.Metric) error {
	return nil
}
