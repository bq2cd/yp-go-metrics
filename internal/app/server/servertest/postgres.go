package servertest

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func ensureNotEmpty(v string) string {
	if v != "" {
		return v
	}
	return rand.Text()
}

func createTemporaryDataDir(t TestingT, dbname string) string {
	base := filepath.Join(os.TempDir(), "embedded-postgres-go")
	err := os.MkdirAll(base, 0755)
	if !os.IsExist(err) {
		require.NoError(t, err)
	}

	tmpdir, err := os.MkdirTemp(base, fmt.Sprintf("db-%s-*", dbname))
	t.Logf("created temporary postgres data directory: %v", err)
	require.NoError(t, err)
	require.DirExists(t, tmpdir)

	t.Cleanup(func() {
		err := os.RemoveAll(tmpdir)
		t.Logf("temporary postgres data directory removed: %v", err)
	})

	return tmpdir
}

// LaunchEmbeddedPostgresWithDelay starts a temporary PostgreSQL instance with empty data directory after specified delay.
func LaunchEmbeddedPostgresWithDelay(t TestingT, user, password, dbname string, delay time.Duration) dbconfig.Config {
	addr := NewListenAddress(t, GetRandomListenAddress(t))
	time.AfterFunc(delay, func() { launchEmbeddedPostgres(t, user, password, dbname, addr) })
	return generateDBConfig(t, user, password, dbname, addr)
}

// LaunchEmbeddedPostgres starts a temporary PostgreSQL instance with empty data directory.
func LaunchEmbeddedPostgres(t TestingT, user, password, dbname string) dbconfig.Config {
	addr := NewListenAddress(t, GetRandomListenAddress(t))
	launchEmbeddedPostgres(t, user, password, dbname, addr)
	return generateDBConfig(t, user, password, dbname, addr)
}

func generateDBConfig(t TestingT, user, password, dbname string, addr ListenAddress) dbconfig.Config {
	t.Helper()

	dbCfg, err := dbconfig.New(url.URL{
		Scheme: "postgres",
		Host:   addr.String(),
		User:   url.UserPassword(user, password),
		Path:   "/" + dbname,
	})
	require.NoError(t, err)

	return dbCfg
}

func launchEmbeddedPostgres(t TestingT, user, password, dbname string, addr ListenAddress) {
	t.Helper()

	user = ensureNotEmpty(user)
	password = ensureNotEmpty(password)
	dbname = ensureNotEmpty(dbname)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tmpdir := createTemporaryDataDir(t, dbname)
	logger := bytes.NewBuffer(nil)
	postgres := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			BinariesPath(filepath.Join(homeDir, ".embedded-postgres-go/binaries/v17")).
			RuntimePath(tmpdir).
			Version(embeddedpostgres.V17).
			Port(addr.Port).
			Username(user).
			Password(password).
			Database(dbname).
			StartTimeout(5 * time.Second).
			Logger(logger),
	)

	start := time.Now()
	t.Logf("staring embedded postgres on %v ...", addr)
	err = postgres.Start()
	t.Logf("embedded postgres started after %v: %v", time.Since(start), err)
	if err != nil {
		t.Logf("POSTGRES LOG:\n%s", logger.String())
	}
	require.NoError(t, err)

	t.Cleanup(func() {
		t.Logf("stopping embedded postgres...")
		err := postgres.Stop()
		t.Logf("embedded postgres stopped: %v", err)
		require.NoError(t, err)
	})
}
