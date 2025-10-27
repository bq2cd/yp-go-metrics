package servertest

import (
	"testing"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestLaunchEmbeddedPostgres(t *testing.T) {
	type args struct {
		t        *testing.T
		user     string
		password string
		dbname   string
	}
	type want struct {
		got dbconfig.Config
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := LaunchEmbeddedPostgres(tt.args.t, tt.args.user, tt.args.password, tt.args.dbname)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_ensureNotEmpty(t *testing.T) {
	type args struct {
		v string
	}
	type want struct {
		got string
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := ensureNotEmpty(tt.args.v)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_createTemporaryDataDir(t *testing.T) {
	type args struct {
		t      *testing.T
		dbname string
	}
	type want struct {
		got string
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := createTemporaryDataDir(tt.args.t, tt.args.dbname)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
