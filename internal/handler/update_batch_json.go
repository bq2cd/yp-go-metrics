package handler

import (
	"context"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type updateBatchJSONHandler struct {
	baseHandler
	metrics   service.MetricStorer
	responder metricBatchJSONResponder
	auditor   service.MetricAuditor
}

// ServeHTTP implements [Handler] for /update endpoint with JSON requests/responses.
func (h *updateBatchJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.validateContentType(w, r) {
		return
	}

	metrics, ok := h.validateMetrics(w, r)
	if !ok {
		return
	}

	keys, ok := h.storeMetrics(w, r.Context(), metrics)
	if !ok {
		return
	}

	h.auditor.RecordMetricsUploaded(r.Context(), model.NewMetricSet(metrics...), h.getClientInfo(r))

	metrics, ok = h.retrieveMetrics(w, r.Context(), keys)
	if !ok {
		return
	}

	h.respondOK(w, metrics)
}

func (h *updateBatchJSONHandler) validateContentType(w http.ResponseWriter, r *http.Request) bool {
	if httpheaders.ContentTypeApplicationJSON.Matches(r.Header) {
		return true
	}

	h.respondError(w, http.StatusBadRequest, h.logger, nil, "invalid content type")

	return false
}

func (h *updateBatchJSONHandler) validateMetrics(w http.ResponseWriter, r *http.Request) ([]model.Metric, bool) {
	var metrics []model.Metric

	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err == nil {
		return metrics, true
	}

	h.respondError(w, http.StatusUnprocessableEntity, h.logger, err, "cannot decode metrics")

	return metrics, false
}

func (h *updateBatchJSONHandler) storeMetrics(w http.ResponseWriter, ctx context.Context, metrics []model.Metric) ([]model.MetricKey, bool) {
	keys := make([]model.MetricKey, len(metrics))
	storable := make([]model.Metric, len(metrics))
	for i, m := range metrics {
		keys[i] = m.Key()
		storable[i] = m.Copy()
	}

	err := h.metrics.StoreBatch(ctx, storable)
	if err == nil {
		return keys, true
	}

	l := h.logger.With(log.Int("num_metrics", len(metrics)))

	h.respondError(w, http.StatusInsufficientStorage, l, err, "cannot store metrics")

	return nil, false
}

func (h *updateBatchJSONHandler) retrieveMetrics(w http.ResponseWriter, ctx context.Context, metricKeys []model.MetricKey) ([]model.Metric, bool) {
	metrics, err := h.metrics.RetrieveBatch(ctx, metricKeys)
	if err == nil {
		return metrics, true
	}

	l := h.logger.With(log.Int("num_metrics", len(metrics)))

	h.respondError(w, http.StatusInternalServerError, l, err, "cannot retrieve metrics")

	return metrics, false
}

func (h *updateBatchJSONHandler) respondOK(w http.ResponseWriter, metrics []model.Metric) {
	err := h.responder.WriteResponse(w, metrics)
	if err == nil {
		return
	}

	h.logger.Error().WithErr(err).Int("num_metrics", len(metrics)).Msg("json encoder failed")
}

type metricBatchJSONResponder interface {
	WriteResponse(w http.ResponseWriter, metrics []model.Metric) error
}

type defaultMetricBatchJSONResponder struct{}

func (r *defaultMetricBatchJSONResponder) WriteResponse(w http.ResponseWriter, metrics []model.Metric) error {
	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(metrics)
}
