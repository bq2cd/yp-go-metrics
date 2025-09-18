package agent

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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

func (r *defaultReporter) reportSingle(metric model.Metric) error {
	var mValue string
	switch metric.Type {
	case model.MetricTypeCounter:
		mValue = strconv.FormatInt(*metric.Delta, 10)
	default:
		mValue = strconv.FormatFloat(*metric.Value, 'g', 10, 64)
	}
	urlPath := fmt.Sprintf("/update/%s/%s/%s", string(metric.Type), metric.ID, mValue)

	req := r.client.R().SetContext(r.context)

	resp, err := req.SetHeader("content-type", "text/plain").Post(urlPath)
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("expected success, got status %v", resp.Status())
	}

	return nil
}

// Report sends incoming metrics to an upstream.
func (r *defaultReporter) Report(metrics []model.Metric) error {
	var errFinal error
	for _, m := range metrics {
		errFinal = errors.Join(errFinal, r.reportSingle(m))
	}
	return errFinal
}
