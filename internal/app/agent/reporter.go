package agent

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

const (
	defaultSenderBatchSize = 10
)

// Reporter abstracts a process of sending metrics to a central storage.
type Reporter interface {
	Report(ctx context.Context, inCh <-chan model.Metric) error
}

type reporter struct {
	sender          SenderBatch
	reported        repository.StorageMulti
	senderPoolSize  uint
	senderBatchSize uint
}

// NewReporter creates an instance of the default reporter with
// specified internal storage.
func NewReporter(sender SenderBatch, storage repository.StorageMulti, numWorkers uint) *reporter {
	return &reporter{
		sender:          sender,
		reported:        storage,
		senderPoolSize:  numWorkers,
		senderBatchSize: defaultSenderBatchSize,
	}
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

func (r *reporter) reportBatch(ctx context.Context, orig model.MetricSet) error {
	sendable := r.getSendableMetrics(ctx, orig)
	if sendable.Empty() {
		return nil
	}

	sent, err := r.sender.SendBatch(ctx, sendable)

	return errors.Join(err,
		r.storeReported(ctx, orig, sent),
	)
}

func (r *reporter) reportWorker(ctx context.Context, inCh <-chan model.MetricSet) error {
	for batch := range inCh {
		err := r.reportBatch(ctx, batch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *reporter) metricBatcher(ctx context.Context, inCh <-chan model.Metric) <-chan model.MetricSet {
	outCh := make(chan model.MetricSet)

	go func() {
		defer close(outCh)
		batch := model.NewMetricSet()
		for metric := range inCh {
			if len(batch) < int(r.senderBatchSize) {
				// either we have room for a single metric or we have room for at least two.
				batch.Upsert(metric)
			}
			if len(batch) < int(r.senderBatchSize) {
				// there is still room for next metric, too early to send our batch.
				continue
			}
			// batch is full, sending it.
			select {
			case <-ctx.Done():
				return
			case outCh <- batch:
				batch = model.NewMetricSet()
			}
		}
		// send incomplete last batch
		if len(batch) > 0 {
			outCh <- batch
		}
	}()

	return outCh
}

func (r *reporter) processBatches(ctx context.Context, inCh <-chan model.MetricSet) error {
	poolSize := max(r.senderPoolSize, 1)

	erg := new(errgroup.Group)
	for range poolSize {
		erg.Go(func() error {
			return r.reportWorker(ctx, inCh)
		})
	}
	return erg.Wait()
}

// Report sends incoming metrics to an upstream.
func (r *reporter) Report(ctx context.Context, inCh <-chan model.Metric) error {
	outCh := r.metricBatcher(ctx, inCh)
	return r.processBatches(ctx, outCh)
}
