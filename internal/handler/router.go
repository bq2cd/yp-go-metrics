package handler

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/middleware"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
)

// route maps a slice of patterns to a single http handler
type route struct {
	patterns []string
	handler  http.Handler
}

func newRoute(handler http.Handler, patterns ...string) route {
	seen := make(map[string]bool, len(patterns))
	filtered := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if !seen[p] {
			seen[p] = true
			filtered = append(filtered, p)
		}
	}
	return route{
		patterns: filtered,
		handler:  handler,
	}
}

// router implements HTTP routing for the server part.
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
	for _, m := range rt.getMiddlewares() {
		r.Use(m)
	}

	// routes
	for _, rr := range rt.getRoutes() {
		for _, p := range rr.patterns {
			r.Handle(p, rr.handler)
		}
	}
}

func (rt *router) getMiddlewares() []middleware.Middleware {
	return []middleware.Middleware{
		middleware.RequestID(),
		middleware.Compressor(rt.logger),
		middleware.Logger(rt.logger),
		middleware.Recoverer(rt.logger),
	}
}

func (rt *router) getRoutes() []route {
	return []route{
		newRoute(&defaultHandler{}, "/*"),
		newRoute(&readHandler{metrics: rt.metrics}, "GET /"),
		newRoute(&updateJSONHandler{logger: rt.logger, metrics: rt.metrics, responder: &defaultMetricJSONResponder{}}, "POST /update", "POST /update/"),
		newRoute(&updateHandler{metrics: rt.metrics}, "POST /update/*"),
		newRoute(&valueJSONHandler{logger: rt.logger, metrics: rt.metrics, responder: &defaultMetricJSONResponder{}}, "POST /value", "POST /value/"),
		newRoute(&valueHandler{metrics: rt.metrics}, "GET /value/*"),
	}
}

// ServeHTTP implements http.Handler interface
func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}
