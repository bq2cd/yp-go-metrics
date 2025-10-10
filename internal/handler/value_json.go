package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/contenttype"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type valueJSONHandler struct {
	logger    log.Logger
	metrics   service.Metrics
	responder metricJSONResponder
}

// ServeHTTP implements http.Handler for /value endpoint with JSON requests/responses.
func (h *valueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !contenttype.ApplicationJSON.MatchesRequest(r) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var needle model.Metric
	err := json.NewDecoder(r.Body).Decode(&needle)
	if err != nil {
		http.Error(w, "", http.StatusUnprocessableEntity)
		return
	}

	if needle.Key().Empty() {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	m, err := h.metrics.RetrieveSingle(needle.Key())
	switch err {
	case nil:
		if err := h.responder.WriteResponse(w, m); err != nil {
			h.logger.Error().Err("error", err).Any("metric", m).Msg("json encoder failed")
		}
	case repository.ErrMetricNotFound:
		http.Error(w, "", http.StatusNotFound)
	default:
		http.Error(w, "", http.StatusInternalServerError)
	}
}
