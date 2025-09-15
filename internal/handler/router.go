package handler

import (
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"net/http"
)

// Router implements HTTP routing for the server part.
type router struct {
	mux     *http.ServeMux
	metrics service.Metrics
}

// NewRouter instantiates a router with necessary dependencies.
func NewRouter(metrics service.Metrics) *router {
	mux := http.NewServeMux()

	mux.Handle("/update/", &updateHandler{metrics: metrics})

	return &router{
		mux:     mux,
		metrics: metrics,
	}
}

// ServeHTTP implements http.Handler interface
func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}
