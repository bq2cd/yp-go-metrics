package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

//go:generate go tool mockgen -destination=handlertest/mock_handler.go -package=handlertest github.com/bq2cd/yp-go-metrics/internal/handler Handler

type Handler = http.Handler

const (
	IdentDefault    = "default"
	IdentRead       = "read"
	IdentUpdate     = "update"
	IdentUpdateJSON = "update_json"
	IdentValue      = "value"
	IdentValueJSON  = "value_json"
	IdentPing       = "ping"
)

func getHandlers(metrics service.MetricStorer, pinger service.StoragePinger) map[Ident]handlerLogger {
	return map[Ident]handlerLogger{
		IdentDefault:    &defaultHandler{},
		IdentRead:       &readHandler{metrics: metrics},
		IdentUpdate:     &updateHandler{metrics: metrics},
		IdentUpdateJSON: &updateJSONHandler{metrics: metrics, responder: &defaultMetricJSONResponder{}},
		IdentValue:      &valueHandler{metrics: metrics},
		IdentValueJSON:  &valueJSONHandler{metrics: metrics, responder: &defaultMetricJSONResponder{}},
		IdentPing:       &pingHandler{pinger: pinger, timeout: 500 * time.Millisecond},
	}
}

// Ident represents a handler ID.
type Ident string

// Registry maps a handler ID ([Ident]) to a [http.Handler] implementation.
type Registry map[Ident]Handler

// NewRegistry creates a new [Registry] map.
func NewRegistry(logger log.Logger, metrics service.MetricStorer, pinger service.StoragePinger) Registry {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	handlers := getHandlers(metrics, pinger)
	reg := make(Registry, len(handlers))
	for ident, h := range handlers {
		h.setLogger(logger.With(log.Str("handler", string(ident))))
		reg[ident] = h
	}
	return reg
}

// handlerLogger is an internal interface to facilitate common logging configuration
// and testing.
type handlerLogger interface {
	Handler
	setLogger(logger log.Logger)
	getLogger() log.Logger
}

type baseHandler struct {
	logger log.Logger
}

func (h *baseHandler) setLogger(logger log.Logger) {
	h.logger = logger
}

func (h *baseHandler) getLogger() log.Logger {
	return h.logger
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
