package handler

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router implements HTTP routing for the server part.
type router struct {
	mux     http.Handler
	metrics service.Metrics
}

// NewRouter instantiates a router with necessary dependencies.
func NewRouter(metrics service.Metrics, mux http.Handler) *router {
	if mux != nil {
		return &router{
			mux:     mux,
			metrics: metrics,
		}
	}

	rt := chi.NewRouter()

	rt.Use(middleware.Logger)

	rt.Handle("/", &defaultHandler{})
	rt.Method(http.MethodPost, "/update/*", &updateHandler{metrics: metrics})

	return &router{
		mux:     rt,
		metrics: metrics,
	}
}

// ServeHTTP implements http.Handler interface
func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}
