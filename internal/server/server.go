package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/log"
)

// ListenerFactory abstracts a way to create a new listener.
// Mostly useful for testing.
type ListenerFactory interface {
	Create(context.Context, string) (net.Listener, error)
}

type listenerFactory struct{}

// Create sets up a new listener on a given address.
func (f *listenerFactory) Create(ctx context.Context, addr string) (net.Listener, error) {
	var lcfg net.ListenConfig
	ln, err := lcfg.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %v: %w", addr, err)
	}
	return ln, nil
}

type server struct {
	logger    log.Logger
	context   context.Context
	config    config.Config
	router    http.Handler
	lnFactory ListenerFactory
}

// NewServer creates an instance of a server worker.
func NewServer(logger log.Logger, ctx context.Context, cfg config.Config, router http.Handler) *server {
	l := logger
	if l == nil {
		l = log.NewNoopLogger()
	}
	return &server{
		logger:    l.With(log.Str("subsystem", "server")),
		context:   ctx,
		config:    cfg,
		router:    router,
		lnFactory: &listenerFactory{},
	}
}

// Run launches main loop of the server worker:
// listening on provided address and serving incoming HTTP requests.
func (s *server) Run() error {
	s.logger.Info().Str("address", s.config.ListenAddress).Msg("listening for incoming connections")

	ln, err := s.lnFactory.Create(s.context, s.config.ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %v: %w", s.config.ListenAddress, err)
	}

	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	srv := &http.Server{
		Handler: s.router,
		BaseContext: func(l net.Listener) context.Context {
			return baseCtx
		},
	}

	errCh := make(chan error, 1)

	go func() {
		err := srv.Serve(ln)
		switch err {
		case nil, http.ErrServerClosed:
			return
		default:
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-s.context.Done():
		s.logger.Info().Dur("timeout", s.config.ShutdownTimeout).Msg("shutting down gracefully")
		ctx, cancel := context.WithTimeout(baseCtx, s.config.ShutdownTimeout)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
