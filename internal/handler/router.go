package handler

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/middleware"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
)

// Router implements HTTP routing for the server part.
type router struct {
	logger  log.Logger
	mux     http.Handler
	metrics service.Metrics
}

// NewRouter instantiates a router with necessary dependencies.
func NewRouter(logger log.Logger, metrics service.Metrics, mux http.Handler) *router {
	if mux != nil {
		return &router{
			logger:  logger.With(log.Str("subsystem", "router")),
			mux:     mux,
			metrics: metrics,
		}
	}

	rt := &router{logger: logger, mux: chi.NewRouter(), metrics: metrics}

	rt.configureChiRouter()

	return rt
}

func (rt *router) configureChiRouter() {
	r, ok := rt.mux.(*chi.Mux)
	if !ok {
		return
	}

	// middlewares
	r.Use(
		middleware.RequestID(),
		middleware.Logger(rt.logger),
		middleware.Recoverer(rt.logger),
	)

	// routes
	r.Handle("/*", &defaultHandler{})

	r.Method(http.MethodGet, "/", &readHandler{metrics: rt.metrics})
	r.Method(http.MethodPost, "/update/*", &updateHandler{metrics: rt.metrics})
	r.Method(http.MethodGet, "/value/*", &valueHandler{metrics: rt.metrics})
}

// ServeHTTP implements http.Handler interface
func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}
