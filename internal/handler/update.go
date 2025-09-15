package handler

import (
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"net/http"
)

type updateHandler struct {
	metrics service.Metrics
}

// ServeHTTP implements http.Handler for /update endpoint
func (h *updateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
