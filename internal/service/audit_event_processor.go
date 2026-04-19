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

// NewAuditEventProcessor creates an instance of [AuditEventProcessor].
func NewAuditEventProcessor(logger log.Logger) *auditEventProcessor {
	return &auditEventProcessor{
		logger:   logger,
		sinks:    make(map[string]repository.AuditSink),
		incoming: make(chan model.AuditEvent, auditEventProcessorBufferSize),
	}
}

type auditEventProcessor struct {
	mu       sync.RWMutex
	logger   log.Logger
	sinks    map[string]repository.AuditSink
	incoming chan model.AuditEvent
	closed   bool
}

// WriteEvent accepts incoming audit event and puts into internal buffered channel.
// It will block if channel is full (e.g. not being processed fast enough).
// Provided context is used as a means for cancellation.
func (ap *auditEventProcessor) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if ap.closed {
		return ErrAuditEventProcessorClosed
	}

	ap.incoming <- event

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
		case event, ok := <-ap.incoming:
			if !ok {
				return
			}

			ap.processEvent(ctx, event)
		}
	}
}

func (ap *auditEventProcessor) closeSinks() {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.closed = true

	close(ap.incoming)

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
	grp.SetLimit(auditEventProcessorConcurrency)

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
