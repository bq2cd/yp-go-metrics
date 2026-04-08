package service

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"golang.org/x/sync/errgroup"
)

const (
	auditEventProcessorBufferSize  = 256
	auditEventProcessorConcurrency = 2
)

var (
	ErrAuditEventProcessorClosed = errors.New("audit processor is closed")
)

// AuditEventProcessor defines a sink that is designed to run in background, processing incoming audit events
// and copying them to all registered sinks.
type AuditEventProcessor interface {
	repository.AuditSink

	RegisterSink(sinkID string, sink repository.AuditSink)
	StartProcessing(ctx context.Context)
}

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

func (ap *auditEventProcessor) RegisterSink(sinkID string, sink repository.AuditSink) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.sinks[sinkID] = sink
}

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
