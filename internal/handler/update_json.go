package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type updateJSONHandler struct {
	baseHandler
	metrics   service.MetricStorer
	responder metricJSONResponder
	auditor   service.MetricAuditor
}

// ServeHTTP implements http.Handler for /update endpoint with JSON requests/responses.
func (h *updateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.validateContentType(w, r) {
		return
	}

	metric, ok := h.validateMetric(w, r)
	if !ok {
		return
	}

	if !h.storeMetric(w, r.Context(), metric) {
		return
	}

	h.auditor.RecordMetricsUploaded(r.Context(), model.NewMetricSet(metric), h.getClientInfo(r))

	metric, ok = h.retrieveMetric(w, r.Context(), metric.Key())
	if !ok {
		return
	}

	h.respondOK(w, metric)
}

func (h *updateJSONHandler) validateContentType(w http.ResponseWriter, r *http.Request) bool {
	if httpheaders.ContentTypeApplicationJSON.Matches(r.Header) {
		return true
	}

	h.respondError(w, http.StatusBadRequest, h.logger, nil, "invalid content type")

	return false
}

func (h *updateJSONHandler) validateMetric(w http.ResponseWriter, r *http.Request) (model.Metric, bool) {
	var metric model.Metric

	err := json.NewDecoder(r.Body).Decode(&metric)
	if err == nil {
		return metric, true
	}

	h.respondError(w, http.StatusUnprocessableEntity, h.logger, err, "cannot decode metric")

	return metric, false
}

func (h *updateJSONHandler) storeMetric(w http.ResponseWriter, ctx context.Context, metric model.Metric) bool {
	err := h.metrics.StoreSingle(ctx, metric.Copy())
	if err == nil {
		return true
	}

	if errors.Is(err, service.ErrMetricIsEmpty) {
		return true
	}

	l := h.logger.With(log.Any("metric", metric))

	switch {
	case errors.Is(err, service.ErrMetricNotFound):
		h.respondError(w, http.StatusNotFound, l, err, "")
	case errors.Is(err, service.ErrMetricKeyIsEmpty):
		h.respondError(w, http.StatusBadRequest, l, err, "empty metric type or id")
	default:
		h.respondError(w, http.StatusInsufficientStorage, l, err, "cannot store metric")
	}

	return false
}

func (h *updateJSONHandler) retrieveMetric(w http.ResponseWriter, ctx context.Context, metricKey model.MetricKey) (model.Metric, bool) {
	metric, err := h.metrics.RetrieveSingle(ctx, metricKey)
	if err == nil {
		return metric, true
	}

	l := h.logger.With(log.Any("metric_key", metricKey))
	status := http.StatusInternalServerError
	msg := "cannot retrieve metric"

	if errors.Is(err, service.ErrMetricNotFound) {
		status = http.StatusNotFound
	}

	h.respondError(w, status, l, err, msg)

	return metric, false
}

func (h *updateJSONHandler) respondOK(w http.ResponseWriter, metric model.Metric) {
	err := h.responder.WriteResponse(w, metric)
	if err == nil {
		return
	}

	h.logger.Error().WithErr(err).Any("metric", metric).Msg("json encoder failed")
}

type metricJSONResponder interface {
	WriteResponse(w http.ResponseWriter, m model.Metric) error
}

type defaultMetricJSONResponder struct{}

func (r *defaultMetricJSONResponder) WriteResponse(w http.ResponseWriter, m model.Metric) error {
	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(m)
}
