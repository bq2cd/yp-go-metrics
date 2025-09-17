package agent

import (
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

// Reporter abstracts a way to send metrics to an upstream.
type Reporter interface {
	Report(metrics []model.Metric) error
}

type defaultReporter struct {
	client *resty.Client
}

// Report sends incoming metrics to an upstream.
func (r *defaultReporter) Report(metrics []model.Metric) error {
	return nil
}
