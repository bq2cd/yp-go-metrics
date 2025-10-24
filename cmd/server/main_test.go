package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/server/servertest"
	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	flagMainArgsJSON = flag.String("mainArgsJSON", "", "flags to pass to main() function during tests, encoded as JSON")
)

func encodeArgsToJSON(t *testing.T, args []string) string {
	out, err := json.Marshal(args)
	require.NoError(t, err)
	return string(out)
}

func decodeArgsFromJSON(t *testing.T, input string) []string {
	out := make([]string, 0)
	err := json.Unmarshal([]byte(input), &out)
	require.NoError(t, err)
	return out
}

func skipIfGithubActions(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("not supported in Github Actions")
	}
}

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
				errCh <- run(ctx, args, tt.args.stderr)
			}()

			time.Sleep(100 * time.Millisecond)

			tt.assertionResp(t, req, cancel)

			err := <-errCh
			t.Logf("run finished with %v", err)

			tt.assertionErr(t, err)
		})
	}
}

func Test_main_subprocess(t *testing.T) {
	if os.Getenv("TEST_MAIN_SUBPROCESS") != "1" {
		return
	}
	t.Logf("orig os args: %v", os.Args)
	t.Logf("main args: %v", *flagMainArgsJSON)
	mainArgs := decodeArgsFromJSON(t, *flagMainArgsJSON)
	os.Args = append([]string{os.Args[0]}, mainArgs...)
	t.Logf("modified os args: %v", os.Args)
	main()
}

func Test_main(t *testing.T) {
	skipIfGithubActions(t)
	addrFactory := servertest.NewListenAddressFactory(t)
	tempFactory := servertest.NewTempFileFactory(t)
	defer tempFactory.RemoveAll()
	type want struct {
		exitCode int
	}
	tests := []struct {
		name          string
		args          []string
		env           map[string]string
		want          want
		assertRunning func(*testing.T, string)
		assertStopped func(*testing.T, map[string]string)
	}{
		{
			name: "address via args",
			args: []string{"-a", addrFactory.New()},
			env:  map[string]string{},
			want: want{
				exitCode: 0,
			},
			assertRunning: func(t *testing.T, addr string) {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", addr), http.NoBody)
				require.NoError(t, err)
				err = servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
				assert.NoError(t, err)
			},
			assertStopped: func(t *testing.T, env map[string]string) {
			},
		},
		{
			name: "address via env",
			args: []string{},
			env: map[string]string{
				"ADDRESS": addrFactory.New(),
			},
			want: want{
				exitCode: 0,
			},
			assertRunning: func(t *testing.T, addr string) {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", addr), http.NoBody)
				require.NoError(t, err)
				err = servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
				assert.NoError(t, err)
			},
			assertStopped: func(t *testing.T, env map[string]string) {
			},
		},
		{
			name: "dump metrics on every write",
			args: []string{"-i=0"},
			env: map[string]string{
				"ADDRESS":           addrFactory.New(),
				"FILE_STORAGE_PATH": tempFactory.Create("test-metrics-dump-*"),
			},
			want: want{
				exitCode: 0,
			},
			assertRunning: func(t *testing.T, addr string) {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/", addr), bytes.NewReader([]byte(`{"id": "id1", "type": "counter", "delta": 78}`)))
				require.NoError(t, err)
				httpheaders.ContentTypeApplicationJSON.Apply(req.Header)
				err = servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
				assert.NoError(t, err)
			},
			assertStopped: func(t *testing.T, env map[string]string) {
				dump, err := os.ReadFile(env["FILE_STORAGE_PATH"])
				require.NoError(t, err)
				t.Logf("dumped metrics (%s): %s", env["FILE_STORAGE_PATH"], string(dump))
				assert.JSONEq(t, `[{"id": "id1", "type": "counter", "delta": 78}]`, string(dump))
			},
		},
		{
			name: "load metrics on startup",
			args: []string{"-i=999", "-r"},
			env: map[string]string{
				"ADDRESS": addrFactory.New(),
				"FILE_STORAGE_PATH": func() string {
					f := tempFactory.Create("test-metrics-dump-*")
					err := os.WriteFile(f, []byte(`[{"id": "id1", "type": "counter", "delta": 78}]`), 0600)
					require.NoError(t, err)
					return f
				}(),
			},
			want: want{
				exitCode: 0,
			},
			assertRunning: func(t *testing.T, addr string) {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/value/", addr), bytes.NewBufferString(`{"id": "id1", "type": "counter"}`))
				require.NoError(t, err)
				httpheaders.ContentTypeApplicationJSON.Apply(req.Header)
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.JSONEq(t, `{"id": "id1", "type": "counter", "delta": 78}`, string(body))
			},
			assertStopped: func(t *testing.T, env map[string]string) {
			},
		},
		{
			name: "server applies migrations on startup",
			args: []string{"-i=0"},
			env: map[string]string{
				"ADDRESS": addrFactory.New(),
				"DATABASE_DSN": func() string {
					dbCfg := servertest.LaunchEmbeddedPostgres(t, "server-test-user", "server-test-password", "server-test-db")
					return dbCfg.DSN()
				}(),
			},
			want: want{
				exitCode: 0,
			},
			assertRunning: func(t *testing.T, addr string) {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/ping", addr), http.NoBody)
				require.NoError(t, err)
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, `OK`, string(body))
			},
			assertStopped: func(t *testing.T, env map[string]string) {
				dbURL, err := url.Parse(env["DATABASE_DSN"])
				require.NoError(t, err)
				cfg, err := dbconfig.New(*dbURL)
				require.NoError(t, err)
				db, err := sql.Open("pgx", cfg.DSN())
				require.NoError(t, err)
				rows, err := db.Query(`
						SELECT schemaname, tablename
						FROM pg_catalog.pg_tables
						WHERE schemaname NOT IN ('information_schema', 'pg_catalog');
					`)
				require.NoError(t, err)
				defer rows.Close()
				assert.Truef(t, rows.Next(), "expected at least a single row")
				assert.NoError(t, rows.Err())
			},
		},
	}
	for tid, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(os.Args[0], "-test.run=Test_main_subprocess", "-test.v", "-mainArgsJSON", encodeArgsToJSON(t, tt.args))
			cmd.Env = append(cmd.Env, "TEST_MAIN_SUBPROCESS=1")
			for k, v := range tt.env {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
			}
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			t.Logf("starting subprocess with args: %v (env: %v)", cmd.Args, cmd.Env)

			err := cmd.Start()
			require.NoError(t, err)

			go func() {
				time.Sleep(500 * time.Millisecond)
				_ = syscall.Kill(cmd.Process.Pid, syscall.SIGINT)
			}()

			time.Sleep(100 * time.Millisecond)
			tt.assertRunning(t, addrFactory.Get(tid))

			_ = cmd.Wait()

			tt.assertStopped(t, tt.env)

			t.Logf("subprocess stdout:\n%s\n", stdout.String())
			t.Logf("subprocess stderr:\n%s\n", stderr.String())

			if tt.want.exitCode == 0 {
				require.NoError(t, err)
			} else {
				status, ok := err.(*exec.ExitError)
				require.True(t, ok)
				assert.Equal(t, tt.want.exitCode, status.ExitCode())
			}
		})
	}
}
