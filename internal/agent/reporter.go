package agent

import (
	"context"
	"errors"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

var (
	ErrReporterEmptyMetric = errors.New("empty metric")
)

// Reporter abstracts a process of sending metrics to a central storage.
type Reporter interface {
	Report(ctx context.Context, metrics []model.Metric) error
}

type reporter struct {
	sender   Sender
	reported repository.Storage
}

// NewReporter creates an instance of the default reporter with
// specified internal storage.
func NewReporter(sender Sender, storage repository.Storage) *reporter {
	return &reporter{sender: sender, reported: storage}
}

func (r *reporter) getSendableMetric(ctx context.Context, metric model.Metric) model.Metric {
	reported, err := r.reported.Get(ctx, metric.Key())
	if err != nil {
		// This includes cases where metric was not found or a storage error;
		// to avoid dropping metrics in case of storage error, we prefer to send
		// the full value instead.
		// This behaviour can be changed in the future if needed.
		return metric
	}

	switch reported.Type {
	case model.MetricTypeCounter:
		if metric.Delta == nil {
			// This looks like an invalid metric,
			// but we will leave it up to the sender to decide
			// what to do with it.
			return metric
		}
		if reported.Delta != nil {
			metric = metric.Copy()
			*metric.Delta -= *reported.Delta
		}
	}

	return metric
}

func (r *reporter) reportSingle(ctx context.Context, metric model.Metric) error {
	if metric.Empty() {
		return ErrReporterEmptyMetric
	}
	sendable := r.getSendableMetric(ctx, metric)

	err := r.sender.Send(sendable)
	if err != nil {
		return err
	}

	// There is a chance of discrepancy here if the underlying storage would
	// fail to store the metric.
	// If that happens, we would report full value instead of delta.
	// On the other hand, if we store metric in memory and restart the agent,
	// we will still report the full value on the first report.
	// This is something we would need to address at a later stage.
	return r.reported.Set(ctx, metric)
}

// Report sends incoming metrics to an upstream.
func (r *reporter) Report(ctx context.Context, metrics []model.Metric) error {
	var errFinal error
	for _, m := range metrics {
		errFinal = errors.Join(errFinal, r.reportSingle(ctx, m))
	}
	return errFinal
}
