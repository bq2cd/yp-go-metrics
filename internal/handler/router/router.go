package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/handler/middleware"
	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

var (
	// ErrRouteEmptyPatterns is returned when [RouteDefinition] contains no patterns.
	ErrRouteEmptyPatterns = errors.New("empty patterns")
	// ErrRouteEmptyHandler is returned when [RouteDefinition] has a `nil` handler.
	ErrRouteEmptyHandler = errors.New("empty handler")
	// ErrRouteDuplicatePatterns is returned when [RouteDefinitions] has duplicate patterns.
	ErrRouteDuplicatePatterns = errors.New("duplicate patterns")
)

// Middlewares returns a pre-configured list of global middleware handlers (applied to all routes).
func Middlewares(logger log.Logger, signer hmacsigner.HMACSigner) []middleware.Middleware {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	return []middleware.Middleware{
		middleware.RequestID(),
		middleware.Compressor(logger),
		middleware.Logger(logger),
		middleware.HMACSigner(logger, signer),
		middleware.Recoverer(logger),
	}
}

// RouteDefinitions returns a pre-configured list of patterns assigned to a [handler.Ident].
func RouteDefinitions() []RouteDefinition {
	return []RouteDefinition{
		{ident: handler.IdentDefault, patterns: []string{"/*"}},
		{ident: handler.IdentRead, patterns: []string{"GET /"}},
		{ident: handler.IdentUpdateJSON, patterns: []string{"POST /update", "POST /update/"}, encrypted: true},
		{ident: handler.IdentUpdateBatchJSON, patterns: []string{"POST /updates", "POST /updates/"}, encrypted: true},
		{ident: handler.IdentUpdate, patterns: []string{"POST /update/*"}},
		{ident: handler.IdentValueJSON, patterns: []string{"POST /value", "POST /value/"}},
		{ident: handler.IdentValue, patterns: []string{"GET /value/*"}},
		{ident: handler.IdentPing, patterns: []string{"GET /ping"}},
	}
}

// RouteDefinition describes a link between a [handler.Ident] and a list of patterns.
type RouteDefinition struct {
	ident     handler.Ident
	patterns  []string
	encrypted bool
}

type routeDefinitionKey string

// Key returns a unique representation of [RouteDefinition] suitable for hashing (so that it can be used
// as a map key).
// This is called only on router initialization, so performance is not critical.
func (rd *RouteDefinition) Key() routeDefinitionKey {
	return routeDefinitionKey(fmt.Sprint(*rd))
}

// Routes returns a pre-configured list of routes with corresponding handlers from the given [handler.Registry].
func Routes(definitions []RouteDefinition, handlers handler.Registry, perRouteMiddlewares map[routeDefinitionKey]chi.Middlewares) ([]Route, error) {
	routes := make([]Route, len(definitions))

	for i, rd := range definitions {
		h, ok := handlers[rd.ident]
		if !ok {
			return nil, fmt.Errorf("missing handler %v", rd.ident)
		}

		route, err := NewRoute(h, rd.patterns...)
		if err != nil {
			return nil, fmt.Errorf("cannot create route: %w", err)
		}

		route.middlewares = perRouteMiddlewares[rd.Key()]

		routes[i] = route
	}

	return routes, nil
}

// Route maps a list of patterns to a single http handler
type Route struct {
	patterns    []string
	handler     http.Handler
	middlewares chi.Middlewares
}

// Validate ensures that route has at least one pattern and a non-nil http handler.
// Duplicate patterns are considered an error.
func (r *Route) Validate() error {
	if r.handler == nil {
		return ErrRouteEmptyHandler
	}
	if len(r.patterns) == 0 {
		return ErrRouteEmptyPatterns
	}
	seen := make(map[string]bool, len(r.patterns))
	for _, p := range r.patterns {
		if seen[p] {
			return ErrRouteDuplicatePatterns
		}
		seen[p] = true
	}
	return nil
}

// NewRoute constructs and validates a new [Route] instance from the
// given handler and patterns.
func NewRoute(handler http.Handler, patterns ...string) (Route, error) {
	seen := make(map[string]bool, len(patterns))
	filtered := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if !seen[p] {
			seen[p] = true
			filtered = append(filtered, p)
		}
	}
	rr := Route{
		patterns: filtered,
		handler:  handler,
	}
	if err := rr.Validate(); err != nil {
		return Route{}, err
	}
	return rr, nil
}

// Router implements [http.Handler] interface.
// Under the hood, it wraps a [chi.Router] instance.
type Router struct {
	mux chi.Router
}

// New returns an instance of the [Router] with preconfigured middlewares and routes.
// For each route, a handler must exist in the given [handler.Registry].
func New(logger log.Logger, handlers handler.Registry, signer hmacsigner.HMACSigner, decryptor asymcrypt.Decryptor) (*Router, error) {
	rtr := &Router{
		mux: chi.NewRouter(),
	}

	if logger == nil {
		logger = log.NewNoopLogger()
	}

	logger = logger.With(log.Str("subsystem", "router"))

	if signer == nil {
		return nil, fmt.Errorf("signer cannot be nil")
	}

	middlewares := Middlewares(logger, signer)
	definitions := RouteDefinitions()
	perRouteMiddlewares := constructPerRouteMiddlewares(definitions, logger, decryptor)

	routes, err := Routes(definitions, handlers, perRouteMiddlewares)
	if err != nil {
		return nil, fmt.Errorf("cannot obtain routes: %w", err)
	}

	err = configureChiRouter(
		rtr.mux,
		middlewares,
		routes,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot configure router: %w", err)
	}
	return rtr, nil
}

func constructPerRouteMiddlewares(definitions []RouteDefinition, logger log.Logger, decryptor asymcrypt.Decryptor) map[routeDefinitionKey]chi.Middlewares {
	middlewares := make(map[routeDefinitionKey]chi.Middlewares)

	requestDecryptor := middleware.RequestDecryptor(logger, decryptor)
	recoverer := middleware.Recoverer(logger)

	for _, rd := range definitions {
		if rd.encrypted {
			middlewares[rd.Key()] = append(middlewares[rd.Key()], requestDecryptor, recoverer) // ensure recoverer is the last
		}
	}

	return middlewares
}

func configureChiRouter(mux chi.Router, middlewares []middleware.Middleware, routes []Route) error {
	for _, m := range middlewares {
		if m == nil {
			return errors.New("middleware cannot be nil")
		}
		mux.Use(m)
	}

	for _, route := range routes {
		if err := route.Validate(); err != nil {
			return fmt.Errorf("invalid route: %w", err)
		}
		for _, p := range route.patterns {
			r := mux
			if len(route.middlewares) > 0 {
				r = mux.With(route.middlewares...)
			}
			r.Handle(p, route.handler)
		}
	}

	return nil
}

// ServeHTTP implements [http.Handler] interface.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rtr.mux.ServeHTTP(w, r)
}
