package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/periodictask"
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
	logger         log.Logger
	config         config.Config
	router         http.Handler
	snapshotter    service.MetricSnapshotter
	batchWriter    service.StorageBatchWriter
	auditProcessor service.AuditEventProcessor
	grpcServer     *grpc.Server
	lnFactory      ListenerFactory
}

// New creates an instance of a server process that runs
// an HTTP server and other background tasks.
func New(
	logger log.Logger,
	cfg config.Config,
	router http.Handler,
	snapshotter service.MetricSnapshotter,
	batchWriter service.StorageBatchWriter,
	auditProcessor service.AuditEventProcessor,
	grpcServer *grpc.Server,
) *server {
	l := logger
	if l == nil {
		l = log.NewNoopLogger()
	}
	return &server{
		logger:         l.With(log.Str("subsystem", "server")),
		config:         cfg,
		router:         router,
		snapshotter:    snapshotter,
		batchWriter:    batchWriter,
		auditProcessor: auditProcessor,
		grpcServer:     grpcServer,
		lnFactory:      &listenerFactory{},
	}
}

// Run launches main loop of the server focused on the following things:
// (1) listening on the provided address and serving incoming HTTP and gRPC requests.
// (2) processing batch metric writes via [service.StorageBatchWriter].
// (3) processing audit events when metrics are uploaded (if configured).
// (4) periodically dumping uploaded metrics to disk (if configured);
func (s *server) Run(baseCtx context.Context) error {
	s.logConfig()

	ln, err := s.prepareListener(baseCtx)
	if err != nil {
		return err
	}

	grp, ctx := errgroup.WithContext(baseCtx)

	// background processing threads
	s.launchBatchWriter(ctx, grp)
	s.launchAuditProcessor(ctx, grp)
	s.launchMetricDumper(ctx, grp)

	// load metrics before starting to accept incoming requests
	s.tryLoadMetrics(ctx)

	// network servers: HTTP + gRPC
	s.launchNetworkServers(ctx, grp, ln)

	// wait for all background activity to finish
	err = grp.Wait()

	// perform one final dump before quitting
	err = errors.Join(err, s.performFinalDump())

	return err
}

func (s *server) logConfig() {
	sanitizedConfig := s.config

	sanitizedConfig.HMACSecretKey = fmt.Appendf([]byte{}, "<redacted(len=%d)>", len(s.config.HMACSecretKey))
	sanitizedConfig.DecryptionPrivateKey = fmt.Appendf([]byte{}, "<redacted(len=%d)>", len(s.config.DecryptionPrivateKey))

	s.logger.Info().Any("config", sanitizedConfig).Msg("starting with config")
}

func (s *server) prepareListener(ctx context.Context) (net.Listener, error) {
	s.logger.Info().Str("address", s.config.ListenAddress).Msg("listening for incoming connections")

	ln, err := s.lnFactory.Create(ctx, s.config.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %v: %w", s.config.ListenAddress, err)
	}

	return ln, nil
}

func (s *server) listenAndServeHTTP(ctx context.Context, ln net.Listener) error {
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

		errCh <- s.filterCloseErrors(err)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info().Dur("timeout", s.config.ShutdownTimeout).Msg("shutting down HTTP server gracefully")

		shutdownCtx, shutdownCancel := context.WithTimeout(baseCtx, s.config.ShutdownTimeout)
		defer shutdownCancel()

		return s.filterCloseErrors(srv.Shutdown(shutdownCtx))
	}
}

func (s *server) listenAndServeGRPC(ctx context.Context, ln net.Listener) error {
	errCh := make(chan error, 1)

	go func() {
		err := s.grpcServer.Serve(ln)

		errCh <- s.filterCloseErrors(err)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info().Dur("timeout", s.config.ShutdownTimeout).Msg("shutting down GRPC server gracefully")

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
		defer shutdownCancel()

		return s.shutdownGRPCServer(shutdownCtx)
	}
}

func (s *server) shutdownGRPCServer(ctx context.Context) error {
	successCh := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			// shutdown timeout expired (`context.DeadlineExceeded`), need to force-stop the server.
			s.grpcServer.Stop()
		case <-successCh:
			// graceful shutdown succeeded before shutdown timeout expired, so context got canceled;
			// this is a "successful" scenario.
		}
	}()

	// Initiate graceful shutdown; this will block and, if stuck, will be unlocked
	// by calling `s.grpcServer.Stop()` in goroutine after shutdown timeout expires (and context gets canceled).
	// See https://github.com/grpc/grpc-go/blob/master/examples/features/gracefulstop/server/main.go
	// for reference.
	s.grpcServer.GracefulStop()

	select {
	case <-ctx.Done():
		// context expired, force-stop should've been called in goroutine.
	default:
		close(successCh)
	}

	// avoid propating [context.DeadlineExceeded] errors as they are guaranteed to happen
	// when shutdown timeout is too short for TCP connection to be closed by a client.
	return nil
}

func (s *server) tryLoadMetrics(ctx context.Context) {
	if !s.config.MetricStoreLoadOnStartup {
		return
	}
	if s.config.MetricStoreFilePath == "" {
		return
	}
	l := s.logger.With(log.Str("path", s.config.MetricStoreFilePath), log.Str("component", "metric_loader"))

	l.Info().Msg("loading metrics from disk")
	f, err := os.OpenFile(s.config.MetricStoreFilePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		l.Error().WithErr(err).Msg("failed to open file")
		return
	}
	// snapshotter will close the reader
	err = s.snapshotter.LoadClose(ctx, f)
	if err != nil {
		l.Error().WithErr(err).Msg("failed to load metrics")
	}
}

func (s *server) dumpMetrics(ctx context.Context) error {
	if s.config.MetricStoreFilePath == "" {
		return nil
	}
	s.logger.Info().Str("path", s.config.MetricStoreFilePath).Msg("dumping metrics to disk")

	tmpfname := s.config.MetricStoreFilePath + ".tmp"
	tmpf, err := os.OpenFile(tmpfname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temporary file (%s): %w", tmpfname, err)
	}
	// snapshotter will close the writer
	err = s.snapshotter.DumpClose(ctx, tmpf)
	if err != nil {
		return fmt.Errorf("failed to dump metrics: %w", err)
	}

	stat, err := os.Stat(tmpfname)
	if err != nil {
		return fmt.Errorf("failed to stat temporary file: %w", err)
	}
	if stat.Size() == 0 {
		// no metrics were dumped - there were probably no writes since the last dump
		return nil
	}
	if err := os.Rename(tmpfname, s.config.MetricStoreFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary file (%s -> %s): %w", tmpfname, s.config.MetricStoreFilePath, err)
	}
	return nil
}

func (s *server) createPeriodicTask(f func(context.Context) error) periodictask.Task {
	var t periodictask.Task
	if s.config.MetricStoreInterval == 0 {
		taskFn := func(ctx context.Context, _ struct{}) error { return f(ctx) }
		t = periodictask.NewChanTask(s.snapshotter.C(), taskFn)
	} else {
		taskFn := func(ctx context.Context) error { return f(ctx) }
		t = periodictask.NewTimerTask(s.config.MetricStoreInterval, taskFn, s.config.MetricStoreInterval)
	}
	return t
}

func (s *server) launchNetworkServers(ctx context.Context, grp *errgroup.Group, rootListener net.Listener) {
	mux := cmux.New(rootListener)

	httpListener := mux.Match(cmux.HTTP1())
	grpcListener := mux.Match(cmux.Any()) // this will handle non-HTTP/1 protocols; for now, we have only gRPC.

	s.launchHTTPServer(ctx, grp, httpListener)
	s.launchGRPCServer(ctx, grp, grpcListener)

	// this goroutine ensures network connections are accepted and distributed to child listeners.
	grp.Go(func() error {
		err := mux.Serve()

		return s.filterCloseErrors(err)
	})
}

func (s *server) launchHTTPServer(ctx context.Context, grp *errgroup.Group, ln net.Listener) {
	s.logger.Info().Msg("launching HTTP server")

	grp.Go(func() error {
		return s.listenAndServeHTTP(ctx, ln)
	})
}

func (s *server) launchGRPCServer(ctx context.Context, grp *errgroup.Group, ln net.Listener) {
	s.logger.Info().Msg("launching GRPC server")

	grp.Go(func() error {
		return s.listenAndServeGRPC(ctx, ln)
	})
}

func (s *server) launchBatchWriter(ctx context.Context, grp *errgroup.Group) {
	s.logger.Info().Msg("launching storage batch writer")

	grp.Go(func() error {
		s.batchWriter.StartProcessing(ctx)

		return nil
	})
}

func (s *server) launchAuditProcessor(ctx context.Context, grp *errgroup.Group) {
	s.logger.Info().Msg("launching audit processor")

	grp.Go(func() error {
		s.auditProcessor.StartProcessing(ctx)

		return nil
	})
}

func (s *server) launchMetricDumper(ctx context.Context, grp *errgroup.Group) {
	s.logger.Info().Msg("launching metric dumper")

	grp.Go(func() error {
		return s.createPeriodicTask(s.dumpMetrics).Run(ctx)
	})
}

func (s *server) performFinalDump() error {
	if s.config.MetricStoreInterval == 0 {
		// metrics were being dumped on every write,
		// so nothing to dump here.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	return s.dumpMetrics(ctx)
}

func (s *server) filterCloseErrors(err error) error {
	if err == nil {
		return nil
	}

	// these errors usually happen on shutdown and are not that important
	// to return them to upstream caller.
	for _, ignored := range []error{
		cmux.ErrServerClosed,
		cmux.ErrListenerClosed,
		http.ErrServerClosed,
		net.ErrClosed,
	} {
		if errors.Is(err, ignored) {
			return nil
		}
	}

	return err
}
