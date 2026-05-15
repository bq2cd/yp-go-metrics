package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/cli"
	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
)

func Test_parseArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name      string
		args      args
		want      config.Config
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "no args",
			args: args{args: []string{}},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "extra positional args",
			args: args{args: []string{"-r", "123", "456"}},
			want: config.Config{
				ListenAddress:            defaultAddress,
				ShutdownTimeout:          defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval:      defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath:      filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				MetricStoreLoadOnStartup: true,
			},
			assertion: assert.NoError,
		},
		{
			name:      "unknown args",
			args:      args{args: []string{"-x", "--test"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, address",
			args:      args{args: []string{"-a"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, address 2",
			args:      args{args: []string{"-a", "host1:host2:host3"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, shutdown timeout",
			args:      args{args: []string{"-t=0"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, metric store interval",
			args:      args{args: []string{"-i=-1"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, metric store path",
			args:      args{args: []string{"-f=" + servertest.GetCwd(t)}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, metric store load at startup",
			args:      args{args: []string{"-r=123"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, invalid database url",
			args:      args{args: []string{"-d=postgres://"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, missing secret key value",
			args:      args{args: []string{"-k"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name: "good args, address",
			args: args{args: []string{"-a=host1"}},
			want: config.Config{
				ListenAddress:       "host1",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, address 1",
			args: args{args: []string{"-a=host1:81"}},
			want: config.Config{
				ListenAddress:       "host1:81",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, address 2",
			args: args{args: []string{"-a", "host2:82"}},
			want: config.Config{
				ListenAddress:       "host2:82",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, address 3",
			args: args{args: []string{"-a", ":83"}},
			want: config.Config{
				ListenAddress:       ":83",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, address 4",
			args: args{args: []string{"-a", ":0"}},
			want: config.Config{
				ListenAddress:       ":0",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, shutdown timeout",
			args: args{args: []string{"-a", ":0", "-t=5"}},
			want: config.Config{
				ListenAddress:       ":0",
				ShutdownTimeout:     5 * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, metric store interval",
			args: args{args: []string{"-a", ":0", "-t=5", "-i=12"}},
			want: config.Config{
				ListenAddress:       ":0",
				ShutdownTimeout:     5 * time.Second,
				MetricStoreInterval: 12 * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, metric store interval 2",
			args: args{args: []string{"-a", ":0", "-t=5", "-i=0"}},
			want: config.Config{
				ListenAddress:       ":0",
				ShutdownTimeout:     5 * time.Second,
				MetricStoreInterval: 0,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, metric store path",
			args: args{args: []string{"-a", ":0", "-t=5", "-i=12", "-f=/some/path/here.txt"}},
			want: config.Config{
				ListenAddress:       ":0",
				ShutdownTimeout:     5 * time.Second,
				MetricStoreInterval: 12 * time.Second,
				MetricStoreFilePath: "/some/path/here.txt",
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, metric store load at startup",
			args: args{args: []string{"-a", ":0", "-t=5", "-i=12", "-f=/some/path/here.txt", "-r"}},
			want: config.Config{
				ListenAddress:            ":0",
				ShutdownTimeout:          5 * time.Second,
				MetricStoreInterval:      12 * time.Second,
				MetricStoreFilePath:      "/some/path/here.txt",
				MetricStoreLoadOnStartup: true,
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, database url",
			args: args{args: []string{"-d=postgresql://localhost:1234"}},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				DatabaseURL:         url.URL{Scheme: "postgresql", Host: "localhost:1234"},
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, empty secret key is valid",
			args: args{args: []string{"-k="}},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				HMACSecretKey:       nil,
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, secret key",
			args: args{args: []string{"-k=MTIz"}},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				HMACSecretKey:       []byte(`123`),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, audit file path",
			args: args{args: []string{"--audit-file=audit.txt"}},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				AuditFilePath:       filepath.Join(servertest.GetCwd(t), "audit.txt"),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, audit url",
			args: args{args: []string{"--audit-url=http://localhost:8888/audit"}},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				AuditURL:            url.URL{Scheme: "http", Host: "localhost:8888", Path: "/audit"},
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet(tt.name, flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			got, err := parseArgs(fs, tt.args.args, envparser.NewParserWithOptions(env.Options{Environment: map[string]string{}}))
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseArgs_withEnv(t *testing.T) {
	type args struct {
		args []string
		env  map[string]string
	}
	tests := []struct {
		name      string
		args      args
		want      config.Config
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "env overrides address",
			args: args{
				args: []string{"-a=localhost:9090"},
				env:  map[string]string{"ADDRESS": "localhost:3333"},
			},
			want: config.Config{
				ListenAddress:       "localhost:3333",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides shutdown timeout",
			args: args{
				args: []string{"-a=localhost:9090"},
				env:  map[string]string{"SHUTDOWN_TIMEOUT": "21"},
			},
			want: config.Config{
				ListenAddress:       "localhost:9090",
				ShutdownTimeout:     21 * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides shutdown timeout 2",
			args: args{
				args: []string{"-a=localhost:9090", "-t=13"},
				env:  map[string]string{"SHUTDOWN_TIMEOUT": "21"},
			},
			want: config.Config{
				ListenAddress:       "localhost:9090",
				ShutdownTimeout:     21 * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "env invalid shutdown timeout",
			args: args{
				args: []string{"-a=localhost:9090"},
				env:  map[string]string{"SHUTDOWN_TIMEOUT": "-5"},
			},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name: "env invalid address",
			args: args{
				args: []string{"-a=localhost:9090"},
				env:  map[string]string{"ADDRESS": "123:456:789"},
			},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name: "env overrides metric store interval",
			args: args{
				args: []string{"-a=localhost:9090", "-t=13", "-i=8"},
				env:  map[string]string{"STORE_INTERVAL": "18"},
			},
			want: config.Config{
				ListenAddress:       "localhost:9090",
				ShutdownTimeout:     13 * time.Second,
				MetricStoreInterval: 18 * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides metric store path",
			args: args{
				args: []string{"-a=localhost:9090", "-t=13", "-i=8", "-f=/a/path/to/some/file.json"},
				env:  map[string]string{"FILE_STORAGE_PATH": "/an/override/to/a/different/file.json"},
			},
			want: config.Config{
				ListenAddress:       "localhost:9090",
				ShutdownTimeout:     13 * time.Second,
				MetricStoreInterval: 8 * time.Second,
				MetricStoreFilePath: "/an/override/to/a/different/file.json",
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides metric store load on startup",
			args: args{
				args: []string{"-a=localhost:9090", "-t=13", "-i=8", "-f=/a/path/to/some/file.json", "-r"},
				env:  map[string]string{"RESTORE": "false"},
			},
			want: config.Config{
				ListenAddress:            "localhost:9090",
				ShutdownTimeout:          13 * time.Second,
				MetricStoreInterval:      8 * time.Second,
				MetricStoreFilePath:      "/a/path/to/some/file.json",
				MetricStoreLoadOnStartup: false,
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides database url",
			args: args{
				args: []string{"-d=postgres://localhost:1234"},
				env:  map[string]string{"DATABASE_DSN": "postgres://user:password@localhost:4567/db1"},
			},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				DatabaseURL:         url.URL{Scheme: "postgres", User: url.UserPassword("user", "password"), Host: "localhost:4567", Path: "/db1"},
			},
			assertion: assert.NoError,
		},
		{
			name: "env invalid database url",
			args: args{
				args: []string{"-d=postgres://localhost:1234"},
				env:  map[string]string{"DATABASE_DSN": "mysql://user:password@localhost:4567/db1"},
			},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name: "env sets secret key",
			args: args{
				args: []string{},
				env:  map[string]string{"KEY": "NDU2"},
			},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				HMACSecretKey:       []byte(`456`),
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides secret key",
			args: args{
				args: []string{"-k=MTIz"},
				env:  map[string]string{"KEY": "NDU2"},
			},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				HMACSecretKey:       []byte(`456`),
			},
			assertion: assert.NoError,
		},
		{
			name: "env does not override secret key with empty value",
			args: args{
				args: []string{"-k=MTIz"},
				env:  map[string]string{"KEY": ""},
			},
			want: config.Config{
				ListenAddress:       defaultAddress,
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				HMACSecretKey:       []byte(`123`),
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides audit file path",
			args: args{
				args: []string{"-a=localhost:9090", "--audit-file=/a/path/to/some/file.json"},
				env:  map[string]string{"AUDIT_FILE": "/an/override/to/a/different/file.json"},
			},
			want: config.Config{
				ListenAddress:       "localhost:9090",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				AuditFilePath:       "/an/override/to/a/different/file.json",
			},
			assertion: assert.NoError,
		},
		{
			name: "env overrides audit url",
			args: args{
				args: []string{"-a=localhost:9090", "--audit-url=http://localhost:8888/audit-1"},
				env:  map[string]string{"AUDIT_URL": "http://localhost:9999/audit-2"},
			},
			want: config.Config{
				ListenAddress:       "localhost:9090",
				ShutdownTimeout:     defaultShutdownTimeoutSec * time.Second,
				MetricStoreInterval: defaultMetricStoreIntervalSec * time.Second,
				MetricStoreFilePath: filepath.Join(servertest.GetCwd(t), defaultMetricStoreFilePath),
				AuditURL:            url.URL{Scheme: "http", Host: "localhost:9999", Path: "/audit-2"},
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet(tt.name, flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			got, err := parseArgs(fs, tt.args.args, envparser.NewParserWithOptions(env.Options{Environment: tt.args.env}))
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_run(t *testing.T) {
	type args struct {
		stderr io.Writer
	}
	type setup struct {
		contextFunc func() (context.Context, context.CancelFunc)
		argsFunc    func() ([]string, *http.Request)
	}
	tests := []struct {
		name          string
		skip          func() (bool, string)
		args          args
		setup         setup
		wantStderr    string
		assertionResp func(*testing.T, *http.Request, context.CancelFunc)
		assertionErr  assert.ErrorAssertionFunc
	}{
		{
			name: "normal flow",
			skip: func() (bool, string) {
				v := (os.Getenv("GITHUB_ACTIONS") != "")
				return v, "takes too long inside Github actions"
			},
			args: args{stderr: os.Stderr},
			setup: setup{
				contextFunc: func() (context.Context, context.CancelFunc) {
					return context.WithCancel(t.Context())
				},
				argsFunc: func() ([]string, *http.Request) {
					addr := servertest.GetRandomListenAddress(t)
					t.Logf("got random listen address: %v", addr)

					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", addr), http.NoBody)
					require.NoError(t, err)

					return []string{"-a", addr}, req
				},
			},
			assertionResp: func(t *testing.T, req *http.Request, cancel context.CancelFunc) {
				defer cancel()
				err := servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
				require.NoError(t, err)
			},
			assertionErr: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "signal imitation",
			skip: func() (bool, string) {
				v := (os.Getenv("GITHUB_ACTIONS") != "")
				return v, "takes too long inside Github actions"
			},
			args: args{stderr: os.Stderr},
			setup: setup{
				contextFunc: func() (context.Context, context.CancelFunc) {
					return context.WithTimeout(t.Context(), 500*time.Millisecond)
				},
				argsFunc: func() ([]string, *http.Request) {
					addr := servertest.GetRandomListenAddress(t)
					t.Logf("got random listen address: %v", addr)

					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", addr), http.NoBody)
					require.NoError(t, err)

					return []string{"-a", addr}, req
				},
			},
			assertionResp: func(t *testing.T, req *http.Request, cancel context.CancelFunc) {
				err := servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
				assert.NoError(t, err)
			},
			assertionErr: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "bad args",
			skip: func() (bool, string) {
				return false, ""
			},
			args: args{stderr: io.Discard},
			setup: setup{
				contextFunc: func() (context.Context, context.CancelFunc) {
					return context.WithCancel(t.Context())
				},
				argsFunc: func() ([]string, *http.Request) {
					return []string{"-zzz", "gibberish"}, nil
				},
			},
			assertionResp: func(t *testing.T, req *http.Request, cancel context.CancelFunc) {
				cancel()
			},
			assertionErr: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, msg := tt.skip(); ok {
				t.Skipf("%s", msg)
			}

			ctx, cancel := tt.setup.contextFunc()
			args, req := tt.setup.argsFunc()

			errCh := make(chan error, 1)

			go func() {
				errCh <- run(ctx, args, cli.TerminalConfig{
					Stdout: io.Discard,
					Stderr: tt.args.stderr,
				})
			}()

			time.Sleep(100 * time.Millisecond)

			tt.assertionResp(t, req, cancel)

			err := <-errCh
			t.Logf("run finished with %v", err)

			tt.assertionErr(t, err)
		})
	}
}
