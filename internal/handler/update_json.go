package handler

import (
	"errors"
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var m model.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		h.logger.Error().WithErr(err).Msg("cannot decode metric")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := h.metrics.StoreSingle(r.Context(), m); err != nil {
		switch {
		case errors.Is(err, service.ErrMetricNotFound):
			h.logger.Error().Msg("metric not found")
			w.WriteHeader(http.StatusNotFound)
			return
		case errors.Is(err, service.ErrMetricKeyIsEmpty):
			h.logger.Error().Msg("empty metric type or id")
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, service.ErrMetricIsEmpty):
			// return existing metric if any
		default:
			h.logger.Error().WithErr(err).Any("metric", m).Msg("cannot store metric")
			w.WriteHeader(http.StatusInsufficientStorage)
			return
		}
	}

	m, err = h.metrics.RetrieveSingle(r.Context(), m.Key())
	if err != nil {
		h.logger.Error().WithErr(err).Any("metric_key", m.Key()).Msg("cannot retrieve metric")
		if errors.Is(err, service.ErrMetricNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := h.responder.WriteResponse(w, m); err != nil {
		h.logger.Error().WithErr(err).Any("metric", m).Msg("json encoder failed")
	}
}
