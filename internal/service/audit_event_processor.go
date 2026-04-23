package service

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/option"
	"github.com/bq2cd/yp-go-metrics/pkg/sharedchan"
)

const (
	auditEventProcessorBufferSize  = 256
	auditEventProcessorConcurrency = 2
)

var (
	// ErrAuditEventProcessorClosed is returned by [WriteEvent] when audit processor is shutting down.
	ErrAuditEventProcessorClosed = errors.New("audit processor is closed")
)

// AuditEventProcessor defines a sink that is designed to run in background, processing incoming audit events
// and copying them to all registered sinks.
type AuditEventProcessor interface {
	repository.AuditSink

	RegisterSink(sinkID string, sink repository.AuditSink)
	StartProcessing(ctx context.Context)
}

// AuditEventProcessorConfig provides configuration knobs for [AuditEventProcessor].
// These knobs are settable via option functions (see `WithAuditEventProcessorXXX` functions).
type AuditEventProcessorConfig struct {
	bufferSize  uint
	concurrency int
}

// NewAuditEventProcessor creates an instance of [AuditEventProcessor].
func NewAuditEventProcessor(logger log.Logger, opts ...option.Option[AuditEventProcessorConfig]) *auditEventProcessor {
	cfg := AuditEventProcessorConfig{
		bufferSize:  auditEventProcessorBufferSize,
		concurrency: auditEventProcessorConcurrency,
	}

	option.Apply(&cfg, opts...)

	return &auditEventProcessor{
		logger:   logger,
		config:   cfg,
		sinks:    make(map[string]repository.AuditSink),
		incoming: sharedchan.NewChannel[model.AuditEvent](cfg.bufferSize),
	}
}

// WithAuditEventProcessorBufferSize sets desired size for buffered incoming channel where
// audit events are enqueued before processing.
// NB. Specifying size equal to zero will turn incoming channel into unbuffered one and
// [AuditEventProcessor.WriteEvent] will block on new event until this event is fully
// processed.
func WithAuditEventProcessorBufferSize(size uint) option.Option[AuditEventProcessorConfig] {
	return func(c *AuditEventProcessorConfig) {
		c.bufferSize = size
	}
}

// WithAuditEventProcessorConcurrency sets desired number of concurrently active sinks while
// processing an audit event. E.g. for `n = 2` and `5` sinks, each event will be sent
// simultaneously to `2` sinks, with other sinks being activated as soon as processing
// of any sink finishes.
// NB. Setting number to zero will lead to concurrent processing of all registered sinks for each event.
func WithAuditEventProcessorConcurrency(n uint) option.Option[AuditEventProcessorConfig] {
	return func(c *AuditEventProcessorConfig) {
		if n == 0 {
			c.concurrency = -1 // no limit on concurrent workers
		} else {
			c.concurrency = int(n)
		}
	}
}

type auditEventProcessor struct {
	mu       sync.RWMutex
	logger   log.Logger
	config   AuditEventProcessorConfig
	sinks    map[string]repository.AuditSink
	incoming *sharedchan.Channel[model.AuditEvent]
	closed   bool
}

// WriteEvent accepts incoming audit event and puts into internal buffered channel.
// It will block if channel is full (e.g. not being processed fast enough).
// Provided context is used as a means for cancellation.
func (ap *auditEventProcessor) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if !ap.incoming.Send(event) {
		return ErrAuditEventProcessorClosed
	}

	return nil
}

// Close is effectively a no-op: registered sinks are closed
// during graceful shutdown in [auditEventProcessor.StartProcessing] when context is canceled.
// This method is here to satisfy [repository.AuditSink] interface.
func (ap *auditEventProcessor) Close() error {
	return nil
}

// RegisterSink add provided [repository.AuditSink] into internal map under provided ID.
// It is a responsibility of the caller to ensure that ID is unique.
func (ap *auditEventProcessor) RegisterSink(sinkID string, sink repository.AuditSink) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.sinks[sinkID] = sink
}

// StartProcessing launches main loop that processes audit events from internal buffered channel.
// For each event, all registered sinks are invoked in parallel (with up to [auditEventProcessorConcurrency]),
// and errors are logged.
// If no sinks could accept the event, the event is skipped (that is, no retries, best-effort delivery).
// When provided context is canceled, the loop stops and all registered sinks are closed
// (which gives them a chance to flush their in-memory data, if any).
func (ap *auditEventProcessor) StartProcessing(ctx context.Context) {
	ap.mainLoop(ctx)
	ap.closeSinks()
}

func (ap *auditEventProcessor) mainLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ap.incoming.Receive():
			if !ok {
				return
			}

			ap.processEvent(ctx, event)
		}
	}
}

func (ap *auditEventProcessor) closeSinks() {
	ap.incoming.Close()

	ap.mu.Lock()
	defer ap.mu.Unlock()

	for sinkID, sink := range ap.sinks {
		err := sink.Close()
		if err != nil {
			ap.logger.Error().WithErr(err).Str("sink", sinkID).Msg("failed to close audit sink")
		} else {
			ap.logger.Info().Str("sink", sinkID).Msg("closed audit sink")
		}
	}
}

func (ap *auditEventProcessor) processEvent(ctx context.Context, event model.AuditEvent) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	grp := new(errgroup.Group)
	grp.SetLimit(ap.config.concurrency)

	for _, sink := range ap.getCurrentSinks() {
		grp.Go(func() error {
			return sink.WriteEvent(ctx, event)
		})
	}

	err := grp.Wait()

	if err != nil {
		ap.logger.Error().WithErr(err).Msg("some sinks failed to process audit event")
	}
}

func (ap *auditEventProcessor) getCurrentSinks() []repository.AuditSink {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	return slices.Collect(maps.Values(ap.sinks))
}
