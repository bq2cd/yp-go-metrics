package handler

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type updateHandler struct {
	metrics service.MetricStorer
}

// ServeHTTP implements http.Handler for /update/* endpoint with plain-text requests/responses.
func (h *updateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric, err := urlpath.NewOperationFromURLPath(r.URL.Path).ToMetric()
	switch err {
	case nil:
		switch metric.Type {
		case model.MetricTypeCounter:
			// ok
		case model.MetricTypeGauge:
			// ok
		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
			return
		}
	case urlpath.ErrMissingMetricID:
		http.Error(w, "missing metric id", http.StatusNotFound)
		return
	default:
		http.Error(w, "unknown metric operation", http.StatusBadRequest)
		return
	}

	err = h.metrics.StoreSingle(metric)
	if err != nil {
		http.Error(w, "cannot store metric", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
