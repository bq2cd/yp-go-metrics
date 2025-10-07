package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type valueJSONHandler struct {
	metrics service.Metrics
}

// ServeHTTP implements http.Handler for /value endpoint with JSON requests/responses.
func (h *valueJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: implement me
	var m model.Metric
	_ = json.NewEncoder(w).Encode(m)
}
