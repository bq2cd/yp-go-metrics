package handler

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"

	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type valueHandler struct {
	metrics service.Metrics
}

// ServeHTTP implements http.Handler for /value/* endpoint with plain-text requests/responses.
func (h *valueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricOp := urlpath.NewOperationFromURLPath(r.URL.Path)
	needle, err := metricOp.ToMetric()
	switch err {
	case urlpath.ErrMissingMetricID:
		http.Error(w, "", http.StatusNotFound)
		return
	case urlpath.ErrMissingMetricValue:
		// OK
	default:
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	metric, err := h.metrics.RetrieveSingle(needle.Key())
	switch {
	case errors.Is(err, repository.ErrMetricNotFound):
		http.Error(w, "", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "", http.StatusInternalServerError)
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
