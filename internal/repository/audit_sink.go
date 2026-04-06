package repository

import (
	"context"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// AuditSink represents a destination for audit events.
type AuditSink interface {
	WriteEvent(ctx context.Context, event model.AuditEvent) error
}
