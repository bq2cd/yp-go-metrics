package handler

import (
	"net/http"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

//go:generate go tool mockgen -destination=handlertest/mock_handler.go -package=handlertest github.com/bq2cd/yp-go-metrics/internal/handler Handler

type Handler = http.Handler

const (
	IdentDefault         = "default"
	IdentRead            = "read"
	IdentUpdate          = "update"
	IdentUpdateJSON      = "update_json"
	IdentUpdateBatchJSON = "update_batch_json"
	IdentValue           = "value"
	IdentValueJSON       = "value_json"
	IdentPing            = "ping"
)

func getHandlers(metrics service.MetricStorer, pinger service.StoragePinger, auditor service.MetricAuditor) map[Ident]handlerLogger {
	return map[Ident]handlerLogger{
		IdentDefault:         &defaultHandler{},
		IdentRead:            &readHandler{metrics: metrics},
		IdentUpdate:          &updateHandler{metrics: metrics, auditor: auditor},
		IdentUpdateJSON:      &updateJSONHandler{metrics: metrics, responder: &defaultMetricJSONResponder{}, auditor: auditor},
		IdentUpdateBatchJSON: &updateBatchJSONHandler{metrics: metrics, responder: &defaultMetricBatchJSONResponder{}, auditor: auditor},
		IdentValue:           &valueHandler{metrics: metrics},
		IdentValueJSON:       &valueJSONHandler{metrics: metrics, responder: &defaultMetricJSONResponder{}},
		IdentPing:            &pingHandler{pinger: pinger, timeout: 500 * time.Millisecond},
	}
}

// Ident represents a handler ID.
type Ident string

// Registry maps a handler ID ([Ident]) to a [http.Handler] implementation.
type Registry map[Ident]Handler

// NewRegistry creates a new [Registry] map.
func NewRegistry(logger log.Logger, metrics service.MetricStorer, pinger service.StoragePinger, auditor service.MetricAuditor) Registry {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	handlers := getHandlers(metrics, pinger, auditor)
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
