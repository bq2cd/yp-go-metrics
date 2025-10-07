package handler

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type updateJSONHandler struct {
	metrics service.Metrics
}

// ServeHTTP implements http.Handler for /update endpoint with JSON requests/responses.
func (h *updateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: implement me
	var m model.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "", http.StatusUnprocessableEntity)
		return
	}
}
