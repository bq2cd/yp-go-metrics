package server

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name      string
		args      args
		want      *Config
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "empty",
			args:      args{},
			want:      &Config{},
			assertion: assert.NoError,
		},
		{
			name:      "listen address",
			args:      args{opts: []Option{ListenAddress("localhost:91")}},
			want:      &Config{ListenAddress: "localhost:91"},
			assertion: assert.NoError,
		},
		{
			name:      "shutdown timeout",
			args:      args{opts: []Option{ShutdownTimeout(5)}},
			want:      &Config{ShutdownTimeout: 5 * time.Second},
			assertion: assert.NoError,
		},
		{
			name: "multiple options",
			args: args{opts: []Option{
				ListenAddress("localhost:83"),
				ShutdownTimeout(3),
			}},
			want: &Config{
				ListenAddress:   "localhost:83",
				ShutdownTimeout: 3 * time.Second,
			},
			assertion: assert.NoError,
		},
		{
			name: "multiple options but some invalid",
			args: args{opts: []Option{
				ListenAddress("localhost:83"),
				ShutdownTimeout(0),
			}},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.opts...)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListenAddress(t *testing.T) {
	type args struct {
		addr string
	}
	type want struct {
		addr string
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "empty",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "invalid",
			args:   args{addr: "localhost:91:invalid"},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "valid with empty config",
			args:   args{addr: "localhost:91"},
			want:   want{addr: "localhost:91"},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.addr, c.ListenAddress)
			},
		},
		{
			name:   "valid with empty config 2",
			args:   args{addr: ":91"},
			want:   want{addr: ":91"},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.addr, c.ListenAddress)
			},
		},
		{
			name:   "valid with empty config 3",
			args:   args{addr: ":0"},
			want:   want{addr: ":0"},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.addr, c.ListenAddress)
			},
		},
		{
			name:   "valid with preexisting config",
			args:   args{addr: "127.0.0.1:39"},
			want:   want{addr: "127.0.0.1:39"},
			config: Config{ListenAddress: ":0"},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.addr, c.ListenAddress)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ListenAddress(tt.args.addr)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func TestShutdownTimeout(t *testing.T) {
	type args struct {
		timeoutSec uint
	}
	type want struct {
		timeout time.Duration
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{timeoutSec: 35},
			want:   want{timeout: 35 * time.Second},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.timeout, c.ShutdownTimeout)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{timeoutSec: 35},
			want:   want{timeout: 35 * time.Second},
			config: Config{ShutdownTimeout: 10 * time.Second},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.timeout, c.ShutdownTimeout)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ShutdownTimeout(tt.args.timeoutSec)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		ListenAddress            string
		ShutdownTimeout          time.Duration
		MetricStoreInterval      time.Duration
		MetricStoreFilePath      string
		MetricStoreLoadOnStartup bool
		DatabaseURL              url.URL
		HMACSecretKey            []byte
		AuditFilePath            string
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "empty",
			fields:    fields{},
			assertion: assert.Error,
		},
		{
			name: "zero shutdown timeout",
			fields: fields{
				ListenAddress: ":0",
			},
			assertion: assert.Error,
		},
		{
			name: "empty store path when loading at startup",
			fields: fields{
				ListenAddress:            ":0",
				ShutdownTimeout:          1 * time.Second,
				MetricStoreLoadOnStartup: true,
			},
			assertion: assert.Error,
		},
		{
			name: "invalid database scheme",
			fields: fields{
				ListenAddress:   ":0",
				ShutdownTimeout: 1 * time.Second,
				DatabaseURL:     url.URL{Scheme: "mysql", Host: "localhost:5432"},
			},
			assertion: assert.Error,
		},
		{
			name: "invalid audit file path",
			fields: fields{
				ListenAddress:       ":0",
				ShutdownTimeout:     1 * time.Second,
				MetricStoreFilePath: "test.json",
				AuditFilePath:       "/tmp", // cannot use directory
			},
			assertion: assert.Error,
		},
		{
			name: "all good",
			fields: fields{
				ListenAddress:       ":0",
				ShutdownTimeout:     1 * time.Second,
				MetricStoreFilePath: "test.json",
			},
			assertion: assert.NoError,
		},
		{
			name: "all good with database",
			fields: fields{
				ListenAddress:       ":0",
				ShutdownTimeout:     1 * time.Second,
				MetricStoreFilePath: "test.json",
				DatabaseURL:         url.URL{Scheme: "postgres", Host: "localhost:5432", Path: "/db1", RawQuery: "sslmode=verify-full"},
			},
			assertion: assert.NoError,
		},
		{
			name: "non-empty secret key",
			fields: fields{
				ListenAddress:   ":0",
				ShutdownTimeout: 1 * time.Second,
				HMACSecretKey:   []byte(`123`),
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ListenAddress:            tt.fields.ListenAddress,
				ShutdownTimeout:          tt.fields.ShutdownTimeout,
				MetricStoreInterval:      tt.fields.MetricStoreInterval,
				MetricStoreFilePath:      tt.fields.MetricStoreFilePath,
				MetricStoreLoadOnStartup: tt.fields.MetricStoreLoadOnStartup,
				DatabaseURL:              tt.fields.DatabaseURL,
				HMACSecretKey:            tt.fields.HMACSecretKey,
				AuditFilePath:            tt.fields.AuditFilePath,
			}
			tt.assertion(t, c.Validate())
		})
	}
}

func TestMetricStoreInterval(t *testing.T) {
	type args struct {
		intervalSec uint
	}
	type want struct {
		interval time.Duration
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.interval, c.MetricStoreInterval)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{MetricStoreInterval: 10 * time.Second},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
				assert.Equal(t, want.interval, c.MetricStoreInterval)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MetricStoreInterval(tt.args.intervalSec)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func TestMetricStoreFilePath(t *testing.T) {
	type args struct {
		path string
	}
	type want struct {
		path string
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "empty is allowed",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "dir not allowed",
			args:   args{path: "."},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name: "relative",
			args: args{path: "test.txt"},
			want: want{path: func() string {
				p, err := os.Getwd()
				require.NoError(t, err)
				return filepath.Join(p, "test.txt")
			}()},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "absolute",
			args:   args{path: "/test/me/here/please.txt"},
			want:   want{path: "/test/me/here/please.txt"},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "overrides previous value",
			args:   args{path: "/test/me/here/please.txt"},
			want:   want{path: "/test/me/here/please.txt"},
			config: Config{MetricStoreFilePath: "/a/default/path"},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MetricStoreFilePath(tt.args.path)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func TestMetricStoreLoadOnStartup(t *testing.T) {
	type args struct {
		action bool
	}
	type want struct {
		action bool
	}
	type testcase struct {
		args   args
		want   want
		config Config
	}
	tests := map[string]testcase{
		"false -> true": {
			args:   args{action: true},
			want:   want{action: true},
			config: Config{},
		},
		"true -> false": {
			args:   args{action: false},
			want:   want{action: false},
			config: Config{MetricStoreLoadOnStartup: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &tt.config
			err := MetricStoreLoadOnStartup(tt.args.action)(c)
			require.NoError(t, err)
			assert.Equal(t, tt.want.action, c.MetricStoreLoadOnStartup)
		})
	}
}

func TestDatabaseURL(t *testing.T) {
	type args struct {
		dsn string
	}
	type want struct {
		url url.URL
	}
	type testcase struct {
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}
	tests := map[string]testcase{
		"empty dsn is allowed": {
			args:   args{dsn: ""},
			want:   want{url: url.URL{}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"invalid url fails": {
			args:   args{dsn: "localhost:" + string([]rune{0x7f}) + "99"},
			want:   want{url: url.URL{}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		"only postgres/postgresql scheme is allowed": {
			args:   args{dsn: "mysql://"},
			want:   want{url: url.URL{}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		"missing host": {
			args:   args{dsn: "postgres://"},
			want:   want{url: url.URL{}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		"missing port": {
			args:   args{dsn: "postgresql://localhost"},
			want:   want{url: url.URL{}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		"valid url with postgres": {
			args:   args{dsn: "postgres://localhost:5432"},
			want:   want{url: url.URL{Scheme: "postgres", Host: "localhost:5432"}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"valid url with postgresql": {
			args:   args{dsn: "postgresql://localhost:5432"},
			want:   want{url: url.URL{Scheme: "postgresql", Host: "localhost:5432"}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"valid url with database": {
			args:   args{dsn: "postgres://localhost:5432/test-db"},
			want:   want{url: url.URL{Scheme: "postgres", Host: "localhost:5432", Path: "/test-db"}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"overrides previous url": {
			args:   args{dsn: "postgresql://localhost:1234/test-db-1"},
			want:   want{url: url.URL{Scheme: "postgresql", Host: "localhost:1234", Path: "/test-db-1"}},
			config: Config{DatabaseURL: url.URL{Scheme: "postgres", Host: "localhost:5432", Path: "/test-db"}},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"valid url with username and password": {
			args:   args{dsn: "postgres://user:password@127.0.0.1:9876/db-11_dev?sslmode=verify-full"},
			want:   want{url: url.URL{Scheme: "postgres", User: url.UserPassword("user", "password"), Host: "127.0.0.1:9876", Path: "/db-11_dev", RawQuery: "sslmode=verify-full"}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &tt.config
			err := DatabaseURL(tt.args.dsn)(c)
			tt.assertion(t, &tt.config, err, tt.want)
			assert.Equal(t, tt.want.url, c.DatabaseURL)
		})
	}
}

func TestHMACSecretKey(t *testing.T) {
	type args struct {
		key string
	}
	type want struct {
		got []byte
	}
	type testcase struct {
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}
	tests := map[string]testcase{
		"empty key is okay": {
			args:   args{key: ""},
			want:   want{got: nil},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"non-empty plain-text key is accepted": {
			args:   args{key: `123`},
			want:   want{got: []byte(`123`)},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"non-empty base64-encoded key is accepted": {
			args:   args{key: "MTIz"},
			want:   want{got: []byte(`123`)},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"existing key can be overwritten": {
			args:   args{key: "MTIz"},
			want:   want{got: []byte(`123`)},
			config: Config{HMACSecretKey: []byte(`something`)},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"existing key can be overwritten with empty value": {
			args:   args{key: ""},
			want:   want{got: nil},
			config: Config{HMACSecretKey: []byte(`something`)},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &tt.config
			err := HMACSecretKey(tt.args.key)(c)
			tt.assertion(t, c, err, tt.want)
			assert.Equal(t, tt.want.got, c.HMACSecretKey)
		})
	}
}

func TestAuditFilePath(t *testing.T) {
	type args struct {
		path string
	}
	type want struct {
		path string
	}
	tests := []struct {
		name      string
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}{
		{
			name:   "empty is allowed",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "dir not allowed",
			args:   args{path: "."},
			want:   want{},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name: "relative path is allowed",
			args: args{path: "test.txt"},
			want: want{path: func() string {
				p, err := os.Getwd()
				require.NoError(t, err)
				return filepath.Join(p, "test.txt")
			}()},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "absolute path is allowed",
			args:   args{path: "/test/me/here/please.txt"},
			want:   want{path: "/test/me/here/please.txt"},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "overrides previous value",
			args:   args{path: "/test/me/here/please.txt"},
			want:   want{path: "/test/me/here/please.txt"},
			config: Config{MetricStoreFilePath: "/a/default/path"},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AuditFilePath(tt.args.path)(&tt.config)
			tt.assertion(t, &tt.config, err, tt.want)
		})
	}
}

func TestAuditURL(t *testing.T) {
	type args struct {
		input string
	}
	type want struct {
		url url.URL
	}
	type testcase struct {
		args      args
		want      want
		config    Config
		assertion func(*testing.T, *Config, error, want)
	}
	tests := map[string]testcase{
		"empty url is allowed": {
			args:   args{input: ""},
			want:   want{url: url.URL{}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
		"invalid url is not allowed": {
			args:   args{input: "::localhost:99"},
			want:   want{url: url.URL{}},
			config: Config{},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.Error(t, err)
			},
		},
		"overrides previous url": {
			args:   args{input: "http://localhost:1234/audit"},
			want:   want{url: url.URL{Scheme: "http", Host: "localhost:1234", Path: "/audit"}},
			config: Config{AuditURL: url.URL{Scheme: "", Host: "localhost:5432", Path: "/audit-old"}},
			assertion: func(t *testing.T, c *Config, err error, want want) {
				require.NoError(t, err)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &tt.config
			err := AuditURL(tt.args.input)(c)
			tt.assertion(t, &tt.config, err, tt.want)
			assert.Equal(t, tt.want.url, c.AuditURL)
		})
	}
}
