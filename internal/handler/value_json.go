package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type valueJSONHandler struct {
	baseHandler

	metrics   service.MetricStorer
	responder metricJSONResponder
}

// ServeHTTP implements [Handler] for /value endpoint with JSON requests/responses.
func (h *valueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !httpheaders.ContentTypeApplicationJSON.Matches(r.Header) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var needle model.Metric
	err := json.NewDecoder(r.Body).Decode(&needle)
	if err != nil {
		h.logger.Error().WithErr(err).Msg("cannot decode metric")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	l := h.logger.With(log.Any("needle", needle))

	m, err := h.metrics.RetrieveSingle(r.Context(), needle.Key())
	switch err {
	case nil:
		err = h.responder.WriteResponse(w, m)
		if err != nil {
			l.Error().WithErr(err).Any("metric", m).Msg("json encoder failed")
		}
	case service.ErrMetricKeyIsEmpty:
		l.Error().Msg("empty metric type or id")
		w.WriteHeader(http.StatusBadRequest)
	case service.ErrMetricNotFound:
		l.Error().Msg("metric not found")
		w.WriteHeader(http.StatusNotFound)
	default:
		l.Error().WithErr(err).Msg("cannot retrieve metric from storage")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
