package service

import (
	"context"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

// MetricBatch represents a list of metrics to be written to the storage
type MetricBatch []model.Metric

// MetricBatchTx represents a "transaction" on the pending-to-be-written [MetricBatch].
// It provides a method [Result] that returns error channel for the caller to subscribe to
// and get batch write result.
type MetricBatchTx interface {
	Result() <-chan error
}

// StorageBatchWriter provides a way to send batch updates to the storage (requires [repository.StorageMulti] capabilities).
// It is intended to run as a goroutine and process incoming
// batches via channel.
type StorageBatchWriter interface {
	WriteBatch(ctx context.Context, batch MetricBatch) MetricBatchTx
	StartProcessing(ctx context.Context)
}

type metricBatchTx struct {
	ctx   context.Context
	batch MetricBatch
	errCh chan error
}

// Result returns an error channel to subscribe to for notification on the batch processing.
func (b *metricBatchTx) Result() <-chan error {
	return b.errCh
}

type storageBatchWriter struct {
	storage    repository.StorageMulti
	incomingCh chan *metricBatchTx
}

// NewStorageBatchWriter creates an instance of [StorageBatchWriter] with provided [repository.StorageMulti] storage.
func NewStorageBatchWriter(storage repository.StorageMulti) *storageBatchWriter {
	return &storageBatchWriter{
		storage:    storage,
		incomingCh: make(chan *metricBatchTx, 256), // reduce blocking on writes
	}
}

// WriteBatch pushes given [MetricBatch] into internal channel and returns a tracking data
func (w *storageBatchWriter) WriteBatch(ctx context.Context, batch MetricBatch) MetricBatchTx {
	batchCtx := &metricBatchTx{ctx: ctx, batch: batch, errCh: make(chan error, 1)}
	switch len(batch) {
	case 0:
		// avoid queueing empty batches
		batchCtx.errCh <- nil
	default:
		w.incomingCh <- batchCtx
	}
	return batchCtx
}

// StartProcessing launches main loop of the writer to read batches from [incomingCh] channel and
// write them to the underlying storage.
func (w *storageBatchWriter) StartProcessing(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			close(w.incomingCh)
			break loop
		case batch, ok := <-w.incomingCh:
			if !ok {
				break loop
			}
			w.processBatchTx(batch)
		}
	}

	// process anything that might still be buffered in the channel
	for batch := range w.incomingCh {
		w.processBatchTx(batch)
	}
}

func (w *storageBatchWriter) processBatchTx(batchTx *metricBatchTx) {
	metrics := w.accumulateCounters(batchTx.ctx, batchTx.batch)
	batchTx.errCh <- w.storage.SetMulti(batchTx.ctx, metrics)
}

// accumulateCounters merges counter values with those already in the storage.
// It collects current counter values from the storage and updates
// provided metric set in place.
func (w *storageBatchWriter) accumulateCounters(ctx context.Context, batch MetricBatch) model.MetricSet {
	metrics := model.NewMetricSetWithStrategy(model.MetricAggregateStrategyCounterValueAccumulates, batch...)
	metricsByType := metrics.GroupByType()

	counters := metricsByType[model.MetricTypeCounter]
	if len(counters) == 0 {
		return metrics
	}

	existing, err := w.storage.GetMulti(ctx, counters.Keys())
	if err != nil {
		return metrics
	}

	for _, c := range existing {
		m, ok := metrics[c.Key()]
		if !ok {
			continue
		}
		m.AddDelta(c.Delta)
	}

	return metrics
}
