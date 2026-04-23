package repository

import (
	"context"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

//go:generate go tool mockgen -destination=auditsink/auditsinktest/mock_audit_sink.go -package auditsinktest github.com/bq2cd/yp-go-metrics/internal/repository AuditSink

// AuditSink represents a destination for audit events.
type AuditSink interface {
	WriteEvent(ctx context.Context, event model.AuditEvent) error
	Close() error
}
