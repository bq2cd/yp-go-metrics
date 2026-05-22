package server_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	logger "github.com/bq2cd/yp-go-metrics/internal/app/logger"
	"github.com/bq2cd/yp-go-metrics/internal/app/server"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler/handlertest"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage"
	"github.com/bq2cd/yp-go-metrics/internal/testutil"
	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
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
	keypair     *asymcrypt.X25519KeyPair
}

type IntegrationCase struct {
	timeout       time.Duration
	warmupDelay   time.Duration
	config        config.Config
	assertErr     func(error)
	assertRunning func(string)
	assertStopped func(config.Config)
}

func (ts *IntegrationTestSuite) SetupSuite() {
	var err error

	ts.keypair, err = asymcrypt.NewX25519KeyPair()
	ts.Require().NoErrorf(err, "unable to generate X25519 key pair")
}

func (ts *IntegrationTestSuite) SetupTest() {
	ts.ctrl = gomock.NewController(ts.T())
	ts.addrFactory = servertest.NewListenAddressFactory(ts.T())
	ts.tempFactory = servertest.NewTempFileFactory(ts.T())
	ts.requester = handlertest.NewRequester(ts.T(), http.DefaultClient)
}

func (ts *IntegrationTestSuite) TearDownTest() {
	ts.ctrl.Finish()
	ts.addrFactory.Clear()
	ts.tempFactory.RemoveAll()
}

func (ts *IntegrationTestSuite) TestDumpMetricsOnEveryWrite() {
	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			MetricStoreFilePath: ts.tempFactory.Create("test-metrics-dump-*"),
		},
		assertRunning: func(addr string) {
			ts.uploadAndValidateMetrics(addr, []model.Metric{model.NewCounterMetric("id1", 78)}, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			ts.validateMetricsInSnapshot(cfg.MetricStoreFilePath, []model.Metric{model.NewCounterMetric("id1", 78)})
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
		assertRunning: func(addr string) {
			resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/value/", addr), handlertest.NewBodyDataFromMetricKey(ts.T(), model.NewMetricKey(model.MetricTypeCounter, "id1")), true)
			ts.Require().NoError(err)
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
			MetricStoreFilePath: ts.tempFactory.Create("test-postgres-metrics-dump-*"),
			DatabaseURL:         *dbURL,
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
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
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
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

			ts.validateMetricsInPostgres(dbCfg, wantMetrics)

			ts.validateMetricsInSnapshot(cfg.MetricStoreFilePath, wantMetrics)
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
		warmupDelay: 4_500 * time.Millisecond,
		config: config.Config{
			MetricStoreFilePath: ts.tempFactory.Create("test-postgres-metrics-dump-*"),
			DatabaseURL:         *dbURL,
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			wantMetrics := []model.Metric{
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
			}

			ts.validateMetricsInPostgres(dbCfg, wantMetrics)

			ts.validateMetricsInSnapshot(cfg.MetricStoreFilePath, wantMetrics)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestAuditFileSink() {
	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			AuditFilePath: ts.tempFactory.Create("test-metrics-audit-*"),
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			wantMetrics := []model.Metric{
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
			}

			ts.validateAuditFileSink(cfg.AuditFilePath, wantMetrics)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestAuditHTTPSink() {
	auditHandler := new(mockAuditReceiver)
	auditSrv := httptest.NewServer(auditHandler)
	auditURL, err := neturl.Parse(auditSrv.URL)

	ts.Require().NoError(err)

	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			AuditURL: *auditURL,
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			wantMetrics := []model.Metric{
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
			}

			ts.validateAuditHTTPSink(auditHandler, wantMetrics)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestAuditHTTPSinkNotAcceptingEvents() {
	auditHandler := new(mockAuditReceiver)
	auditHandler.SetFaulty()

	auditSrv := httptest.NewServer(auditHandler)
	auditURL, err := neturl.Parse(auditSrv.URL)

	ts.Require().NoError(err)

	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			AuditURL: *auditURL,
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			ts.validateAuditHTTPSink(auditHandler, nil)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestAuditFileAndHTTPSinks() {
	auditHandler := new(mockAuditReceiver)
	auditSrv := httptest.NewServer(auditHandler)
	auditURL, err := neturl.Parse(auditSrv.URL)

	ts.Require().NoError(err)

	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			AuditURL:      *auditURL,
			AuditFilePath: ts.tempFactory.Create("test-metrics-audit-*"),
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			wantMetrics := []model.Metric{
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
			}

			ts.validateAuditFileSink(cfg.AuditFilePath, wantMetrics)

			ts.validateAuditHTTPSink(auditHandler, wantMetrics)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestAuditFileAndHTTPSinksWithHTTPSinkFaulty() {
	auditHandler := new(mockAuditReceiver)
	auditHandler.SetFaulty()

	auditSrv := httptest.NewServer(auditHandler)
	auditURL, err := neturl.Parse(auditSrv.URL)

	ts.Require().NoError(err)

	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			AuditURL:      *auditURL,
			AuditFilePath: ts.tempFactory.Create("test-metrics-audit-*"),
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			wantMetrics := []model.Metric{
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
			}

			ts.validateAuditFileSink(cfg.AuditFilePath, wantMetrics)

			ts.validateAuditHTTPSink(auditHandler, nil)
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestEncryptedMetricsUpload() {
	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			DecryptionPrivateKey: ts.keypair.Private,
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			ts.uploadAndValidateMetrics(addr, collectedMetrics, ts.metricsToEncryptedBodyData, 50*time.Millisecond)
		},
		assertStopped: func(cfg config.Config) {
			// all validation happens while the server is running;
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) TestClearTextMetricsUploadFailsWhenEncryptionEnabled() {
	ic := IntegrationCase{
		timeout:     500 * time.Millisecond,
		warmupDelay: 100 * time.Millisecond,
		config: config.Config{
			DecryptionPrivateKey: ts.keypair.Private,
		},
		assertRunning: func(addr string) {
			collectedMetrics := []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewGaugeMetric("id10", 1.23),
				model.NewGaugeMetric("id20", -1.23),
			}

			resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/updates/", addr), ts.metricsToBodyData(collectedMetrics), true)

			ts.Require().NoError(err)
			ts.Equal(http.StatusBadRequest, resp.Status)
		},
		assertStopped: func(cfg config.Config) {
			// all validation happens while the server is running;
		},
	}

	ts.runIntegrationCase(ic)
}

func (ts *IntegrationTestSuite) runIntegrationCase(ic IntegrationCase) {
	addr := ts.addrFactory.New()

	ic.config.ListenAddress = addr

	errCh := make(chan error, 1)

	ctx, cancel := context.WithTimeout(ts.T().Context(), ic.timeout)
	defer cancel()

	go func() {
		defer close(errCh)
		errCh <- server.Run(ctx, logger.NewDevelopment(), ic.config)
	}()

	ts.waitUntilServerIsUp(ctx, addr, ic.warmupDelay)

	ic.assertRunning(addr)

	errRun := <-errCh

	if ic.assertStopped != nil {
		ic.assertStopped(ic.config)
	}

	if ic.assertErr == nil {
		ts.NoError(errRun)
	} else {
		ic.assertErr(errRun)
	}
}

func (ts *IntegrationTestSuite) waitUntilServerIsUp(baseCtx context.Context, addr string, timeout time.Duration) {
	ctx, cancel := context.WithTimeoutCause(baseCtx, timeout, fmt.Errorf("warmup time exceeded"))
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/", addr), http.NoBody)
			if err != nil {
				continue
			}

			err = servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
			if err == nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ts *IntegrationTestSuite) uploadAndValidateMetrics(addr string, metrics []model.Metric, bodyDataEncoder func([]model.Metric) handlertest.BodyData, delay time.Duration) {
	ts.T().Helper()

	resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/updates/", addr), bodyDataEncoder(metrics), true)

	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, resp.Status)

	uploadedMetrics, err := handlertest.DecodeBodyDataAsJSON[[]model.Metric](&resp.Body)

	ts.Require().NoError(err)

	uploadedSet := model.NewMetricSet(uploadedMetrics...)

	// wait
	time.Sleep(delay)

	// validate metrics retrieval
	for key := range model.NewMetricSet(metrics...) {
		resp, err := ts.requester.Do(http.MethodPost, fmt.Sprintf("http://%s/value/", addr), handlertest.NewBodyDataFromMetricKey(ts.T(), key), true)

		ts.Require().NoError(err)
		ts.Equal(http.StatusOK, resp.Status)

		metric, err := handlertest.DecodeBodyDataAsJSON[model.Metric](&resp.Body)

		ts.Require().NoError(err)

		ts.Equal(uploadedSet[key], metric)
	}
}

func (ts *IntegrationTestSuite) metricsToBodyData(metrics []model.Metric) handlertest.BodyData {
	return handlertest.NewBodyDataFromMetrics(ts.T(), metrics)
}

func (ts *IntegrationTestSuite) metricsToEncryptedBodyData(metrics []model.Metric) handlertest.BodyData {
	return handlertest.NewBodyDataFromMetrics(ts.T(), metrics).TransformData(func(data []byte) []byte {
		pubkey, err := asymcrypt.ParsePublicKey(ts.keypair.Public)
		ts.Require().NoErrorf(err, "unable to parse pre-generated X25519 public key")

		encryptor, err := asymcrypt.NewEncryptor(pubkey)
		ts.Require().NoErrorf(err, "unable to initialize encryptor")

		ciphertext, err := encryptor.Encrypt(data)
		ts.Require().NoErrorf(err, "unable to encrypt clear text data")

		return ciphertext
	})
}

func (ts *IntegrationTestSuite) validateMetricsInPostgres(dbCfg dbconfig.Config, wantMetrics []model.Metric) {
	ts.T().Helper()

	retrierFactory := mockretrierfactory.NewMockRetrierFactory(ts.ctrl)
	retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")

	storage, err := sqlstorage.New(dbCfg, retrierFactory)
	ts.Require().NoError(err)

	metrics, err := storage.GetAll(ts.T().Context())
	ts.Require().NoError(err)

	ts.ElementsMatch(wantMetrics, metrics)
}

func (ts *IntegrationTestSuite) validateMetricsInSnapshot(snapshotPath string, wantMetrics []model.Metric) {
	ts.T().Helper()

	f, err := os.Open(snapshotPath)
	ts.Require().NoError(err)

	defer func() { _ = f.Close() }()

	var metrics []model.Metric

	err = json.NewDecoder(f).Decode(&metrics)
	ts.Require().NoError(err)

	ts.ElementsMatch(wantMetrics, metrics)
}

func (ts *IntegrationTestSuite) validateAuditFileSink(sinkPath string, wantMetrics []model.Metric) {
	ts.T().Helper()

	wantEvent := model.NewAuditEvent(model.NewMetricSet(wantMetrics...), "127.0.0.1")

	f, err := os.Open(sinkPath)
	ts.Require().NoError(err)

	defer func() { _ = f.Close() }()

	lines := []string{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	ts.Require().NoError(scanner.Err())
	ts.Require().Len(lines, 1)

	var event model.AuditEvent

	err = json.Unmarshal([]byte(lines[0]), &event)

	ts.Require().NoError(err)

	ts.validateAuditEvent(wantEvent, event)
}

func (ts *IntegrationTestSuite) validateAuditHTTPSink(receiver *mockAuditReceiver, wantMetrics []model.Metric) {
	ts.T().Helper()

	gotEvents := receiver.GetEvents()

	if len(wantMetrics) == 0 {
		ts.Require().Empty(gotEvents)

		return
	}

	wantEvent := model.NewAuditEvent(model.NewMetricSet(wantMetrics...), "127.0.0.1")

	ts.Require().Len(gotEvents, 1)

	ts.validateAuditEvent(wantEvent, gotEvents[0])
}

func (ts *IntegrationTestSuite) validateAuditEvent(wantEvent, gotEvent model.AuditEvent) {
	ts.T().Helper()

	ts.Greater(gotEvent.Timestamp, wantEvent.Timestamp-5) // 5 seconds difference should be enough
	ts.Equal(wantEvent.IPAddress, gotEvent.IPAddress)
	ts.ElementsMatch(wantEvent.MetricNames, gotEvent.MetricNames)
}

type mockAuditReceiver struct {
	mu     sync.RWMutex
	events []model.AuditEvent
	faulty bool
}

func (m *mockAuditReceiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.faulty {
		w.WriteHeader(http.StatusBadGateway)

		return
	}

	var event model.AuditEvent

	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	m.mu.Lock()
	m.events = append(m.events, event)
	m.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (m *mockAuditReceiver) GetEvents() []model.AuditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.events
}

func (m *mockAuditReceiver) SetFaulty() {
	m.faulty = true
}
