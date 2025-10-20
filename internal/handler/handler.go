package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

const (
	IdentDefault    = "default"
	IdentRead       = "read"
	IdentUpdate     = "update"
	IdentUpdateJSON = "update_json"
	IdentValue      = "value"
	IdentValueJSON  = "value_json"
)

// Ident represents a handler ID.
type Ident string

// Registry maps a handler ID ([Ident]) to a handler implementation.
type Registry map[Ident]http.Handler

// NewRegistry creates a new [Registry] map.
func NewRegistry(logger log.Logger, metrics service.MetricStorer) Registry {
	return Registry{
		IdentDefault:    &defaultHandler{},
		IdentRead:       &readHandler{metrics: metrics},
		IdentUpdate:     &updateHandler{metrics: metrics},
		IdentUpdateJSON: &updateJSONHandler{logger: logger, metrics: metrics, responder: &defaultMetricJSONResponder{}},
		IdentValue:      &valueHandler{metrics: metrics},
		IdentValueJSON:  &valueJSONHandler{logger: logger, metrics: metrics, responder: &defaultMetricJSONResponder{}},
	}
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
