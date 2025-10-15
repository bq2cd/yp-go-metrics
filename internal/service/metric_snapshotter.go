package service

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// MetricSnapshotter adds snapshotting functionality to a regular [MetricStorer] service.
// It adds two methods on top of it: `DumpClose` and `LoadClose`; both will close corresponding
// writer/reader after completing their activities.
// MetricSnapshotter also exposes a buffered `struct{}` channel with the capacity of 1
// for notifications on new writes (only the fact of the new writes, not the number of them).
type MetricSnapshotter interface {
	MetricStorer
	DumpClose(w io.WriteCloser) error
	LoadClose(r io.ReadCloser) error
	C() <-chan struct{}
}

type metricSnapshotter struct {
	MetricStorer
	mu          sync.RWMutex
	encoder     MetricEncoder
	decoder     MetricDecoder
	dirtyWrites atomic.Int64
	notifyCh    chan struct{}
}

// NewMetricSnapshotter creates an instance of metric snapshotter service
// which also implements [MetricStorer] interface.
// It also needs an instance of [MetricEncoder] and [MetricDecoder] to perform
// snapshots.
func NewMetricSnapshotter(storer MetricStorer, encoder MetricEncoder, decoder MetricDecoder) *metricSnapshotter {
	return &metricSnapshotter{
		MetricStorer: storer,
		encoder:      encoder,
		decoder:      decoder,
		notifyCh:     make(chan struct{}, 1),
	}
}

func (p *metricSnapshotter) markDirty(numWrites int) {
	p.dirtyWrites.Add(int64(numWrites))
	select {
	case p.notifyCh <- struct{}{}:
	default:
		// channel is full, do not block here
	}
}

// StoreSingle wraps corresponding [MetricStorer.StoreSingle] method
// to record a fact of a new write.
func (p *metricSnapshotter) StoreSingle(m model.Metric) error {
	p.mu.Lock()
	err := p.MetricStorer.StoreSingle(m)
	p.mu.Unlock()
	if err != nil {
		return err
	}
	p.markDirty(1)
	return nil
}

// StoreSingle wraps corresponding [MetricStorer.StoreBatch] method
// to record a fact of new writes.
func (p *metricSnapshotter) StoreBatch(metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	p.mu.Lock()
	err := p.MetricStorer.StoreBatch(metrics)
	p.mu.Unlock()
	if err != nil {
		return err
	}
	p.markDirty(len(metrics))
	return nil
}

// DumpClose creates a snapshot of all metrics in the underlying [MetricStorer] instance
// by encoding them with [MetricEncoder] and writing to provided [io.WriteCloser] writer.
// It closes the writer upon completion regardless of underlying errors.
func (p *metricSnapshotter) DumpClose(w io.WriteCloser) (errFinal error) {
	defer func() {
		errFinal = errors.Join(errFinal, w.Close())
	}()
	dirtyWrites := p.dirtyWrites.Load()
	if dirtyWrites == 0 {
		return
	}
	p.mu.RLock()
	metrics, err := p.RetrieveAll()
	p.mu.RUnlock()
	if err != nil {
		errFinal = fmt.Errorf("failed to retrieve all metrics: %w", err)
		return
	}
	err = p.encoder.EncodeBatch(w, metrics)
	if err != nil {
		errFinal = fmt.Errorf("failed to encode metrics: %w", err)
		return
	}
	p.dirtyWrites.CompareAndSwap(dirtyWrites, 0)
	return
}

// DumpClose reads encoded metrics from [io.ReadCloser] reader, decodes them with [MetricDecoder]
// and stores them into in the underlying [MetricStorer] instance.
// It closes the reader upon completion regardless of underlying errors.
func (p *metricSnapshotter) LoadClose(r io.ReadCloser) (errFinal error) {
	defer func() {
		errFinal = errors.Join(errFinal, r.Close())
	}()
	metrics, err := p.decoder.DecodeBatch(r)
	if err != nil {
		errFinal = fmt.Errorf("failed to decode metrics: %w", err)
		return
	}
	err = p.MetricStorer.StoreBatch(metrics)
	if err != nil {
		errFinal = fmt.Errorf("failed to store metrics: %w", err)
		return
	}
	return
}

// C returns a channel with capacity of 1 for notifications on writes.
func (p *metricSnapshotter) C() <-chan struct{} {
	return p.notifyCh
}
