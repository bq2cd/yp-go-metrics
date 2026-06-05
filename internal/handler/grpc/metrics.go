package grpc

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	pbmetrics "github.com/bq2cd/yp-go-metrics/api/gen/metrics/v1"
	"github.com/bq2cd/yp-go-metrics/internal/handler/grpc/converters"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type metricsHandler struct {
	pbmetrics.UnimplementedMetricsServer

	logger  log.Logger
	storer  service.MetricStorer
	auditor service.MetricAuditor
}

// NewMetricsHandler returns an implementation of [pbmetrics.MetricsServer].
func NewMetricsHandler(logger log.Logger, storer service.MetricStorer, auditor service.MetricAuditor) *metricsHandler {
	return &metricsHandler{
		logger:  logger,
		storer:  storer,
		auditor: auditor,
	}
}

// UpdateMetrics takes incoming metrics, performs batch update of the server's storage, and
// returns metrics with updated values.
// Updated value for gauges would be the same as incoming value, but for counters it will be added
// to the existing value.
func (h *metricsHandler) UpdateMetrics(ctx context.Context, req *pbmetrics.UpdateMetricsRequest) (*pbmetrics.UpdateMetricsResponse, error) {
	var resp pbmetrics.UpdateMetricsResponse

	metrics := converters.ProtoToMetrics(req.GetMetrics())

	stored, err := h.storeMetrics(ctx, metrics)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot store metrics: %v", err)
	}

	h.auditor.RecordMetricsUploaded(ctx, model.NewMetricSet(metrics...), h.getClientInfo(ctx))

	metrics, err = h.storer.RetrieveBatch(ctx, stored)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "cannot retrieve stored metrics: %v", err)
	}

	resp.SetMetrics(converters.MetricsToProto(metrics...))

	return &resp, nil
}

func (h *metricsHandler) storeMetrics(ctx context.Context, metrics []model.Metric) ([]model.MetricKey, error) {
	keys := make([]model.MetricKey, len(metrics))
	storable := make([]model.Metric, len(metrics))
	for i, m := range metrics {
		keys[i] = m.Key()
		storable[i] = m.Copy() // when being stored, [model.Metric] is modified in place to add delta for existing counters
	}

	err := h.storer.StoreBatch(ctx, storable)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (h *metricsHandler) getClientInfo(ctx context.Context) model.ClientInfo {
	ip := h.getRemoteIPFromMetadata(ctx)
	if ip == nil {
		ip = h.getRemoteIPFromPeer(ctx)
	}

	return model.ClientInfo{
		IPAddress: ip.String(), // will return `<nil>` if ip is `nil`; we're okay with that
	}
}

func (h *metricsHandler) getRemoteIPFromMetadata(ctx context.Context) net.IP {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	values := md.Get(strings.ToLower(httpheaders.HeaderKeyXRealIP))
	if len(values) != 1 {
		return nil
	}

	return httpheaders.GetXRealIPFromBytes([]byte(values[0])).IP
}

func (h *metricsHandler) getRemoteIPFromPeer(ctx context.Context) net.IP {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}

	addr, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		return nil
	}

	return addr.IP
}
