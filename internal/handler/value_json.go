package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type valueJSONHandler struct {
	baseHandler
	metrics   service.MetricStorer
	responder metricJSONResponder
}

// ServeHTTP implements http.Handler for /value endpoint with JSON requests/responses.
func (h *valueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !httpheaders.ContentTypeApplicationJSON.Matches(r.Header) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	var needle model.Metric
	err := json.NewDecoder(r.Body).Decode(&needle)
	if err != nil {
		http.Error(w, "cannot decode metric", http.StatusUnprocessableEntity)
		return
	}

	if needle.Key().Empty() {
		http.Error(w, "empty metric type or id", http.StatusBadRequest)
		return
	}

	m, err := h.metrics.RetrieveSingle(needle.Key())
	switch err {
	case nil:
		if err := h.responder.WriteResponse(w, m); err != nil {
			h.logger.Error().Err("error", err).Any("metric", m).Msg("json encoder failed")
		}
	case service.ErrMetricNotFound:
		http.Error(w, "metric not found", http.StatusNotFound)
	default:
		http.Error(w, "cannot retrieve metric", http.StatusInternalServerError)
	}
}
