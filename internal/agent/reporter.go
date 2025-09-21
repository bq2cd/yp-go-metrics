package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/go-resty/resty/v2"
)

// Reporter abstracts a way to send metrics to an upstream.
type Reporter interface {
	Report(metrics []model.Metric) error
}

type defaultReporter struct {
	context  context.Context
	client   *resty.Client
	reported repository.Storage
}

// NewDefaultReporter creates an instance of the default reporter
// with in-memory internal storage.
func NewDefaultReporter(ctx context.Context, client *resty.Client) *defaultReporter {
	return &defaultReporter{context: ctx, client: client, reported: repository.NewMemStorage()}
}

// NewReporter creates an instance of the default reporter with
// specified internal storage.
func NewReporter(ctx context.Context, client *resty.Client, storage repository.Storage) *defaultReporter {
	return &defaultReporter{context: ctx, client: client, reported: storage}
}

func (r *defaultReporter) reportSingle(metric model.Metric) error {
	current := metric.Copy()

	if reported, err := r.reported.Get(current.Key()); err == nil {
		switch current.Type {
		case model.MetricTypeCounter:
			if current.Delta == nil {
				return fmt.Errorf("empty counter")
			}
			if reported.Delta != nil {
				*current.Delta -= *reported.Delta
			}
		}
	}

	metricOp := urlpath.NewOperationFromMetric(urlpath.OperationTypeUpdate, current)
	urlPath, err := metricOp.ToURLPath()
	if err != nil {
		return fmt.Errorf("cannot convert metric to url path: %w", err)
	}

	req := r.client.R().SetContext(r.context)

	resp, err := req.SetHeader("content-type", "text/plain").Post(urlPath)
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("expected success, got status %v", resp.Status())
	}

	// There is a chance of discrepancy here if the underlying storage would
	// fail to store the metric.
	// If that happens, we would report full value instead of delta.
	// On the other hand, if we store metric in memory and restart the agent,
	// we will still report the full value on the first report.
	// This is something we would need to address at a later stage.
	return r.reported.Set(metric)
}

// Report sends incoming metrics to an upstream.
func (r *defaultReporter) Report(metrics []model.Metric) error {
	var errFinal error
	for _, m := range metrics {
		errFinal = errors.Join(errFinal, r.reportSingle(m))
	}
	return errFinal
}
