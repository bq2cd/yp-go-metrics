package e2e_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage"
	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
	"github.com/bq2cd/yp-go-metrics/test/e2e"
)

var (
	tempFiles *servertest.TempFileFactory
)

type E2ECase struct {
	timeout       time.Duration // run agent+server pair for this duration
	agentEnv      map[string]string
	agentGRPC     bool
	serverEnv     map[string]string
	wantErrString string
}

func TestE2E(t *testing.T) {
	skipIfNotE2E(t)

	tempFiles = servertest.NewTempFileFactory(t)
	t.Cleanup(tempFiles.RemoveAll)

	tests := map[string]func(*testing.T){
		"default settings":                                                    testDefaultSettings,
		"poll and report metrics every second":                                testPollReportEverySecond,
		"report metrics with HMAC signature":                                  testReportWithHMACSignature,
		"report metrics with asymmetric encryption":                           testReportWithEncryption,
		"report metrics with asymmetric encryption and HMAC signature":        testReportWithEncryptionAndHMACSignature,
		"report metrics with trusted subnet":                                  testReportWithTrustedSubnet,
		"report metrics with trusted subnet and hash ignored":                 testReportWithTrustedSubnetHashIgnored,
		"report metrics with trusted subnet and correct hash":                 testReportWithTrustedSubnetAndCorrectHash,
		"report metrics via GRPC with trusted subnet":                         testReportViaGRPCWithTrustedSubnet,
		"report metrics via GRPC with trusted subnet and correct hash":        testReportViaGRPCWithTrustedSubnetAndCorrectHash,
		"no metrics reported with incorrect encryption configuration":         testNoReportWithIncorrectEncryptionConfig,
		"no metrics reported with invalid HMAC signature":                     testNoReportWithInvalidHMACSignature,
		"no metrics reported with different trusted subnet":                   testNoReportWithDifferentTrustedSubnet,
		"no metrics reported with trusted subnet but without hash":            testNoReportWithTrustedSubnetButWithoutHash,
		"no metrics reported via GRPC with different trusted subnet":          testNoReportViaGRPCWithDifferentTrustedSubnet,
		"no metrics reported via GRPC with trusted subnet and incorrect hash": testNoReportViaGRPCWithTrustedSubnetAndIncorrectHash,
	}

	for tname, tfunc := range tests {
		t.Run(tname, tfunc)
	}
}

func testDefaultSettings(t *testing.T) {
	t.Parallel()

	tc := E2ECase{
		timeout: 15 * time.Second, // default report interval is 10 seconds
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testPollReportEverySecond(t *testing.T) {
	t.Parallel()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportWithHMACSignature(t *testing.T) {
	t.Parallel()

	hmacKey := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"KEY":             hmacKey,
		},
		serverEnv: map[string]string{
			"KEY": hmacKey,
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportWithEncryption(t *testing.T) {
	t.Parallel()

	keyFiles := generateKeyPairFiles(t)

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"CRYPTO_KEY":      keyFiles.Public,
		},
		serverEnv: map[string]string{
			"CRYPTO_KEY": keyFiles.Private,
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportWithEncryptionAndHMACSignature(t *testing.T) {
	t.Parallel()

	keyFiles := generateKeyPairFiles(t)

	hmacKey := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"CRYPTO_KEY":      keyFiles.Public,
			"KEY":             hmacKey,
		},
		serverEnv: map[string]string{
			"CRYPTO_KEY": keyFiles.Private,
			"KEY":        hmacKey,
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportWithTrustedSubnet(t *testing.T) {
	t.Parallel()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
		},
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "127.0.0.0/8", // X-Real-IP is 127.0.0.1
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportWithTrustedSubnetHashIgnored(t *testing.T) {
	t.Parallel()

	hmacKey := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"KEY":             hmacKey,
		},
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "127.0.0.0/8", // X-Real-IP is 127.0.0.1
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportWithTrustedSubnetAndCorrectHash(t *testing.T) {
	t.Parallel()

	hmacKey := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"KEY":             hmacKey,
		},
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "127.0.0.0/8", // X-Real-IP is 127.0.0.1
			"KEY":            hmacKey,
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportViaGRPCWithTrustedSubnet(t *testing.T) {
	t.Parallel()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
		},
		agentGRPC: true,
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "127.0.0.0/8", // X-Real-IP is 127.0.0.1
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testReportViaGRPCWithTrustedSubnetAndCorrectHash(t *testing.T) {
	t.Parallel()

	hmacKey := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"KEY":             hmacKey,
		},
		agentGRPC: true,
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "127.0.0.0/8", // X-Real-IP is 127.0.0.1
			"KEY":            hmacKey,
		},
	}

	runE2ECase(t, tc, assertCollectedMetricsNotZero)
}

func testNoReportWithIncorrectEncryptionConfig(t *testing.T) {
	t.Parallel()

	keyFiles1 := generateKeyPairFiles(t)
	keyFiles2 := generateKeyPairFiles(t)

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"CRYPTO_KEY":      keyFiles1.Public,
		},
		serverEnv: map[string]string{
			"CRYPTO_KEY": keyFiles2.Private,
		},
		wantErrString: "exit status 1",
	}

	runE2ECase(t, tc, assertMetricsNotCollected)
}

func testNoReportWithInvalidHMACSignature(t *testing.T) {
	t.Parallel()

	hmacKey1 := servertest.GenerateHMACKeyBase64()
	hmacKey2 := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"KEY":             hmacKey1,
		},
		serverEnv: map[string]string{
			"KEY": hmacKey2,
		},
		wantErrString: "exit status 1",
	}

	runE2ECase(t, tc, assertMetricsNotCollected)
}

func testNoReportWithDifferentTrustedSubnet(t *testing.T) {
	t.Parallel()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
		},
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "10.0.0.0/24", // X-Real-IP is 127.0.0.1
		},
		wantErrString: "exit status 1",
	}

	runE2ECase(t, tc, assertMetricsNotCollected)
}

func testNoReportWithTrustedSubnetButWithoutHash(t *testing.T) {
	t.Parallel()

	hmacKey := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
		},
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "127.0.0.0/8", // X-Real-IP is 127.0.0.1
			"KEY":            hmacKey,
		},
		wantErrString: "exit status 1",
	}

	runE2ECase(t, tc, assertMetricsNotCollected)
}

func testNoReportViaGRPCWithDifferentTrustedSubnet(t *testing.T) {
	t.Parallel()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
		},
		agentGRPC: true,
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "10.0.0.0/24", // X-Real-IP is 127.0.0.1
		},
		wantErrString: "exit status 1",
	}

	runE2ECase(t, tc, assertMetricsNotCollected)
}

func testNoReportViaGRPCWithTrustedSubnetAndIncorrectHash(t *testing.T) {
	t.Parallel()

	hmacKey1 := servertest.GenerateHMACKeyBase64()
	hmacKey2 := servertest.GenerateHMACKeyBase64()

	tc := E2ECase{
		timeout: 5 * time.Second,
		agentEnv: map[string]string{
			"POLL_INTERVAL":   "1", // 1 second
			"REPORT_INTERVAL": "1", // 1 second
			"KEY":             hmacKey1,
		},
		agentGRPC: true,
		serverEnv: map[string]string{
			"TRUSTED_SUBNET": "127.0.0.0/8", // X-Real-IP is 127.0.0.1
			"KEY":            hmacKey2,
		},
		wantErrString: "exit status 1",
	}

	runE2ECase(t, tc, assertMetricsNotCollected)
}

func runE2ECase(t *testing.T, tc E2ECase, metricValidator func(*testing.T, []model.Metric)) {
	t.Helper()

	tempFactory := servertest.NewTempFileFactory(t)
	t.Cleanup(tempFactory.RemoveAll)

	if tc.serverEnv == nil {
		tc.serverEnv = make(map[string]string)
	}

	// ensure every test has independent file for dumping metrics
	tc.serverEnv["FILE_STORAGE_PATH"] = tempFactory.Create("metric-dump-*")

	launcher := e2e.NewLauncher(t, e2e.LauncherOpts{
		Timeout:      tc.timeout,
		AgentEnv:     tc.agentEnv,
		AgentOutput:  os.Stderr,
		AgentGRPC:    tc.agentGRPC,
		ServerEnv:    tc.serverEnv,
		ServerOutput: os.Stderr,
	})
	t.Cleanup(launcher.Cleanup)

	// Act
	err := launcher.Run()

	// Assert
	if tc.wantErrString == "" {
		require.NoError(t, err)
	} else {
		require.ErrorContains(t, err, tc.wantErrString)
	}

	metrics := retrieveMetricsFromDB(t, launcher.DBConfig())

	metricValidator(t, metrics)
}

func retrieveMetricsFromDB(t *testing.T, dbCfg dbconfig.Config) []model.Metric {
	t.Helper()

	storage, err := sqlstorage.New(dbCfg, e2e.NewNoopRetrierFactory())
	require.NoErrorf(t, err, "cannot connect to SQL database")

	metrics, err := storage.GetAll(t.Context())
	require.NoErrorf(t, err, "cannot retrieve metrics from SQL database")

	return metrics
}

func assertCollectedMetricsNotZero(t *testing.T, metrics []model.Metric) {
	t.Helper()

	allowedZeroValueMetricIDs := []string{
		"NumForcedGC",
		"NumGC",
		"LastGC",
		"Lookups",
		"PauseTotalNs",
		"GCCPUFraction",
	}

	metricSet := model.NewMetricSet(metrics...)

	allowedZeroValue := map[string]bool{}
	for _, mID := range allowedZeroValueMetricIDs {
		allowedZeroValue[mID] = true
	}

	for _, src := range source.DefaultSources() {
		for mID, mType := range src.AvailableMetricNames() {
			key := model.NewMetricKey(mType, mID)
			metric, ok := metricSet[key]

			if !assert.Truef(t, ok, "missing metric with key %v", key) {
				continue
			}

			switch metric.Type {
			case model.MetricTypeCounter:
				if assert.NotNilf(t, metric.Delta, "counter metric with nil delta (%#v)", metric) {
					if allowedZeroValue[metric.ID] {
						continue
					}
					assert.NotZerof(t, *metric.Delta, "expected non-zero counter metric delta (%s)", metricToJSON(t, metric))
				}
			case model.MetricTypeGauge:
				if assert.NotNilf(t, metric.Value, "gauge metric with nil value (%#v)", metric) {
					if allowedZeroValue[metric.ID] {
						continue
					}
					assert.NotZerof(t, *metric.Value, "expected non-zero gauge metric value (%s)", metricToJSON(t, metric))
				}
			default:
				assert.Failf(t, "unsupported metric type", "metric: %#v", metric)
			}
		}
	}
}

func assertMetricsNotCollected(t *testing.T, metrics []model.Metric) {
	t.Helper()

	assert.Emptyf(t, metrics, "no metrics must be collected")
}

func metricToJSON(t *testing.T, m model.Metric) string {
	t.Helper()

	out, err := json.Marshal(m)
	require.NoErrorf(t, err, "cannot marshal metric %#v to JSON", m)

	return string(out)
}

// keyPairFiles contains paths to temporary files containing X25519 key pair encoded with PEM format.
type keyPairFiles struct {
	Private string
	Public  string
}

func generateKeyPairFiles(t *testing.T) keyPairFiles {
	keypair, err := asymcrypt.NewX25519KeyPair()
	require.NoErrorf(t, err, "unable to generate X25519 key pair")

	privateKeyFile := tempFiles.Create("private-key-*")
	publicKeyFile := tempFiles.Create("public-key-*")

	err = os.WriteFile(privateKeyFile, keypair.Private, 0600)
	require.NoErrorf(t, err, "cannot write private key to file")

	os.WriteFile(publicKeyFile, keypair.Public, 0600)
	require.NoErrorf(t, err, "cannot write public key to file")

	return keyPairFiles{
		Private: privateKeyFile,
		Public:  publicKeyFile,
	}
}

func skipIfNotE2E(t *testing.T) {
	if os.Getenv("TEST_E2E") == "" {
		t.Skip("set TEST_E2E=1 to enable end-to-end tests")
	}
}
