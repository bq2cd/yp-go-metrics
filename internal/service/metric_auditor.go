package service

import (
	"context"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

//go:generate go tool mockgen -destination=servicetest/mock_metric_auditor.go -package servicetest github.com/bq2cd/yp-go-metrics/internal/service MetricAuditor

// MetricAuditor allows for recording of client actions, e.g. uploading metrics.
type MetricAuditor interface {
	RecordMetricsUploaded(ctx context.Context, metrics model.MetricSet, clientInfo model.ClientInfo)
}

// NewMetricAuditor create an instance of [MetricAuditor].
func NewMetricAuditor(logger log.Logger, sink repository.AuditSink) *metricAuditor {
	return &metricAuditor{
		logger: logger,
		sink:   sink,
	}
}

type metricAuditor struct {
	logger log.Logger
	sink   repository.AuditSink
}

func (ma *metricAuditor) RecordMetricsUploaded(ctx context.Context, metrics model.MetricSet, clientInfo model.ClientInfo) {
	event := model.NewAuditEvent(metrics, clientInfo.IPAddress)

	err := ma.sink.WriteEvent(ctx, event)
	if err != nil {
		ma.logger.Error().WithErr(err).Any("event", event).Msg("cannot write audit event")
	}
}
