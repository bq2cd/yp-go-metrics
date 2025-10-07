package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type updateJSONHandler struct {
	logger  log.Logger
	metrics service.Metrics
}

// ServeHTTP implements http.Handler for /update endpoint with JSON requests/responses.
func (h *updateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !contentTypeApplicationJSON.matchesRequest(r) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var m model.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "", http.StatusUnprocessableEntity)
		return
	}

	if m.Empty() {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if err := h.metrics.StoreSingle(m); err != nil {
		http.Error(w, "", http.StatusInsufficientStorage)
		return
	}

	contentTypeApplicationJSON.applyToResponse(w)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		h.logger.Error().Err("error", err).Any("metric", m).Msg("json encoder failed")
	}
}
