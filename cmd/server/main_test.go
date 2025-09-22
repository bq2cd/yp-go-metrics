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

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/server/servertest"
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
			want: config.Config{ListenAddress: defaultAddress, ShutdownTimeout: defaultShutdownTimeout},
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
			name: "good args",
			args: args{args: []string{"-a=host1"}},
			want: config.Config{ListenAddress: "host1", ShutdownTimeout: defaultShutdownTimeout},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 1",
			args: args{args: []string{"-a=host1:81"}},
			want: config.Config{ListenAddress: "host1:81", ShutdownTimeout: defaultShutdownTimeout},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 2",
			args: args{args: []string{"-a", "host2:82"}},
			want: config.Config{ListenAddress: "host2:82", ShutdownTimeout: defaultShutdownTimeout},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 3",
			args: args{args: []string{"-a", ":83"}},
			want: config.Config{ListenAddress: ":83", ShutdownTimeout: defaultShutdownTimeout},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 4",
			args: args{args: []string{"-a", ":0"}},
			want: config.Config{ListenAddress: ":0", ShutdownTimeout: defaultShutdownTimeout},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet(tt.name, flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			got, err := parseArgs(fs, tt.args.args)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRun(t *testing.T) {
	if v := os.Getenv("GITHUB_ACTIONS"); v != "" {
		t.Skipf("this test takes too long to complete in Github actions")
	}

	ctx, cancel := context.WithCancel(t.Context())

	addr := servertest.GetRandomListenAddress(t)

	t.Logf("got random listen address: %v", addr)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", addr), http.NoBody)
	require.NoError(t, err)

	errCh := make(chan error, 1)

	go func() {
		errCh <- run(ctx, []string{"-a", addr}, os.Stderr)
	}()

	time.Sleep(100 * time.Millisecond)

	err = servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
	assert.NoError(t, err)

	cancel()

	err = <-errCh
	t.Logf("run finished with %v", err)
	assert.NoError(t, err)
}

func TestRun_BadArgs(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	errCh := make(chan error, 1)

	go func() {
		errCh <- run(ctx, []string{"-zzz", "gibberish"}, io.Discard)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	err := <-errCh
	t.Logf("run finished with %v", err)
	assert.Error(t, err)
}

func TestRun_SignalImitation(t *testing.T) {
	if v := os.Getenv("GITHUB_ACTIONS"); v != "" {
		t.Skipf("this test takes too long to complete in Github actions")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()

	addr := servertest.GetRandomListenAddress(t)

	t.Logf("got random listen address: %v", addr)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", addr), http.NoBody)
	require.NoError(t, err)

	errCh := make(chan error, 1)

	go func() {
		errCh <- run(ctx, []string{"-a", addr}, os.Stderr)
	}()

	time.Sleep(100 * time.Millisecond)

	err = servertest.MakeRequestDiscardResponse(http.DefaultClient, req)
	assert.NoError(t, err)

	err = <-errCh
	t.Logf("run finished with %v", err)
	assert.NoError(t, err)
}
