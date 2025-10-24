package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/periodictask"
	"github.com/bq2cd/yp-go-metrics/internal/server/servertest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRouter struct {
	mock.Mock
	timeout time.Duration
}

func (m *mockRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called()
	if m.timeout <= 0 {
		return
	}
	timer := time.NewTimer(m.timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-r.Context().Done():
	}
}

type faultyListenerFactory struct{}

func (f *faultyListenerFactory) Create(ctx context.Context, addr string) (net.Listener, error) {
	return &faultyListener{}, nil
}

type faultyListener struct{}

func (l *faultyListener) Accept() (net.Conn, error) {
	return nil, fmt.Errorf("forced accept error")
}
func (l *faultyListener) Close() error   { return nil }
func (l *faultyListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4zero, Port: 0} }

func newMockMetricSnapshotter() *mockMetricSnapshotter {
	return &mockMetricSnapshotter{
		notifyCh: make(chan struct{}, 1),
	}
}

type mockMetricSnapshotter struct {
	mock.Mock
	notifyCh chan struct{}
}

func (m *mockMetricSnapshotter) StoreSingle(ctx context.Context, metric model.Metric) error {
	m.Called(ctx, metric)
	select {
	case m.notifyCh <- struct{}{}:
	default:
	}
	return nil
}
func (m *mockMetricSnapshotter) StoreBatch(ctx context.Context, metrics []model.Metric) error {
	m.Called(ctx, metrics)
	select {
	case m.notifyCh <- struct{}{}:
	default:
	}
	return nil
}
func (m *mockMetricSnapshotter) RetrieveSingle(ctx context.Context, key model.MetricKey) (model.Metric, error) {
	m.Called(ctx, key)
	return model.Metric{}, nil
}
func (m *mockMetricSnapshotter) RetrieveBatch(ctx context.Context, keys []model.MetricKey) ([]model.Metric, error) {
	m.Called(ctx, keys)
	return model.MetricSet{}, nil
}
func (m *mockMetricSnapshotter) RetrieveAll(ctx context.Context) ([]model.Metric, error) {
	m.Called(ctx)
	return model.MetricSet{}, nil
}
func (m *mockMetricSnapshotter) DumpClose(ctx context.Context, w io.WriteCloser) error {
	m.Called(ctx, w)
	return nil
}
func (m *mockMetricSnapshotter) LoadClose(ctx context.Context, r io.ReadCloser) error {
	m.Called(ctx, r)
	return nil
}
func (m *mockMetricSnapshotter) C() <-chan struct{} {
	m.Called()
	return m.notifyCh
}

func createTempFile(t *testing.T, pattern string) string {
	f, err := os.CreateTemp("", pattern)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestNew(t *testing.T) {
	type args struct {
		logger      log.Logger
		ctx         context.Context
		cfg         config.Config
		router      http.Handler
		snapshotter service.MetricSnapshotter
	}
	tests := []struct {
		name      string
		args      args
		assertion func(assert.TestingT, args, *server)
	}{
		{
			name: "no args",
			args: args{},
			assertion: func(t assert.TestingT, args args, s *server) {
				assert.NotNil(t, s.logger)
				assert.Equal(t, log.NewNoopLogger(), s.logger)
				assert.Nil(t, s.context)
				assert.Equal(t, config.Config{}, s.config)
				assert.Nil(t, s.router)
				assert.Nil(t, s.snapshotter)
				assert.NotNil(t, s.lnFactory)
				assert.Implements(t, (*ListenerFactory)(nil), s.lnFactory)
			},
		},
		{
			name: "logger only",
			args: args{logger: log.NewTestLogger()},
			assertion: func(t assert.TestingT, args args, s *server) {
				assert.Implements(t, (*log.TestLogger)(nil), s.logger)
				assert.Equal(t, config.Config{}, s.config)
				assert.Nil(t, s.router)
				assert.Nil(t, s.snapshotter)
				assert.NotNil(t, s.lnFactory)
				assert.Implements(t, (*ListenerFactory)(nil), s.lnFactory)
			},
		},
		{
			name: "ctx only",
			args: args{ctx: context.Background()},
			assertion: func(t assert.TestingT, args args, s *server) {
				assert.Equal(t, args.ctx, s.context)
				assert.Equal(t, config.Config{}, s.config)
				assert.Nil(t, s.router)
				assert.Nil(t, s.snapshotter)
				assert.NotNil(t, s.lnFactory)
				assert.Implements(t, (*ListenerFactory)(nil), s.lnFactory)
			},
		},
		{
			name: "ctx + router",
			args: args{ctx: context.Background(), router: http.NewServeMux()},
			assertion: func(t assert.TestingT, args args, s *server) {
				assert.Equal(t, args.ctx, s.context)
				assert.Equal(t, config.Config{}, s.config)
				assert.Equal(t, args.router, s.router)
				assert.Nil(t, s.snapshotter)
				assert.NotNil(t, s.lnFactory)
				assert.Implements(t, (*ListenerFactory)(nil), s.lnFactory)
			},
		},
		{
			name: "ctx + router + snapshotter",
			args: args{ctx: context.Background(), router: http.NewServeMux(), snapshotter: newMockMetricSnapshotter()},
			assertion: func(t assert.TestingT, args args, s *server) {
				assert.Equal(t, args.ctx, s.context)
				assert.Equal(t, config.Config{}, s.config)
				assert.Equal(t, args.router, s.router)
				assert.Equal(t, args.snapshotter, s.snapshotter)
				assert.NotNil(t, s.lnFactory)
				assert.Implements(t, (*ListenerFactory)(nil), s.lnFactory)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, New(tt.args.ctx, tt.args.logger, tt.args.cfg, tt.args.router, tt.args.snapshotter))
		})
	}
}

func Test_server_Run(t *testing.T) {
	type fields struct {
		config      config.Config
		router      *mockRouter
		snapshotter *mockMetricSnapshotter
		lnFactory   ListenerFactory
	}
	type innerTest struct {
		name              string
		fields            fields
		assertRouter      func(*testing.T, *mockRouter, *http.Request, chan error)
		expectSnapshotter func(*testing.T, *mockMetricSnapshotter)
		assertSnapshotter func(*testing.T, *mockMetricSnapshotter)
	}
	tests := []struct {
		name  string
		cases []innerTest
	}{
		{
			name: "metric snapshotter not configured",
			cases: []innerTest{
				{
					name: "normal run",
					fields: fields{
						config: config.Config{
							ListenAddress:   servertest.GetRandomListenAddress(t),
							ShutdownTimeout: 100 * time.Millisecond},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.NoError(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan")).Once()
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "DumpClose")
						m.AssertNotCalled(t, "LoadClose")
					},
				},
				{
					name: "invalid listen address",
					fields: fields{
						config: config.Config{
							ListenAddress:   "localhost1:123:456",
							ShutdownTimeout: 100 * time.Millisecond},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.Error(t, err)
						m.AssertNotCalled(t, "ServeHTTP")
						require.Error(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan")).Once()
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "DumpClose")
						m.AssertNotCalled(t, "LoadClose")
					},
				},
				{
					name: "slow router",
					fields: fields{
						config: config.Config{
							ListenAddress:   servertest.GetRandomListenAddress(t),
							ShutdownTimeout: 200 * time.Millisecond},
						router:      &mockRouter{timeout: 2_000 * time.Millisecond},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						err = servertest.MakeRequestDiscardResponse(nil, req)
						require.Error(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.Error(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan")).Once()
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "DumpClose")
						m.AssertNotCalled(t, "LoadClose")
					},
				},
				{
					name: "listener error",
					fields: fields{
						config: config.Config{
							ListenAddress:   servertest.GetRandomListenAddress(t),
							ShutdownTimeout: 100 * time.Millisecond},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &faultyListenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.Error(t, err)
						m.AssertNotCalled(t, "ServeHTTP")
						require.Error(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan")).Once()
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "DumpClose")
						m.AssertNotCalled(t, "LoadClose")
					},
				},
			},
		},
		{
			name: "metric snapshotter runs on interval",
			cases: []innerTest{
				{
					name: "store path is empty",
					fields: fields{
						config: config.Config{
							ListenAddress:       servertest.GetRandomListenAddress(t),
							ShutdownTimeout:     100 * time.Millisecond,
							MetricStoreInterval: 10 * time.Millisecond,
						},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.NoError(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertNotCalled(t, "C")
						m.AssertNotCalled(t, "DumpClose")
						m.AssertNotCalled(t, "LoadClose")
					},
				},
				{
					name: "store path defined",
					fields: fields{
						config: config.Config{
							ListenAddress:       servertest.GetRandomListenAddress(t),
							ShutdownTimeout:     100 * time.Millisecond,
							MetricStoreInterval: 12 * time.Millisecond,
							MetricStoreFilePath: createTempFile(t, "test-metric-store-*"),
						},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.NoError(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("DumpClose", mock.Anything, mock.Anything).Return(nil).Times(9)
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "C")
						m.AssertNotCalled(t, "LoadClose")
					},
				},
			},
		},
		{
			name: "metric snapshotter runs on each update",
			cases: []innerTest{
				{
					name: "store path is empty",
					fields: fields{
						config: config.Config{
							ListenAddress:       servertest.GetRandomListenAddress(t),
							ShutdownTimeout:     100 * time.Millisecond,
							MetricStoreInterval: 0,
						},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.NoError(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan")).Once()
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "DumpClose")
						m.AssertNotCalled(t, "LoadClose")
					},
				},
				{
					name: "store path defined",
					fields: fields{
						config: config.Config{
							ListenAddress:       servertest.GetRandomListenAddress(t),
							ShutdownTimeout:     100 * time.Millisecond,
							MetricStoreInterval: 0,
							MetricStoreFilePath: createTempFile(t, "test-metric-store-*"),
						},
						router: &mockRouter{},
						snapshotter: func() *mockMetricSnapshotter {
							m := newMockMetricSnapshotter()
							go func() {
								for range 8 {
									m.notifyCh <- struct{}{}
								}
							}()
							return m
						}(),
						lnFactory: &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.NoError(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan")).Once()
						m.On("DumpClose", mock.Anything, mock.Anything).Return(nil).Times(8)
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "LoadClose")
					},
				},
			},
		},
		{
			name: "metric loader runs on startup",
			cases: []innerTest{
				{
					name: "store path is empty",
					fields: fields{
						config: config.Config{
							ListenAddress:            servertest.GetRandomListenAddress(t),
							ShutdownTimeout:          100 * time.Millisecond,
							MetricStoreLoadOnStartup: true,
						},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.NoError(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan"))
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "LoadClose")
						m.AssertNotCalled(t, "DumpClose")
					},
				},
				{
					name: "store path defined",
					fields: fields{
						config: config.Config{
							ListenAddress:            servertest.GetRandomListenAddress(t),
							ShutdownTimeout:          100 * time.Millisecond,
							MetricStoreLoadOnStartup: true,
							MetricStoreFilePath:      createTempFile(t, "test-metric-store-*"),
						},
						router:      &mockRouter{},
						snapshotter: newMockMetricSnapshotter(),
						lnFactory:   &listenerFactory{},
					},
					assertRouter: func(t *testing.T, m *mockRouter, req *http.Request, errCh chan error) {
						err := servertest.MakeRequestDiscardResponse(nil, req)
						require.NoError(t, err)
						m.AssertExpectations(t)
						m.AssertNumberOfCalls(t, "ServeHTTP", 1)
						require.NoError(t, <-errCh)
					},
					expectSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.On("C").Return(mock.AnythingOfType("chan"))
						m.On("LoadClose", mock.Anything, mock.Anything).Return(nil)
					},
					assertSnapshotter: func(t *testing.T, m *mockMetricSnapshotter) {
						m.AssertExpectations(t)
						m.AssertNotCalled(t, "DumpClose")
					},
				},
			},
		},
	}
	for _, outer := range tests {
		t.Run(outer.name, func(t *testing.T) {
			for _, tt := range outer.cases {
				t.Run(tt.name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), tt.fields.config.ShutdownTimeout)
					defer cancel()

					if tt.fields.config.MetricStoreFilePath != "" {
						defer func() { _ = os.Remove(tt.fields.config.MetricStoreFilePath) }()
					}

					s := &server{
						logger:      log.NewNoopLogger(),
						context:     ctx,
						config:      tt.fields.config,
						router:      tt.fields.router,
						snapshotter: tt.fields.snapshotter,
						lnFactory:   tt.fields.lnFactory,
					}

					tt.fields.router.On("ServeHTTP", mock.Anything, mock.Anything).Return()
					tt.expectSnapshotter(t, tt.fields.snapshotter)

					errCh := make(chan error, 1)
					go func() {
						errCh <- s.Run()
					}()

					time.Sleep(10 * time.Millisecond)

					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", s.config.ListenAddress), http.NoBody)
					require.NoError(t, err)

					tt.assertRouter(t, tt.fields.router, req, errCh)
					tt.assertSnapshotter(t, tt.fields.snapshotter)
				})
			}
		})
	}
}

func Test_listenerFactory_Create(t *testing.T) {
	type args struct {
		ctx  context.Context
		addr string
	}
	tests := []struct {
		name      string
		f         *listenerFactory
		args      args
		want      net.Listener
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &listenerFactory{}
			got, err := f.Create(tt.args.ctx, tt.args.addr)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_server_listenAndServe(t *testing.T) {
	type fields struct {
		logger      log.Logger
		context     context.Context
		config      config.Config
		router      http.Handler
		snapshotter service.MetricSnapshotter
		lnFactory   ListenerFactory
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				logger:      tt.fields.logger,
				context:     tt.fields.context,
				config:      tt.fields.config,
				router:      tt.fields.router,
				snapshotter: tt.fields.snapshotter,
				lnFactory:   tt.fields.lnFactory,
			}
			tt.assertion(t, s.listenAndServe())
		})
	}
}

func Test_server_dumpMetrics(t *testing.T) {
	type fields struct {
		logger      log.Logger
		context     context.Context
		config      config.Config
		router      http.Handler
		snapshotter service.MetricSnapshotter
		lnFactory   ListenerFactory
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				logger:      tt.fields.logger,
				context:     tt.fields.context,
				config:      tt.fields.config,
				router:      tt.fields.router,
				snapshotter: tt.fields.snapshotter,
				lnFactory:   tt.fields.lnFactory,
			}
			tt.assertion(t, s.dumpMetrics())
		})
	}
}

func Test_server_createPeriodicTask(t *testing.T) {
	type fields struct {
		logger      log.Logger
		context     context.Context
		config      config.Config
		router      http.Handler
		snapshotter service.MetricSnapshotter
		lnFactory   ListenerFactory
	}
	type args struct {
		f func() error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   periodictask.Task
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				logger:      tt.fields.logger,
				context:     tt.fields.context,
				config:      tt.fields.config,
				router:      tt.fields.router,
				snapshotter: tt.fields.snapshotter,
				lnFactory:   tt.fields.lnFactory,
			}
			assert.Equal(t, tt.want, s.createPeriodicTask(tt.args.f))
		})
	}
}

func Test_server_tryLoadMetrics(t *testing.T) {
	type fields struct {
		logger      log.Logger
		context     context.Context
		config      config.Config
		router      http.Handler
		snapshotter service.MetricSnapshotter
		lnFactory   ListenerFactory
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				logger:      tt.fields.logger,
				context:     tt.fields.context,
				config:      tt.fields.config,
				router:      tt.fields.router,
				snapshotter: tt.fields.snapshotter,
				lnFactory:   tt.fields.lnFactory,
			}
			s.tryLoadMetrics()
		})
	}
}
