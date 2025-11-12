package handler

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"

	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type valueHandler struct {
	baseHandler
	metrics service.MetricStorer
}

// ServeHTTP implements http.Handler for /value/* endpoint with plain-text requests/responses.
func (h *valueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricOp := urlpath.NewOperationFromURLPath(r.URL.Path)
	needle, err := metricOp.ToMetric()
	l := h.logger.With(log.Str("urlpath", r.URL.Path), log.Any("needle", needle))
	switch err {
	case urlpath.ErrMissingMetricID:
		l.Error().Msg("missing metric id")
		w.WriteHeader(http.StatusNotFound)
		return
	case urlpath.ErrMissingMetricValue:
		// OK
	default:
		l.Error().Msg("missing metric id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.metrics.RetrieveSingle(r.Context(), needle.Key())
	switch {
	case errors.Is(err, service.ErrMetricNotFound):
		l.Error().Msg("metric not found")
		w.WriteHeader(http.StatusNotFound)
		return
	case err != nil:
		l.Error().WithErr(err).Msg("cannot retrieve metric from storage")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	out := new(bytes.Buffer)
	switch metric.Type {
	case model.MetricTypeCounter:
		out.WriteString(strconv.FormatInt(*metric.Delta, 10))
	default:
		out.WriteString(strconv.FormatFloat(*metric.Value, 'g', 10, 64))
	}

	w.Write(out.Bytes())
}
