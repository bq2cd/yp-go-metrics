package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/periodictask"
	"github.com/bq2cd/yp-go-metrics/internal/service"
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
	logger      log.Logger
	context     context.Context
	config      config.Config
	router      http.Handler
	snapshotter service.MetricSnapshotter
	lnFactory   ListenerFactory
}

// NewServer creates an instance of a server worker.
func NewServer(logger log.Logger, ctx context.Context, cfg config.Config, router http.Handler, snapshotter service.MetricSnapshotter) *server {
	l := logger
	if l == nil {
		l = log.NewNoopLogger()
	}
	return &server{
		logger:      l.With(log.Str("subsystem", "server")),
		context:     ctx,
		config:      cfg,
		router:      router,
		snapshotter: snapshotter,
		lnFactory:   &listenerFactory{},
	}
}

func (s *server) listenAndServe() error {
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

func (s *server) loadMetrics() error {
	if !s.config.MetricStoreLoadOnStartup {
		return nil
	}
	if s.config.MetricStoreFilePath == "" {
		return nil
	}
	f, err := os.OpenFile(s.config.MetricStoreFilePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	// snapshotter will close the reader
	return s.snapshotter.LoadClose(f)
}

func (s *server) dumpMetrics() error {
	if s.config.MetricStoreFilePath == "" {
		return nil
	}
	f, err := os.OpenFile(s.config.MetricStoreFilePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	// snapshotter will close the writer
	return s.snapshotter.DumpClose(f)

}

func (s *server) createPeriodicTask(f func() error) periodictask.Task {
	var t periodictask.Task
	if s.config.MetricStoreInterval == 0 {
		taskFn := func(_ context.Context, _ struct{}) error { return f() }
		t = periodictask.NewChanTask(s.context, s.snapshotter.C(), taskFn)
	} else {
		taskFn := func(_ context.Context) error { return f() }
		t = periodictask.NewTimerTask(s.context, s.config.MetricStoreInterval, taskFn, s.config.MetricStoreInterval)
	}
	return t
}

// Run launches main loop of the server focused on two things:
// (1) listening on provided address and serving incoming HTTP requests;
// (2) periodically dumping received metrics to disk;
func (s *server) Run() error {
	if err := s.loadMetrics(); err != nil {
		return fmt.Errorf("failed to load metrics: %w", err)
	}
	var wg sync.WaitGroup

	errCh := make(chan error, 1)

	// launch http server
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- s.listenAndServe()
	}()

	// launch metric dumper
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- s.createPeriodicTask(s.dumpMetrics).Run()
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errFinal error
	for err := range errCh {
		errFinal = errors.Join(errFinal, err)
	}

	return errFinal
}
