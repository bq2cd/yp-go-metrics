package service

import (
	"context"
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
}

func (ap *auditEventProcessor) RegisterSink(sinkID string, sink repository.AuditSink) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.sinks[sinkID] = sink
}

func (ap *auditEventProcessor) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ap.incoming <- event:
		return nil
	}
}

func (ap *auditEventProcessor) StartProcessing(ctx context.Context) {
	ap.mainLoop(ctx)
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
