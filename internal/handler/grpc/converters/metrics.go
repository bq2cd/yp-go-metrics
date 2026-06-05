package converters

import (
	"strings"

	pbmetrics "github.com/bq2cd/yp-go-metrics/api/gen/metrics/v1"
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

// ProtoToMetrics converts protobuf metrics to a slice of [model.Metric].
// `nil` protobuf metrics are skipped.
func ProtoToMetrics(pbMetrics []*pbmetrics.Metric) []model.Metric {
	out := make([]model.Metric, 0, len(pbMetrics))

	for _, pb := range pbMetrics {
		if pb == nil {
			continue
		}

		var m model.Metric

		switch pb.GetType() {
		case pbmetrics.Metric_COUNTER:
			m = model.NewCounterMetric(pb.GetId(), pb.GetDelta())
		case pbmetrics.Metric_GAUGE:
			m = model.NewGaugeMetric(pb.GetId(), pb.GetValue())
		default:
			delta := pb.GetDelta()
			value := pb.GetValue()

			m = model.Metric{
				ID:    pb.GetId(),
				Type:  model.MetricType(strings.ToLower(pb.GetType().String())),
				Delta: &delta,
				Value: &value,
			}
		}

		out = append(out, m)
	}

	return out
}

// MetricsToProto converts a slice of metrics into protobuf metrics.
func MetricsToProto(metrics ...model.Metric) []*pbmetrics.Metric {
	out := make([]*pbmetrics.Metric, 0, len(metrics))

	for _, m := range metrics {
		pb := new(pbmetrics.Metric)
		pb.SetId(m.ID)

		switch m.Type {
		case model.MetricTypeCounter:
			pb.SetType(pbmetrics.Metric_COUNTER)
			if m.Delta != nil {
				pb.SetDelta(*m.Delta)
			}
		case model.MetricTypeGauge:
			pb.SetType(pbmetrics.Metric_GAUGE)
			if m.Value != nil {
				pb.SetValue(*m.Value)
			}
		default:
			pb.SetType(pbmetrics.Metric_UNSPECIFIED)
			if m.Delta != nil {
				pb.SetDelta(*m.Delta)
			}
			if m.Value != nil {
				pb.SetValue(*m.Value)
			}
		}

		out = append(out, pb)
	}

	return out
}
