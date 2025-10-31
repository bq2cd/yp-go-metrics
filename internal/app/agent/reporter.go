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
	sender   SenderBatch
	reported repository.StorageMulti
}

// NewReporter creates an instance of the default reporter with
// specified internal storage.
func NewReporter(sender SenderBatch, storage repository.StorageMulti) *reporter {
	return &reporter{sender: sender, reported: storage}
}

func (r *reporter) getSendableMetric(ctx context.Context, metric model.Metric) model.Metric {
	reported, err := r.reported.Get(ctx, metric.Key())
	if err != nil {
		// This includes cases where metric was not found or a storage error;
		// to avoid dropping metrics in case of storage error, we prefer to send
		// the full value instead.
		// This behavior can be changed in the future if needed.
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

func (r *reporter) getSendableMetrics(ctx context.Context, orig model.MetricSet) model.MetricSet {
	sendable := model.NewMetricSet()

	for _, m := range orig {
		sendable.Upsert(r.getSendableMetric(ctx, m))
	}

	return sendable
}

func (r *reporter) storeReported(ctx context.Context, orig, sent model.MetricSet) error {
	reported := model.NewMetricSet()
	for _, s := range sent {
		if m, ok := orig[s.Key()]; ok {
			reported.Upsert(m)
		}
	}
	return r.reported.SetMulti(ctx, reported)
}

func (r *reporter) reportSingle(ctx context.Context, metric model.Metric) error {
	if metric.Empty() {
		return ErrReporterEmptyMetric
	}
	return r.reportBatch(ctx, []model.Metric{metric})
}

func (r *reporter) reportBatch(ctx context.Context, metrics []model.Metric) error {
	orig := model.NewMetricSet(metrics...)

	sendable := r.getSendableMetrics(ctx, orig)
	if sendable.Empty() {
		return nil
	}

	sent, err := r.sender.SendBatch(ctx, sendable)

	return errors.Join(err,
		r.storeReported(ctx, orig, sent),
	)
}

// Report sends incoming metrics to an upstream.
func (r *reporter) Report(ctx context.Context, metrics []model.Metric) error {
	return r.reportBatch(ctx, metrics)
}
