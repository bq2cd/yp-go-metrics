package handler

import (
	"context"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type updateHandler struct {
	baseHandler
	metrics service.MetricStorer
	auditor service.MetricAuditor
}

// ServeHTTP implements http.Handler for /update/* endpoint with plain-text requests/responses.
func (h *updateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric, ok, l := h.validateMetric(w, r)
	if !ok {
		return
	}

	if !h.storeMetric(w, l, r.Context(), metric) {
		return
	}

	h.auditor.RecordMetricsUploaded(r.Context(), model.NewMetricSet(metric), h.getClientInfo(r))

	h.respondOK(w)
}

func (h *updateHandler) validateMetric(w http.ResponseWriter, r *http.Request) (model.Metric, bool, log.Logger) {
	metric, err := urlpath.NewOperationFromURLPath(r.URL.Path).ToMetric()

	l := h.logger.With(log.Str("urlpath", r.URL.Path), log.Any("metric", metric))

	switch err {
	case nil:
		return metric, h.validateMetricType(w, l, metric), l
	case urlpath.ErrMissingMetricID:
		h.respondError(w, http.StatusNotFound, l, err, "")
	default:
		h.respondError(w, http.StatusBadRequest, l, err, "unknown metric operation")
	}

	return metric, false, l
}

func (h *updateHandler) validateMetricType(w http.ResponseWriter, l log.Logger, metric model.Metric) bool {
	switch metric.Type {
	case model.MetricTypeCounter:
		return true
	case model.MetricTypeGauge:
		return true
	default:
		h.respondError(w, http.StatusBadRequest, l, nil, "unknown metric type")
	}

	return false
}

func (h *updateHandler) storeMetric(w http.ResponseWriter, l log.Logger, ctx context.Context, metric model.Metric) bool {
	err := h.metrics.StoreSingle(ctx, metric.Copy())
	if err == nil {
		return true
	}

	h.respondError(w, http.StatusInternalServerError, l, err, "cannot store metric")

	return false
}

func (h *updateHandler) respondOK(w http.ResponseWriter) {
	httpheaders.ContentTypeTextPlain.UTF8().Apply(w.Header())
	w.WriteHeader(http.StatusOK)
}
