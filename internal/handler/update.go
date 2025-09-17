package handler

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type updateHandler struct {
	metrics service.Metrics
}

// ServeHTTP implements http.Handler for /update endpoint
func (h *updateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric, err := NewMetricFromURLPath(r.URL.Path)
	switch err {
	case nil:
		// pass
	case ErrEmptyMetricID:
		http.Error(w, "", http.StatusNotFound)
		return
	default:
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	err = h.metrics.Store(metric)
	if err != nil {
		http.Error(w, "failed to update metric", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
