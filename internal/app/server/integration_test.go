package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	neturl "net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	logger "github.com/bq2cd/yp-go-metrics/internal/app/logger"
	"github.com/bq2cd/yp-go-metrics/internal/app/server"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler/handlertest"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage"
	"github.com/bq2cd/yp-go-metrics/internal/testutil"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest/mockretrierfactory"
)

func TestIntegration(t *testing.T) {
	testutil.SkipTestInGithubActions(t)

	suite.Run(t, new(IntegrationTestSuite))
}

type IntegrationTestSuite struct {
	suite.Suite

	ctrl        *gomock.Controller
	addrFactory *servertest.ListenAddressFactory
	tempFactory *servertest.TempFileFactory
	requester   *handlertest.Requester
	responses   sync.Map
}

type IntegrationCase struct {
	timeout       time.Duration
	warmupDelay   time.Duration
	config        config.Config
	assertErr     func(*testing.T, error)
	assertRunning func(*testing.T, string)
	assertStopped func(*testing.T, config.Config)
}

func (ts *IntegrationTestSuite) SetupTest() {
	ts.ctrl = gomock.NewController(ts.T())
	ts.addrFactory = servertest.NewListenAddressFactory(ts.T())
	ts.tempFactory = servertest.NewTempFileFactory(ts.T())
	ts.requester = handlertest.NewRequester(ts.T(), http.DefaultClient)
}

func (ts *IntegrationTestSuite) TearDownTest() {
	ts.ctrl.Finish()
	ts.tempFactory.RemoveAll()
	ts.responses.Clear()
}

func (ts *IntegrationTestSuite) TestDumpMetricsOnEveryWrite() {
	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			MetricStoreInterval: 0,
			MetricStoreFilePath: ts.tempFactory.Create("test-metrics-dump-*"),
		},
		assertErr: func(t *testing.T, err error) {
			assert.NoError(t, err)
		},
		assertRunning: func(t *testing.T, addr string) {
			_, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/update/", addr), handlertest.NewBodyDataFromMetric(t, model.NewCounterMetric("id1", 78)), true)
			require.NoError(t, err)
		},
		assertStopped: func(t *testing.T, cfg config.Config) {
			dump, err := os.ReadFile(cfg.MetricStoreFilePath)
			require.NoError(t, err)
			t.Logf("dumped metrics (%s): %s", cfg.MetricStoreFilePath, string(dump))
			assert.JSONEq(t, `[{"id": "id1", "type": "counter", "delta": 78}]`, string(dump))
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestLoadMetricsOnStartup() {
	metricsDump := ts.tempFactory.Create("test-metrics-dump-*")
	err := os.WriteFile(metricsDump, []byte(`[{"id": "id1", "type": "counter", "delta": 78}]`), 0600)
	ts.Require().NoError(err)

	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			MetricStoreInterval:      24 * time.Hour,
			MetricStoreLoadOnStartup: true,
			MetricStoreFilePath:      metricsDump,
		},
		assertRunning: func(t *testing.T, addr string) {
			resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/value/", addr), handlertest.NewBodyDataFromMetricKey(t, model.NewMetricKey(model.MetricTypeCounter, "id1")), true)
			require.NoError(t, err)
			resp.Body.AssertData([]byte(`{"id": "id1", "type": "counter", "delta": 78}`))
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestServerUsesPostgresWhenConfigured() {
	dbCfg := servertest.LaunchEmbeddedPostgres(ts.T(), "server-test-user", "server-test-password", "server-test-db")
	dbURL, err := neturl.Parse(dbCfg.DSN())
	ts.Require().NoError(err)

	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			MetricStoreInterval: 0,
			MetricStoreFilePath: ts.tempFactory.Create("test-postgres-metrics-dump-*"),
			DatabaseURL:         *dbURL,
		},
		assertRunning: func(t *testing.T, addr string) {
			// populate metrics
			for _, m := range []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id3", -456),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
				model.NewGaugeMetric("id30", 456),
				model.NewGaugeMetric("id10", -4.56),
				model.NewCounterMetric("max_u32", math.MaxUint32),
				model.NewCounterMetric("max_i64", math.MaxInt64),
				model.NewGaugeMetric("max_f32", math.MaxFloat32),
				model.NewGaugeMetric("max_fu32", float64(math.MaxUint32)),
				model.NewGaugeMetric("max_f64", math.MaxFloat64),
			} {
				resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/update/", addr), handlertest.NewBodyDataFromMetric(t, m), true)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.Status)
				ts.responses.Store(m.Key(), resp)
			}

			// wait
			time.Sleep(50 * time.Millisecond)

			// retrieve metrics
			for _, k := range []model.MetricKey{
				model.NewMetricKey(model.MetricTypeCounter, "id1"),
				model.NewMetricKey(model.MetricTypeCounter, "id2"),
				model.NewMetricKey(model.MetricTypeCounter, "id3"),
				model.NewMetricKey(model.MetricTypeGauge, "id10"),
				model.NewMetricKey(model.MetricTypeGauge, "id20"),
				model.NewMetricKey(model.MetricTypeGauge, "id30"),
				model.NewMetricKey(model.MetricTypeCounter, "max_u32"),
				model.NewMetricKey(model.MetricTypeCounter, "max_i64"),
				model.NewMetricKey(model.MetricTypeGauge, "max_f32"),
				model.NewMetricKey(model.MetricTypeGauge, "max_fu32"),
				model.NewMetricKey(model.MetricTypeGauge, "max_f64"),
			} {
				resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/value/", addr), handlertest.NewBodyDataFromMetricKey(t, k), true)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.Status)
				up, ok := ts.responses.Load(k)
				require.Truef(t, ok, "missing response for metric key %v", k)
				up.(*handlertest.Response).Body.AssertEqual(resp.Body)
			}
		},
		assertStopped: func(t *testing.T, cfg config.Config) {
			wantMetrics := []model.Metric{
				model.NewGaugeMetric("id10", -4.56),
				model.NewGaugeMetric("id20", -1.23),
				model.NewGaugeMetric("id30", 456),
				model.NewCounterMetric("id1", 246),
				model.NewCounterMetric("id2", -123),
				model.NewCounterMetric("id3", -456),
				model.NewCounterMetric("max_u32", math.MaxUint32),
				model.NewCounterMetric("max_i64", math.MaxInt64),
				model.NewGaugeMetric("max_f32", math.MaxFloat32),
				model.NewGaugeMetric("max_fu32", float64(math.MaxUint32)),
				model.NewGaugeMetric("max_f64", math.MaxFloat64),
			}

			// db check
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ts.ctrl)
			retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")
			storage, err := sqlstorage.New(dbCfg, retrierFactory)
			require.NoError(t, err)
			metrics, err := storage.GetAll(t.Context())
			require.NoError(t, err)
			assert.ElementsMatch(t, wantMetrics, metrics)

			// snapshot check
			f, err := os.Open(cfg.MetricStoreFilePath)
			require.NoError(t, err)
			var snapshotMetrics []model.Metric
			err = json.NewDecoder(f).Decode(&snapshotMetrics)
			require.NoError(t, err)
			assert.ElementsMatch(t, wantMetrics, snapshotMetrics)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestServerRetriesPostgresQueriesOnConnectionErrors() {
	dbCfg := servertest.LaunchEmbeddedPostgresWithDelay(ts.T(), "server-test-user", "server-test-password", "server-test-db", 100*time.Millisecond)
	dbURL, err := neturl.Parse(dbCfg.DSN())
	ts.Require().NoError(err)

	ic := IntegrationCase{
		timeout:     5_000 * time.Millisecond,
		warmupDelay: 4_000 * time.Millisecond,
		config: config.Config{
			MetricStoreInterval: 0,
			MetricStoreFilePath: ts.tempFactory.Create("test-postgres-metrics-dump-*"),
			DatabaseURL:         *dbURL,
		},
		assertRunning: func(t *testing.T, addr string) {
			// populate metrics
			for _, m := range []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			} {
				resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/update/", addr), handlertest.NewBodyDataFromMetric(t, m), true)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.Status)
				ts.responses.Store(m.Key(), resp)
			}

			// wait
			time.Sleep(50 * time.Millisecond)

			// retrieve metrics
			for _, k := range []model.MetricKey{
				model.NewMetricKey(model.MetricTypeCounter, "id1"),
				model.NewMetricKey(model.MetricTypeCounter, "id2"),
				model.NewMetricKey(model.MetricTypeGauge, "id10"),
				model.NewMetricKey(model.MetricTypeGauge, "id20"),
			} {
				resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/value/", addr), handlertest.NewBodyDataFromMetricKey(t, k), true)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.Status)
				up, ok := ts.responses.Load(k)
				require.Truef(t, ok, "missing response for metric key %v", k)
				up.(*handlertest.Response).Body.AssertEqual(resp.Body)
			}
		},
		assertStopped: func(t *testing.T, cfg config.Config) {
			wantMetrics := []model.Metric{
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
			}

			// db check
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ts.ctrl)
			retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")
			storage, err := sqlstorage.New(dbCfg, retrierFactory)
			require.NoError(t, err)
			metrics, err := storage.GetAll(t.Context())
			require.NoError(t, err)
			assert.ElementsMatch(t, wantMetrics, metrics)

			// snapshot check
			f, err := os.Open(cfg.MetricStoreFilePath)
			require.NoError(t, err)
			var snapshotMetrics []model.Metric
			err = json.NewDecoder(f).Decode(&snapshotMetrics)
			require.NoError(t, err)
			assert.ElementsMatch(t, wantMetrics, snapshotMetrics)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) runIntegrationCase(ic IntegrationCase) {
	ic.config.ListenAddress = ts.addrFactory.New()

	if ic.assertErr == nil {
		ic.assertErr = func(t *testing.T, err error) {
			assert.NoError(t, err)
		}
	}

	errCh := make(chan error, 1)

	ctx, cancel := context.WithTimeout(ts.T().Context(), ic.timeout)
	defer cancel()

	go func() {
		defer close(errCh)
		errCh <- server.Run(ctx, logger.NewDevelopment(), ic.config)
	}()

	time.Sleep(ic.warmupDelay)

	ic.assertRunning(ts.T(), ts.addrFactory.Get(0))

	errRun := <-errCh

	if ic.assertStopped != nil {
		ic.assertStopped(ts.T(), ic.config)
	}

	if ic.assertErr == nil {
		ts.NoError(errRun)
	} else {
		ic.assertErr(ts.T(), errRun)
	}
}
