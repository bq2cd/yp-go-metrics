package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type metricJSONResponder interface {
	WriteResponse(w http.ResponseWriter, m model.Metric) error
}

type defaultMetricJSONResponder struct{}

func (r *defaultMetricJSONResponder) WriteResponse(w http.ResponseWriter, m model.Metric) error {
	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(m)
}
