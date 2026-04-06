package service

import (
	"context"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

//go:generate go tool mockgen -destination=servicetest/mock_metric_auditor.go -package servicetest github.com/bq2cd/yp-go-metrics/internal/service MetricAuditor

// MetricAuditor allows for recording of client actions, e.g. uploading metrics.
type MetricAuditor interface {
	RegisterSink(sink repository.AuditSink)
	RecordMetricsUploaded(ctx context.Context, metrics model.MetricSet, clientInfo model.ClientInfo)
}

// NewMetricAuditor create an instance of [MetricAuditor].
func NewMetricAuditor() *metricAuditor {
	return &metricAuditor{
		sinks: make([]repository.AuditSink, 0),
	}
}

type metricAuditor struct {
	mu    sync.RWMutex
	sinks []repository.AuditSink
}

func (ma *metricAuditor) RegisterSink(sink repository.AuditSink) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	ma.sinks = append(ma.sinks, sink)
}

func (ma *metricAuditor) RecordMetricsUploaded(ctx context.Context, metrics model.MetricSet, clientInfo model.ClientInfo) {

}
