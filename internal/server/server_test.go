package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/bq2cd/yp-go-metrics/internal/server/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestNewServer(t *testing.T) {
	type args struct {
		logger log.Logger
		ctx    context.Context
		cfg    config.Config
		router http.Handler
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
				assert.NotNil(t, s.lnFactory)
				assert.Implements(t, (*ListenerFactory)(nil), s.lnFactory)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewServer(tt.args.logger, tt.args.ctx, tt.args.cfg, tt.args.router))
		})
	}
}

func Test_server_Run(t *testing.T) {
	type fields struct {
		config    config.Config
		router    *mockRouter
		lnFactory ListenerFactory
	}
	tests := []struct {
		name      string
		fields    fields
		assertion func(*testing.T, *mock.Call, *http.Request, chan error)
	}{
		{
			name: "normal run",
			fields: fields{
				config: config.Config{
					ListenAddress:   servertest.GetRandomListenAddress(t),
					ShutdownTimeout: 100 * time.Millisecond},
				router:    &mockRouter{},
				lnFactory: &listenerFactory{},
			},
			assertion: func(t *testing.T, m *mock.Call, req *http.Request, errCh chan error) {
				err := servertest.MakeRequestDiscardResponse(nil, req)
				assert.NoError(t, err)

				m.Parent.AssertExpectations(t)
				m.Parent.AssertNumberOfCalls(t, "ServeHTTP", 1)

				assert.NoError(t, <-errCh)
			},
		},
		{
			name: "invalid listen address",
			fields: fields{
				config: config.Config{
					ListenAddress:   "localhost1:123:456",
					ShutdownTimeout: 100 * time.Millisecond},
				router:    &mockRouter{},
				lnFactory: &listenerFactory{},
			},
			assertion: func(t *testing.T, m *mock.Call, req *http.Request, errCh chan error) {
				err := servertest.MakeRequestDiscardResponse(nil, req)
				assert.Error(t, err)

				m.Parent.AssertNotCalled(t, "ServeHTTP")

				assert.Error(t, <-errCh)
			},
		},
		{
			name: "slow router",
			fields: fields{
				config: config.Config{
					ListenAddress:   servertest.GetRandomListenAddress(t),
					ShutdownTimeout: 200 * time.Millisecond},
				router:    &mockRouter{timeout: 2_000 * time.Millisecond},
				lnFactory: &listenerFactory{},
			},
			assertion: func(t *testing.T, m *mock.Call, req *http.Request, errCh chan error) {
				err := servertest.MakeRequestDiscardResponse(nil, req)
				assert.NoError(t, err)

				err = servertest.MakeRequestDiscardResponse(nil, req)
				assert.Error(t, err)

				m.Parent.AssertExpectations(t)
				m.Parent.AssertNumberOfCalls(t, "ServeHTTP", 1)

				assert.Error(t, <-errCh)
			},
		},
		{
			name: "listener error",
			fields: fields{
				config: config.Config{
					ListenAddress:   servertest.GetRandomListenAddress(t),
					ShutdownTimeout: 100 * time.Millisecond},
				router:    &mockRouter{},
				lnFactory: &faultyListenerFactory{},
			},
			assertion: func(t *testing.T, m *mock.Call, req *http.Request, errCh chan error) {
				err := servertest.MakeRequestDiscardResponse(nil, req)
				assert.Error(t, err)

				m.Parent.AssertNotCalled(t, "ServeHTTP")

				assert.Error(t, <-errCh)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), tt.fields.config.ShutdownTimeout)
			defer cancel()

			s := &server{
				logger:    log.NewNoopLogger(),
				context:   ctx,
				config:    tt.fields.config,
				router:    tt.fields.router,
				lnFactory: tt.fields.lnFactory,
			}

			m := tt.fields.router.On("ServeHTTP", mock.Anything, mock.Anything).Return()

			errCh := make(chan error, 1)
			go func() {
				errCh <- s.Run()
			}()

			time.Sleep(10 * time.Millisecond)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", s.config.ListenAddress), http.NoBody)
			assert.NoError(t, err)

			tt.assertion(t, m, req, errCh)
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
