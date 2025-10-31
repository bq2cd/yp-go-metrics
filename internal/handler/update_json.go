package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type updateJSONHandler struct {
	baseHandler
	metrics   service.MetricStorer
	responder metricJSONResponder
}

// ServeHTTP implements http.Handler for /update endpoint with JSON requests/responses.
func (h *updateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !httpheaders.ContentTypeApplicationJSON.Matches(r.Header) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var m model.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "cannot decode metric", http.StatusUnprocessableEntity)
		return
	}

	if m.Key().Empty() {
		http.Error(w, "empty metric type or id", http.StatusBadRequest)
		return
	}

	if err := h.metrics.StoreSingle(r.Context(), m); err != nil {
		h.logger.Error().Err("error", err).Any("metric", m).Msg("cannot store metric")
		http.Error(w, "cannot store metric", http.StatusInsufficientStorage)
		return
	}

	m, err = h.metrics.RetrieveSingle(r.Context(), m.Key())
	if err != nil {
		h.logger.Error().Err("error", err).Any("metric_key", m.Key()).Msg("cannot retrieve metric")
		http.Error(w, "cannot retrieve metric", http.StatusNotFound)
		return
	}

	if err := h.responder.WriteResponse(w, m); err != nil {
		h.logger.Error().Err("error", err).Any("metric", m).Msg("json encoder failed")
	}
}
