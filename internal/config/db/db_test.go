package db

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Enabled(t *testing.T) {
	type fields struct {
		driver Driver
		dsn    string
	}
	type want struct {
		got bool
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		"empty dsn returns false": {
			fields: fields{driver: DriverPgx, dsn: ""},
			want:   want{got: false},
		},
		"non-empty dsn returns true": {
			fields: fields{driver: DriverPgx, dsn: "postgres://"},
			want:   want{got: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := Config{
				driver: tt.fields.driver,
				dsn:    tt.fields.dsn,
			}
			got := c.Enabled()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestConfig_Driver(t *testing.T) {
	type fields struct {
		driver Driver
		dsn    string
	}
	type want struct {
		got Driver
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		"returns field value": {
			fields: fields{driver: DriverNone},
			want:   want{got: DriverNone},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := Config{
				driver: tt.fields.driver,
				dsn:    tt.fields.dsn,
			}
			got := c.Driver()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestConfig_DSN(t *testing.T) {
	type fields struct {
		driver Driver
		dsn    string
	}
	type want struct {
		got string
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		"returns field value": {
			fields: fields{dsn: "123"},
			want:   want{got: "123"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := Config{
				driver: tt.fields.driver,
				dsn:    tt.fields.dsn,
			}
			got := c.DSN()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		dbURL url.URL
	}
	type want struct {
		got Config
		err error
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty url returns empty dsn": {
			args: args{dbURL: url.URL{}},
			want: want{got: Config{driver: DriverPgx}, err: nil},
		},
		"unsupported db type returns error": {
			args: args{dbURL: url.URL{Scheme: "mysql"}},
			want: want{got: Config{}, err: ErrUnsupportedDBType},
		},
		"postgres url returns proper config": {
			args: args{dbURL: url.URL{Scheme: "postgres", Host: "localhost:1234"}},
			want: want{got: Config{driver: DriverPgx, dsn: "postgres://localhost:1234"}, err: nil},
		},
		"postgresql url returns proper config": {
			args: args{dbURL: url.URL{Scheme: "postgresql", Host: "localhost:1234"}},
			want: want{got: Config{driver: DriverPgx, dsn: "postgresql://localhost:1234"}, err: nil},
		},
		"user and password are recognised": {
			args: args{dbURL: url.URL{Scheme: "postgres", Host: "localhost:1234", User: url.UserPassword("user1", "password1")}},
			want: want{got: Config{driver: DriverPgx, dsn: "postgres://user1:password1@localhost:1234"}, err: nil},
		},
		"query arguments are recognised": {
			args: args{dbURL: url.URL{Scheme: "postgres", Host: "localhost:1234", RawQuery: "sslmode=disable&extra=true"}},
			want: want{got: Config{driver: DriverPgx, dsn: "postgres://localhost:1234?sslmode=disable&extra=true"}, err: nil},
		},
		"database path is recognised": {
			args: args{dbURL: url.URL{Scheme: "postgres", Host: "localhost:1234", Path: "/db1", RawQuery: "sslmode=disable"}},
			want: want{got: Config{driver: DriverPgx, dsn: "postgres://localhost:1234/db1?sslmode=disable"}, err: nil},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.dbURL)
			require.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
