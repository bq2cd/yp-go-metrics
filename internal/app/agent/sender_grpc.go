package agent

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"

	pbmetrics "github.com/bq2cd/yp-go-metrics/api/gen/metrics/v1"
	"github.com/bq2cd/yp-go-metrics/internal/handler/grpc/converters"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

// NewSenderGRPC creates an instance of [SenderBatch] that reports collected metrics
// via gRPC using Protocol Buffers encoding.
func NewSenderGRPC(
	logger log.Logger,
	client pbmetrics.MetricsClient,
	realIP httpheaders.XRealIP,
) *senderGRPC {
	if logger == nil {
		logger = log.NewNoopLogger()
	}

	return &senderGRPC{
		logger: logger.With(log.Str("sender", "grpc")),
		client: client,
		realIP: realIP,
	}
}

type senderGRPC struct {
	logger log.Logger
	client pbmetrics.MetricsClient
	realIP httpheaders.XRealIP
}

// Send reports single metric to upstream server via gRPC.
func (s *senderGRPC) Send(ctx context.Context, metric model.Metric) error {
	if metric.Empty() {
		return ErrSenderEmptyMetric
	}

	_, err := s.SendBatch(ctx, model.NewMetricSet(metric))

	return err
}

// SendBatch reports multiple metrics to upstream server via gRPC.
func (s *senderGRPC) SendBatch(ctx context.Context, metrics model.MetricSet) (model.MetricSet, error) {
	if metrics.Empty() {
		return model.NewMetricSet(), nil
	}

	req := new(pbmetrics.UpdateMetricsRequest)
	req.SetMetrics(s.prepareProtoMetrics(metrics))

	resp, err := s.client.UpdateMetrics(s.prepareMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("cannot upload metrics via gRPC: %w", err)
	}

	return s.protoResponseToMetrics(resp), nil
}

func (s *senderGRPC) prepareProtoMetrics(metrics model.MetricSet) []*pbmetrics.Metric {
	return converters.MetricsToProto(metrics.Values()...)
}

func (s *senderGRPC) prepareMetadata(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{
		strings.ToLower(httpheaders.HeaderKeyXRealIP): s.realIP.String(),
	})

	return metadata.NewOutgoingContext(ctx, md)
}

func (s *senderGRPC) protoResponseToMetrics(resp *pbmetrics.UpdateMetricsResponse) model.MetricSet {
	metrics := converters.ProtoToMetrics(resp.GetMetrics())

	return model.NewMetricSet(metrics...)
}
