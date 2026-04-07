package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"golang.org/x/sync/errgroup"
)

var (
	ErrMetricCollectionFailed = errors.New("metric collection failed")
)

// Collector abstracts a way to collect metrics.
type Collector interface {
	Collect(ctx context.Context) error
	Snapshot(ctx context.Context) (<-chan model.Metric, error)
}

type collector struct {
	sources   []source.Source
	collected repository.Storage
}

// NewCollector creates an instance of the default collector with specific
// metric sources and internal storage.
func NewCollector(sources []source.Source, storage repository.Storage) *collector {
	return &collector{sources: sources, collected: storage}
}

func (c *collector) collectFromSource(ctx context.Context, src source.Source, outCh chan<- model.Metric) error {
	defer close(outCh)
	metrics, err := src.ReadMetrics()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		outCh <- metric
	}
	return nil
}

func (c *collector) fanInMetrics(ctx context.Context, inCh <-chan model.Metric, outCh chan<- model.Metric) error {
	for metric := range inCh {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		outCh <- metric
	}
	return nil
}

func (c *collector) storeMetrics(ctx context.Context, inChs []<-chan model.Metric) error {
	outCh := make(chan model.Metric)

	// fan-in metrics into a single channel from multiple goroutines.
	erg := new(errgroup.Group)
	for _, inCh := range inChs {
		erg.Go(func() error {
			return c.fanInMetrics(ctx, inCh, outCh)
		})
	}
	go func() {
		erg.Wait() // we do not care for context errors here
		close(outCh)
	}()

	// write metrics to the storage.
	var errFinal error
	for metric := range outCh {
		err := c.collected.Set(ctx, metric)
		if err != nil {
			errFinal = errors.Join(errFinal, err)
			// do not return straight away to allow other metrics
			// to be stored (in case of transient errors).
			continue
		}
	}
	return errFinal
}

// Collect queries metric sources and stores obtained metrics
// into the internal storage.
func (c *collector) Collect(ctx context.Context) error {
	outChs := make([]<-chan model.Metric, 0, len(c.sources))
	erg := new(errgroup.Group)

	for _, src := range c.sources {
		outCh := make(chan model.Metric)
		outChs = append(outChs, outCh)
		erg.Go(func() error {
			return c.collectFromSource(ctx, src, outCh)
		})
	}

	erg.Go(func() error {
		return c.storeMetrics(ctx, outChs)
	})

	err := erg.Wait()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMetricCollectionFailed, err)
	}

	return nil
}

// Snapshot returns latest values of all metrics from the internal storage.
func (c *collector) Snapshot(ctx context.Context) (<-chan model.Metric, error) {
	// Strictly speaking, it does not make much sense to implement a generator pattern here,
	// but we will do this nonetheless for learning purposes.
	outCh := make(chan model.Metric)

	metrics, err := c.collected.GetAll(ctx)
	if err != nil {
		close(outCh)
		return outCh, err
	}

	go func() {
		defer close(outCh)
		for _, metric := range metrics {
			select {
			case <-ctx.Done():
				return
			case outCh <- metric:
				// business as usual
			}
		}
	}()

	return outCh, nil
}
