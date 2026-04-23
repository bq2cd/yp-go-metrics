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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
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
	return []model.Metric{}, nil
}
func (m *mockMetricSnapshotter) RetrieveAll(ctx context.Context) ([]model.Metric, error) {
	m.Called(ctx)
	return []model.Metric{}, nil
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

type mockStorageBatchWriter struct {
	mock.Mock
}

func newMockStorageBatchWriter() *mockStorageBatchWriter {
	return &mockStorageBatchWriter{}
}

func (m *mockStorageBatchWriter) WriteBatch(ctx context.Context, batch service.MetricBatch) service.MetricBatchTx {
	m.Called(ctx, batch)
	return nil
}

func (m *mockStorageBatchWriter) StartProcessing(ctx context.Context) {
	m.Called(ctx)
}

type mockAuditEventProcessor struct {
	mock.Mock
}

func newMockAuditEventProcessor() *mockAuditEventProcessor {
	return &mockAuditEventProcessor{}
}

func (m *mockAuditEventProcessor) StartProcessing(ctx context.Context) {
	m.Called(ctx)
}

func (m *mockAuditEventProcessor) RegisterSink(sinkID string, sink repository.AuditSink) {
	m.Called(sinkID, sink)
}

func (m *mockAuditEventProcessor) WriteEvent(ctx context.Context, event model.AuditEvent) error {
	m.Called(ctx, event)
	return nil
}

func (m *mockAuditEventProcessor) Close() error {
	m.Called()
	return nil
}

func createTempFile(t *testing.T, pattern string) string {
	f, err := os.CreateTemp("", pattern)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func Test_server_Run(t *testing.T) {
	type fields struct {
		config         config.Config
		router         *mockRouter
		snapshotter    *mockMetricSnapshotter
		batchWriter    *mockStorageBatchWriter
		auditProcessor *mockAuditEventProcessor
		lnFactory      ListenerFactory
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						router:         &mockRouter{timeout: 2_000 * time.Millisecond},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &faultyListenerFactory{},
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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
						router:         &mockRouter{},
						snapshotter:    newMockMetricSnapshotter(),
						batchWriter:    newMockStorageBatchWriter(),
						auditProcessor: newMockAuditEventProcessor(),
						lnFactory:      &listenerFactory{},
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

					s := New(log.NewNoopLogger(), tt.fields.config, tt.fields.router, tt.fields.snapshotter, tt.fields.batchWriter, tt.fields.auditProcessor)
					s.lnFactory = tt.fields.lnFactory

					tt.fields.router.On("ServeHTTP", mock.Anything, mock.Anything).Return()
					tt.expectSnapshotter(t, tt.fields.snapshotter)
					tt.fields.batchWriter.On("StartProcessing", ctx).Return().Once()
					tt.fields.auditProcessor.On("StartProcessing", ctx).Return().Once()

					errCh := make(chan error, 1)
					go func() {
						errCh <- s.Run(ctx)
					}()

					time.Sleep(10 * time.Millisecond)

					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", s.config.ListenAddress), http.NoBody)
					require.NoError(t, err)

					tt.assertRouter(t, tt.fields.router, req, errCh)
					tt.assertSnapshotter(t, tt.fields.snapshotter)
					tt.fields.batchWriter.AssertExpectations(t)
				})
			}
		})
	}
}
