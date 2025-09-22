package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:8080"}, PollInterval: defaultPollIntervalSec * time.Second, ReportInterval: defaultReportIntervalSec * time.Second},
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
			args: args{args: []string{"-r", "not-a-string"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 3a",
			args: args{args: []string{"-r", "-10"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 3b",
			args: args{args: []string{"-r", "0"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 4",
			args: args{args: []string{"-p", "not-a-string"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 4a",
			args: args{args: []string{"-p", "-2"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 4b",
			args: args{args: []string{"-p", "0"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "bad args 5",
			args: args{args: []string{"-r=2", "-p=10"}},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "good args",
			args: args{args: []string{"-a=localhost:9090"}},
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:9090"}, PollInterval: defaultPollIntervalSec * time.Second, ReportInterval: defaultReportIntervalSec * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 2",
			args: args{args: []string{"-a=localhost:9090", "-r=20"}},
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:9090"}, PollInterval: defaultPollIntervalSec * time.Second, ReportInterval: 20 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "good args 3",
			args: args{args: []string{"-a=localhost:9090", "-r=20", "-p=5"}},
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:9090"}, PollInterval: 5 * time.Second, ReportInterval: 20 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args.args)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type mockServer struct {
	mock.Mock
	t *testing.T
}

func (m *mockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called()
	w.Header().Set("content-type", "test/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 1_500*time.Millisecond)

	m := &mockServer{t: t}
	m.On("ServeHTTP", mock.Anything, mock.Anything).Return()

	ts := httptest.NewServer(m)

	errCh := make(chan error, 1)

	go func() {
		errCh <- run(ctx, []string{"-a", ts.URL, "-p=1", "-r=1"})
	}()

	ticker := time.NewTicker(100 * time.Millisecond)

	var err error

loop:
	for {
		select {
		case <-ticker.C:
			if len(m.Calls) > 0 {
				break loop
			}
		case <-ctx.Done():
			break loop
		}
	}

	m.AssertExpectations(t)

	cancel()

	err = <-errCh
	t.Logf("run finished with %v", err)
	assert.NoError(t, err)
}

func TestRun_BadArgs(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	errCh := make(chan error, 1)

	go func() {
		errCh <- run(ctx, []string{"-zzz", "gibberish"})
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	err := <-errCh
	t.Logf("run finished with %v", err)
	assert.Error(t, err)
}

func TestRun_SignalImitation(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)

	m := &mockServer{t: t}

	ts := httptest.NewServer(m)

	errCh := make(chan error, 1)

	go func() {
		errCh <- run(ctx, []string{"-a", ts.URL, "-p=1", "-r=1"})
	}()

	ticker := time.NewTicker(100 * time.Millisecond)

	var err error

loop:
	for {
		select {
		case <-ticker.C:
			if len(m.Calls) > 0 {
				break loop
			}
		case <-ctx.Done():
			break loop
		}
	}

	m.AssertNotCalled(t, "ServeHTTP")

	cancel()

	err = <-errCh
	t.Logf("run finished with %v", err)
	assert.NoError(t, err)
}
