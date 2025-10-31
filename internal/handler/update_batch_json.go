package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type updateBatchJSONHandler struct {
	baseHandler
	metrics   service.MetricStorer
	responder metricBatchJSONResponder
}

// ServeHTTP implements http.Handler for /update endpoint with JSON requests/responses.
func (h *updateBatchJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !httpheaders.ContentTypeApplicationJSON.Matches(r.Header) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}
	var metrics []model.Metric
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		h.logger.Error().WithErr(err).Msg("cannot decode metrics")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := h.metrics.StoreBatch(r.Context(), metrics); err != nil {
		h.logger.Error().WithErr(err).Int("num_metrics", len(metrics)).Msg("cannot store metrics")
		w.WriteHeader(http.StatusInsufficientStorage)
		return
	}

	keys := make([]model.MetricKey, 0, len(metrics))
	for _, m := range metrics {
		keys = append(keys, m.Key())
	}

	metrics, err = h.metrics.RetrieveBatch(r.Context(), keys)
	if err != nil {
		h.logger.Error().WithErr(err).Int("num_metrics", len(metrics)).Msg("cannot retrieve metrics")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.responder.WriteResponse(w, metrics); err != nil {
		h.logger.Error().WithErr(err).Int("num_metrics", len(metrics)).Msg("json encoder failed")
	}
}
