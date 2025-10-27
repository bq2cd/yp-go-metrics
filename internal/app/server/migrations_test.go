package server

import (
	"database/sql"
	"os"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMigrationsDBName     = "migrations-db"
	testMigrationsDBUser     = "migrations-user"
	testMigrationsDBPassword = "migrations-password"
)

func getDBConnectionEmbedded(t *testing.T) (*sql.DB, dbconfig.Config) {
	t.Helper()

	dbCfg := servertest.LaunchEmbeddedPostgres(t, testMigrationsDBUser, testMigrationsDBPassword, testMigrationsDBName)

	db, err := sql.Open(string(dbCfg.Driver()), dbCfg.DSN())
	require.NoError(t, err)

	return db, dbCfg
}

func Test_applyMigrations(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("needs INTEGRATION=1 env var")
	}
	db, dbCfg := getDBConnectionEmbedded(t)

	type want struct {
		wantErr  bool
		assertDB func(*testing.T, *sql.DB)
	}
	type testcase struct {
		want want
	}
	tests := map[string]testcase{
		"migrations succeed": {
			want: want{
				assertDB: func(t *testing.T, db *sql.DB) {
					// counters
					{
						res, err := db.Exec(`INSERT INTO
							metrics_counter(metric_id, value)
							VALUES
								('id1', 123),
								('id2', 456),
								('id3', -789)
						`)
						require.NoError(t, err)
						n, err := res.RowsAffected()
						require.NoError(t, err)
						assert.Equal(t, 3, int(n))
					}

					// gauges
					{
						res, err := db.Exec(`INSERT INTO
							metrics_gauge(metric_id, value)
							VALUES
								('id1', 1.23),
								('id2', 456),
								('id3', -7.89),
								('id4', -123)
						`)
						require.NoError(t, err)
						n, err := res.RowsAffected()
						require.NoError(t, err)
						assert.Equal(t, 4, int(n))
					}
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			logger := log.NewTestLogger()
			err := applyMigrations(t.Context(), logger, dbCfg)
			if tt.want.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.want.assertDB(t, db)
		})
	}
}

func Test_gooseLogger_Fatalf(t *testing.T) {
	type fields struct {
		logger log.Logger
	}
	type args struct {
		format string
		v      []any
	}
	type want struct {
	}
	type testcase struct {
		fields fields
		args   args
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := &gooseLogger{
				logger: tt.fields.logger,
			}
			l.Fatalf(tt.args.format, tt.args.v...)
		})
	}
}

func Test_gooseLogger_Printf(t *testing.T) {
	type fields struct {
		logger log.Logger
	}
	type args struct {
		format string
		v      []any
	}
	type want struct {
	}
	type testcase struct {
		fields fields
		args   args
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := &gooseLogger{
				logger: tt.fields.logger,
			}
			l.Printf(tt.args.format, tt.args.v...)
		})
	}
}
