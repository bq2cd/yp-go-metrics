package grpc

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pbmetrics "github.com/bq2cd/yp-go-metrics/api/gen/metrics/v1"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	"github.com/bq2cd/yp-go-metrics/internal/handler/grpc/converters"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/storagetest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/internal/service/servicetest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func Test_metrics_UpdateMetrics(t *testing.T) {
	type deps struct {
		storer           service.MetricStorer
		auditorNotCalled bool
	}
	type args struct {
		metrics []model.Metric
	}
	type want struct {
		code    codes.Code
		metrics []model.Metric
	}
	type testcase struct {
		deps deps
		args args
		want want
	}

	tests := map[string]testcase{
		"add counters to empty storage": {
			deps: deps{
				storer: newMetricStorer(t, storagetest.NewMockStorage()),
			},
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 7),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -5),
				},
			},
			want: want{
				code: codes.OK,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 7),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -5),
				},
			},
		},
		"add counters to non-empty storage with overlapping ids": {
			deps: deps{
				storer: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 10),
					model.NewCounterMetric("id2", -32),
					model.NewCounterMetric("id3", 11),
				)),
			},
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 7),
					model.NewCounterMetric("id3", -5),
					model.NewCounterMetric("id5", 12),
					model.NewCounterMetric("id7", 8),
					model.NewCounterMetric("id5", -19),
				},
			},
			want: want{
				code: codes.OK,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 17),
					model.NewCounterMetric("id3", 6),
					model.NewCounterMetric("id5", -7),
					model.NewCounterMetric("id7", 8),
				},
			},
		},
		"add mixed metrics to non-empty storage with overlapping ids": {
			deps: deps{
				storer: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id1", 10),
					model.NewCounterMetric("id2", -32),
					model.NewCounterMetric("id3", 11),
					model.NewGaugeMetric("id10", 8.3),
					model.NewGaugeMetric("id11", -5.6),
					model.NewGaugeMetric("id12", 0.032),
				)),
			},
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", -3),
					model.NewGaugeMetric("id10", -0.8),
					model.NewCounterMetric("id3", -5),
					model.NewGaugeMetric("id11", 7.11),
					model.NewCounterMetric("id5", 12),
					model.NewGaugeMetric("id10", 9.3),
					model.NewCounterMetric("id1", 5),
					model.NewGaugeMetric("id12", 99),
					model.NewCounterMetric("id7", 8),
					model.NewGaugeMetric("id13", -0.21),
					model.NewCounterMetric("id5", -19),
					model.NewGaugeMetric("id15", 8.345),
					model.NewCounterMetric("id1", 7),
				},
			},
			want: want{
				code: codes.OK,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 19),
					model.NewCounterMetric("id3", 6),
					model.NewCounterMetric("id5", -7),
					model.NewCounterMetric("id7", 8),
					model.NewGaugeMetric("id10", 9.3),
					model.NewGaugeMetric("id11", 7.11),
					model.NewGaugeMetric("id12", 99),
					model.NewGaugeMetric("id13", -0.21),
					model.NewGaugeMetric("id15", 8.345),
				},
			},
		},
		"empty request returns empty response": {
			deps: deps{
				storer: newMetricStorer(t, storagetest.NewMockStorage()),
			},
			args: args{
				metrics: []model.Metric{},
			},
			want: want{
				code:    codes.OK,
				metrics: []model.Metric{},
			},
		},
		"empty metrics are skipped": {
			deps: deps{
				storer: newMetricStorer(t, storagetest.NewMockStorage()),
			},
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 7),
					{},
					model.NewCounterMetric("id2", 10),
					{Type: model.MetricTypeCounter},
					model.NewCounterMetric("id3", -5),
					{Type: model.MetricTypeCounter, ID: "id4"},
				},
			},
			want: want{
				code: codes.OK,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 7),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -5),
					model.NewCounterMetric("id4", 0), // returns zero value due to metrics proto not supporting optional values
				},
			},
		},
		"empty metrics are skipped but pre-existing are returned": {
			deps: deps{
				storer: newMetricStorer(t, storagetest.NewMockStorage(
					model.NewCounterMetric("id4", 35),
				)),
			},
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 7),
					{},
					model.NewCounterMetric("id2", 10),
					{Type: model.MetricTypeCounter},
					model.NewCounterMetric("id3", -5),
					{Type: model.MetricTypeCounter, ID: "id4"},
				},
			},
			want: want{
				code: codes.OK,
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 7),
					model.NewCounterMetric("id2", 10),
					model.NewCounterMetric("id3", -5),
					model.NewCounterMetric("id4", 35),
				},
			},
		},
		"faulty storage": {
			deps: deps{
				storer:           newMetricStorer(t, storagetest.NewMockStorage().MakeFaulty()),
				auditorNotCalled: true,
			},
			args: args{
				metrics: []model.Metric{
					model.NewGaugeMetric(storagetest.FaultyStorageErrorTrigger, 0.05),
				},
			},
			want: want{
				code:    codes.Aborted,
				metrics: []model.Metric{},
			},
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := log.NewTestLogger()

			auditor := servicetest.NewMockMetricAuditor(ctrl)
			if !tc.deps.auditorNotCalled {
				// have to double convert to obtain the same metrics as are passed to the auditor;
				// this happens because metrics proto definition does not support optional values,
				// so invalid metrics get zero values during conversion.
				auditedMetrics := converters.ProtoToMetrics(converters.MetricsToProto(tc.args.metrics...))
				auditor.EXPECT().RecordMetricsUploaded(gomock.Any(), model.NewMetricSet(auditedMetrics...), model.ClientInfo{IPAddress: "127.0.0.1"})
			}

			handler := NewMetricsHandler(logger, tc.deps.storer, auditor)

			addr, stopGRPCServer := launchGRPCServer(t, handler)
			defer stopGRPCServer()

			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoErrorf(t, err, "cannot create GRPC connection")

			client := pbmetrics.NewMetricsClient(conn)

			req := new(pbmetrics.UpdateMetricsRequest)
			req.SetMetrics(converters.MetricsToProto(tc.args.metrics...))

			// Act
			resp, err := client.UpdateMetrics(t.Context(), req)

			// Assert
			st, _ := status.FromError(err)
			assert.Equalf(t, tc.want.code, st.Code(), "unexpected GRPC status code")

			gotMetrics := converters.ProtoToMetrics(resp.GetMetrics())
			assert.ElementsMatchf(t, tc.want.metrics, gotMetrics, "unexpected metrics returned by GRPC server")
		})
	}
}

func newMetricStorer(t *testing.T, storage *storagetest.MockStorage) service.MetricStorer {
	w := service.NewStorageBatchWriter(storage)

	go w.StartProcessing(t.Context()) // stops on context cancellation

	return service.NewMetricStorer(storage, w)
}

func launchGRPCServer(t *testing.T, handler *metricsHandler) (string, func() error) {
	addrFactory := servertest.NewListenAddressFactory(t)
	t.Cleanup(addrFactory.Clear)

	addr := addrFactory.New()

	ln, err := net.Listen("tcp", addr)
	require.NoErrorf(t, err, "cannot start GRPC listener")

	srv := grpc.NewServer()

	pbmetrics.RegisterMetricsServer(srv, handler)

	g := new(errgroup.Group)

	g.Go(func() error {
		return srv.Serve(ln)
	})

	return addr, func() error {
		srv.Stop()

		return g.Wait()
	}
}
