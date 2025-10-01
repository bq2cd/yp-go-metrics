package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/server/servertest"
	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			want: config.Config{ListenAddress: defaultAddress, ShutdownTimeout: defaultShutdownTimeoutSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "unknown args",
			args: args{args: []string{"-x", "--test"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args",
			args: args{args: []string{"-a"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 2",
			args: args{args: []string{"-a", "host1:host2:host3"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 3",
			args: args{args: []string{"-t=0"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "good args",
			args: args{args: []string{"-a=host1"}},
			want: config.Config{ListenAddress: "host1", ShutdownTimeout: defaultShutdownTimeoutSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 1",
			args: args{args: []string{"-a=host1:81"}},
			want: config.Config{ListenAddress: "host1:81", ShutdownTimeout: defaultShutdownTimeoutSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 2",
			args: args{args: []string{"-a", "host2:82"}},
			want: config.Config{ListenAddress: "host2:82", ShutdownTimeout: defaultShutdownTimeoutSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 3",
			args: args{args: []string{"-a", ":83"}},
			want: config.Config{ListenAddress: ":83", ShutdownTimeout: defaultShutdownTimeoutSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 4",
			args: args{args: []string{"-a", ":0"}},
			want: config.Config{ListenAddress: ":0", ShutdownTimeout: defaultShutdownTimeoutSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 5",
			args: args{args: []string{"-a", ":0", "-t=1"}},
			want: config.Config{ListenAddress: ":0", ShutdownTimeout: 1 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
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
			want: config.Config{ListenAddress: "localhost:3333", ShutdownTimeout: defaultShutdownTimeoutSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "env overrides shutdown timeout",
			args: args{
				args: []string{"-a=localhost:9090"},
				env:  map[string]string{"SHUTDOWN_TIMEOUT": "21"},
			},
			want: config.Config{ListenAddress: "localhost:9090", ShutdownTimeout: 21 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "env overrides shutdown timeout 2",
			args: args{
				args: []string{"-a=localhost:9090", "-t=13"},
				env:  map[string]string{"SHUTDOWN_TIMEOUT": "21"},
			},
			want: config.Config{ListenAddress: "localhost:9090", ShutdownTimeout: 21 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "env invalid shutdown timeout",
			args: args{
				args: []string{"-a=localhost:9090"},
				env:  map[string]string{"SHUTDOWN_TIMEOUT": "-5"},
			},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "env invalid address",
			args: args{
				args: []string{"-a=localhost:9090"},
				env:  map[string]string{"ADDRESS": "123:456:789"},
			},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
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

func Test_runServer(t *testing.T) {
	type args struct {
		ctx context.Context
		cfg config.Config
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// `runServer` is always called by `run`, so there is no
		// need for a separate test here.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, runServer(tt.args.ctx, tt.args.cfg))
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
		// TODO: Add test cases.
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
				err := servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
				assert.NoError(t, err)
				cancel()
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

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// `main` only calls `run` under the hood,
		// so there's not much to test here
		// unless we could mock `run` function;
		// and the only way to mock it would be
		// to assign it to a global variable.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
