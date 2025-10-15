package server

import (
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
			name: "empty",
			args: args{},
			want: &Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "listen address",
			args: args{opts: []Option{ListenAddress("localhost:91")}},
			want: &Config{ListenAddress: "localhost:91"},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "shutdown timeout",
			args: args{opts: []Option{ShutdownTimeout(5)}},
			want: &Config{ShutdownTimeout: 5 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
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
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple options but some invalid",
			args: args{opts: []Option{
				ListenAddress("localhost:83"),
				ShutdownTimeout(0),
			}},
			want: nil,
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
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
		assertion func(assert.TestingT, *Config, error, want)
	}{
		{
			name:   "empty",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "invalid",
			args:   args{addr: "localhost:91:invalid"},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "valid with empty config",
			args:   args{addr: "localhost:91"},
			want:   want{addr: "localhost:91"},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.addr, c.ListenAddress)
			},
		},
		{
			name:   "valid with empty config 2",
			args:   args{addr: ":91"},
			want:   want{addr: ":91"},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.addr, c.ListenAddress)
			},
		},
		{
			name:   "valid with empty config 3",
			args:   args{addr: ":0"},
			want:   want{addr: ":0"},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.addr, c.ListenAddress)
			},
		},
		{
			name:   "valid with preexisting config",
			args:   args{addr: "127.0.0.1:39"},
			want:   want{addr: "127.0.0.1:39"},
			config: Config{ListenAddress: ":0"},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
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
		assertion func(assert.TestingT, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.Error(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{timeoutSec: 35},
			want:   want{timeout: 35 * time.Second},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.timeout, c.ShutdownTimeout)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{timeoutSec: 35},
			want:   want{timeout: 35 * time.Second},
			config: Config{ShutdownTimeout: 10 * time.Second},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
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
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty",
			fields: fields{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "zero shutdown timeout",
			fields: fields{
				ListenAddress: ":0",
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "empty store path when loading at startup",
			fields: fields{
				ListenAddress:            ":0",
				ShutdownTimeout:          1 * time.Second,
				MetricStoreLoadOnStartup: true,
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "all good",
			fields: fields{
				ListenAddress:       ":0",
				ShutdownTimeout:     1 * time.Second,
				MetricStoreFilePath: "test.json",
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
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
		assertion func(assert.TestingT, *Config, error, want)
	}{
		{
			name:   "zero",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "positive with empty config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
				assert.Equal(t, want.interval, c.MetricStoreInterval)
			},
		},
		{
			name:   "positive with existing config",
			args:   args{intervalSec: 35},
			want:   want{interval: 35 * time.Second},
			config: Config{MetricStoreInterval: 10 * time.Second},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
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
		assertion func(assert.TestingT, *Config, error, want)
	}{
		{
			name:   "empty is allowed",
			args:   args{},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "dir not allowed",
			args:   args{path: "."},
			want:   want{},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
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
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "absolute",
			args:   args{path: "/test/me/here/please.txt"},
			want:   want{path: "/test/me/here/please.txt"},
			config: Config{},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "overrides previous value",
			args:   args{path: "/test/me/here/please.txt"},
			want:   want{path: "/test/me/here/please.txt"},
			config: Config{MetricStoreFilePath: "/a/default/path"},
			assertion: func(t assert.TestingT, c *Config, err error, want want) {
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
	tests := []struct {
		name string
		args args
		want Option
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MetricStoreLoadOnStartup(tt.args.action))
		})
	}
}
