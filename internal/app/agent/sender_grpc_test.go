package agent

import (
	"context"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pbmetrics "github.com/bq2cd/yp-go-metrics/api/gen/metrics/v1"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	"github.com/bq2cd/yp-go-metrics/internal/handler/grpc/converters"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type mockGRPCServer struct {
	pbmetrics.UnimplementedMetricsServer

	wantRealIP httpheaders.XRealIP
	calls      atomic.Int64
	uploaded   model.MetricSet
	mu         sync.RWMutex
}

func (m *mockGRPCServer) UpdateMetrics(ctx context.Context, in *pbmetrics.UpdateMetricsRequest) (*pbmetrics.UpdateMetricsResponse, error) {
	m.calls.Add(1)

	ipValues := metadata.ValueFromIncomingContext(ctx, strings.ToLower(httpheaders.HeaderKeyXRealIP))
	if len(ipValues) < 1 {
		return nil, status.Error(codes.FailedPrecondition, "missing x-real-ip metadata")
	}
	if len(ipValues) > 1 {
		return nil, status.Errorf(codes.FailedPrecondition, "too many x-real-ip metadata values: %v", ipValues)
	}

	realIP := httpheaders.GetXRealIPFromBytes([]byte(ipValues[0]))

	if !m.wantRealIP.Equal(realIP) {
		return nil, status.Errorf(codes.FailedPrecondition, "incorrect x-real-ip metadata: expected %v, got %v", m.wantRealIP, realIP)
	}

	m.mu.Lock()
	if m.uploaded == nil {
		m.uploaded = model.NewMetricSet()
	}
	for _, metric := range converters.ProtoToMetrics(in.GetMetrics()) {
		m.uploaded.Upsert(metric)
	}
	m.mu.Unlock()

	// send all uploaded metrics back
	resp := new(pbmetrics.UpdateMetricsResponse)
	m.mu.RLock()
	resp.SetMetrics(converters.MetricsToProto(m.uploaded.Values()...))
	m.mu.RUnlock()

	return resp, nil
}

func (m *mockGRPCServer) GetUploadedMetrics() model.MetricSet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := model.NewMetricSet()
	for _, m := range m.uploaded {
		out.Upsert(m.Copy()) // prevent accidental value/delta pointers modifications
	}

	return out
}

func Test_senderGRPC_SendBatch(t *testing.T) {
	defaultRealIP := httpheaders.XRealIP{
		IP: net.ParseIP("127.0.0.1"),
	}

	type testcase struct {
		realIP        httpheaders.XRealIP
		metrics       []model.Metric
		wantMetrics   model.MetricSet
		wantRealIP    httpheaders.XRealIP
		wantCalls     int64
		wantErrString string
	}

	tests := map[string]testcase{
		"empty metrics never hit the server": {
			realIP:        defaultRealIP,
			metrics:       []model.Metric{},
			wantMetrics:   model.NewMetricSet(),
			wantRealIP:    defaultRealIP,
			wantCalls:     0,
			wantErrString: "",
		},
		"empty or incomplete metrics never hit the server": {
			realIP: defaultRealIP,
			metrics: []model.Metric{
				{},
				{Type: model.MetricTypeCounter},
				{Type: model.MetricTypeGauge},
				{Type: model.MetricTypeCounter, ID: "id1"},
				{Type: model.MetricTypeGauge, ID: "id2"},
			},
			wantMetrics:   model.NewMetricSet(),
			wantRealIP:    defaultRealIP,
			wantCalls:     0,
			wantErrString: "",
		},
		"single counter is sent OK": {
			realIP: defaultRealIP,
			metrics: []model.Metric{
				model.NewCounterMetric("id1", 5),
			},
			wantMetrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 5),
			),
			wantRealIP:    defaultRealIP,
			wantCalls:     1,
			wantErrString: "",
		},
		"single gauge is sent OK": {
			realIP: defaultRealIP,
			metrics: []model.Metric{
				model.NewGaugeMetric("id2", -0.5),
			},
			wantMetrics: model.NewMetricSet(
				model.NewGaugeMetric("id2", -0.5),
			),
			wantRealIP:    defaultRealIP,
			wantCalls:     1,
			wantErrString: "",
		},
		"multiple counters and gauges are sent OK": {
			realIP: defaultRealIP,
			metrics: []model.Metric{
				model.NewCounterMetric("id1", 5),
				model.NewCounterMetric("id2", -5),
				model.NewGaugeMetric("id3", 1.5),
				model.NewGaugeMetric("id4", -1.5),
			},
			wantMetrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 5),
				model.NewCounterMetric("id2", -5),
				model.NewGaugeMetric("id3", 1.5),
				model.NewGaugeMetric("id4", -1.5),
			),
			wantRealIP:    defaultRealIP,
			wantCalls:     1,
			wantErrString: "",
		},
		"when empty real IP is sent, error is returned": {
			realIP: httpheaders.XRealIP{},
			metrics: []model.Metric{
				model.NewCounterMetric("id1", 5),
			},
			wantMetrics:   nil,
			wantRealIP:    defaultRealIP,
			wantCalls:     1,
			wantErrString: "incorrect x-real-ip metadata",
		},
		"when incorrect real IP is sent, error is returned": {
			realIP: defaultRealIP,
			metrics: []model.Metric{
				model.NewCounterMetric("id1", 5),
			},
			wantMetrics:   nil,
			wantRealIP:    httpheaders.XRealIP{IP: net.ParseIP("10.0.0.1"), Hash: []byte(`ip-signature`)},
			wantCalls:     1,
			wantErrString: "incorrect x-real-ip metadata",
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			mockServer := &mockGRPCServer{
				wantRealIP: tc.wantRealIP,
			}

			addr, stopGRPCServer := launchMockGRPCServer(t, mockServer)

			time.Sleep(20 * time.Millisecond) // warmup delay for server

			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoErrorf(t, err, "cannot create GRPC dialer")

			logger := log.NewTestLogger()
			client := pbmetrics.NewMetricsClient(conn)

			sender := NewSenderGRPC(logger, client, tc.realIP)

			// Act
			sent, err := sender.SendBatch(t.Context(), model.NewMetricSet(tc.metrics...))

			// Assert
			if tc.wantErrString == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErrString)
			}

			assert.Equalf(t, tc.wantMetrics, sent, "expected sent metrics to mirror original metrics")

			err = stopGRPCServer()
			require.NoErrorf(t, err, "expected clean stop of GRPC server")

			assert.Equalf(t, tc.wantCalls, mockServer.calls.Load(), "unexpected mock server calls")
		})
	}
}

func Test_senderGRPC_Send(t *testing.T) {
	defaultRealIP := httpheaders.XRealIP{
		IP: net.ParseIP("127.0.0.1"),
	}

	type testcase struct {
		realIP        httpheaders.XRealIP
		metric        model.Metric
		wantMetrics   model.MetricSet
		wantRealIP    httpheaders.XRealIP
		wantCalls     int64
		wantErrString string
	}

	tests := map[string]testcase{
		"empty metric never hit the server": {
			realIP:        defaultRealIP,
			metric:        model.Metric{},
			wantMetrics:   model.NewMetricSet(),
			wantRealIP:    defaultRealIP,
			wantCalls:     0,
			wantErrString: "empty metric",
		},
		"counter is sent OK": {
			realIP: defaultRealIP,
			metric: model.NewCounterMetric("id1", 5),
			wantMetrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 5),
			),
			wantRealIP:    defaultRealIP,
			wantCalls:     1,
			wantErrString: "",
		},
		"gauge is sent OK": {
			realIP: defaultRealIP,
			metric: model.NewGaugeMetric("id2", -0.5),
			wantMetrics: model.NewMetricSet(
				model.NewGaugeMetric("id2", -0.5),
			),
			wantRealIP:    defaultRealIP,
			wantCalls:     1,
			wantErrString: "",
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			mockServer := &mockGRPCServer{
				wantRealIP: tc.wantRealIP,
			}

			addr, stopGRPCServer := launchMockGRPCServer(t, mockServer)

			time.Sleep(20 * time.Millisecond) // warmup delay for server

			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoErrorf(t, err, "cannot create GRPC dialer")

			logger := log.NewTestLogger()
			client := pbmetrics.NewMetricsClient(conn)

			sender := NewSenderGRPC(logger, client, tc.realIP)

			// Act
			err = sender.Send(t.Context(), tc.metric)

			// Assert
			if tc.wantErrString == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErrString)
			}

			assert.Equalf(t, tc.wantMetrics, mockServer.GetUploadedMetrics(), "expected sent metrics to mirror original metrics")

			err = stopGRPCServer()
			require.NoErrorf(t, err, "expected clean stop of GRPC server")

			assert.Equalf(t, tc.wantCalls, mockServer.calls.Load(), "unexpected mock server calls")
		})
	}
}

func launchMockGRPCServer(t *testing.T, mockServer *mockGRPCServer) (string, func() error) {
	addrFactory := servertest.NewListenAddressFactory(t)
	t.Cleanup(addrFactory.Clear)

	addr := addrFactory.New()

	ln, err := net.Listen("tcp", addr)
	require.NoErrorf(t, err, "cannot start GRPC listener")

	srv := grpc.NewServer()

	pbmetrics.RegisterMetricsServer(srv, mockServer)

	g := new(errgroup.Group)

	g.Go(func() error {
		return srv.Serve(ln)
	})

	return addr, func() error {
		srv.Stop()

		return g.Wait()
	}
}
