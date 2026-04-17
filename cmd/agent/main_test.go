package main

import (
	"context"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
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
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: defaultUpstreamURL}, PollInterval: defaultPollIntervalSec * time.Second, ReportInterval: defaultReportIntervalSec * time.Second},
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
			name:      "bad args, missing secret key value",
			args:      args{args: []string{"-k"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, sender pool size negative",
			args:      args{args: []string{"-l=-2"}},
			want:      config.Config{},
			assertion: assert.Error,
		},
		{
			name:      "bad args, sender pool size missing",
			args:      args{args: []string{"-l"}},
			want:      config.Config{},
			assertion: assert.Error,
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
		{
			name: "good args, secret key",
			args: args{args: []string{"-k=MTIz"}},
			want: config.Config{
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				HMACSecretKey:  []byte(`123`),
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, empty secret key is valid",
			args: args{args: []string{"-k="}},
			want: config.Config{
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				HMACSecretKey:  nil,
			},
			assertion: assert.NoError,
		},
		{
			name: "good args, sender pool size (aka rate limit)",
			args: args{args: []string{"-l=2"}},
			want: config.Config{
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				SenderPoolSize: 2,
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
				args: []string{"-a=localhost:9090", "-r=20", "-p=5"},
				env:  map[string]string{"ADDRESS": "localhost:3333"},
			},
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:3333"}, PollInterval: 5 * time.Second, ReportInterval: 20 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "env overrides poll interval",
			args: args{
				args: []string{"-a=localhost:9090", "-r=20", "-p=5"},
				env:  map[string]string{"POLL_INTERVAL": "19"},
			},
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:9090"}, PollInterval: 19 * time.Second, ReportInterval: 20 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "env overrides report interval",
			args: args{
				args: []string{"-a=localhost:9090", "-r=20", "-p=5"},
				env:  map[string]string{"REPORT_INTERVAL": "81"},
			},
			want: config.Config{UpstreamURL: url.URL{Scheme: "http", Host: "localhost:9090"}, PollInterval: 5 * time.Second, ReportInterval: 81 * time.Second},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "env invalid report interval",
			args: args{
				args: []string{"-a=localhost:9090", "-r=20", "-p=5"},
				env:  map[string]string{"REPORT_INTERVAL": "-30"},
			},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "env invalid poll interval",
			args: args{
				args: []string{"-a=localhost:9090", "-r=20", "-p=5"},
				env:  map[string]string{"POLL_INTERVAL": "not a number"},
			},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "env overrides secret key",
			args: args{
				args: []string{"-k=MTIz"},
				env:  map[string]string{"KEY": "NDU2"},
			},
			want: config.Config{
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				HMACSecretKey:  []byte(`456`),
			},
			assertion: assert.NoError,
		},
		{
			name: "env sets secret key",
			args: args{
				args: []string{},
				env:  map[string]string{"KEY": "NDU2"},
			},
			want: config.Config{
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				HMACSecretKey:  []byte(`456`),
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
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				HMACSecretKey:  []byte(`456`),
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
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				HMACSecretKey:  []byte(`123`),
			},
			assertion: assert.NoError,
		},
		{
			name: "env negative sender pool size (rate limit)",
			args: args{
				args: []string{"-l=2"},
				env:  map[string]string{"RATE_LIMIT": "-3"},
			},
			want: config.Config{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "env overrides sender pool size (rate limit)",
			args: args{
				args: []string{"-l=18"},
				env:  map[string]string{"RATE_LIMIT": "3"},
			},
			want: config.Config{
				UpstreamURL:    url.URL{Scheme: "http", Host: defaultUpstreamURL},
				PollInterval:   defaultPollIntervalSec * time.Second,
				ReportInterval: defaultReportIntervalSec * time.Second,
				SenderPoolSize: 3,
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
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

type mockServer struct {
	mock.Mock
	mu sync.RWMutex
}

func (m *mockServer) NumCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Calls)
}

func (m *mockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.Called()
	m.mu.Unlock()
	httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`[]`))
}

func Test_run(t *testing.T) {
	type args struct {
		stderr io.Writer
	}
	type setup struct {
		contextFunc    func() (context.Context, context.CancelFunc)
		mockServerFunc func() (*httptest.Server, *mockServer)
		argsFunc       func(*httptest.Server) []string
	}
	tests := []struct {
		name          string
		skip          func() (bool, string)
		args          args
		setup         setup
		assertionMock func(*testing.T, *mockServer)
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
					return context.WithTimeout(t.Context(), 1_500*time.Millisecond)
				},
				mockServerFunc: func() (*httptest.Server, *mockServer) {
					m := &mockServer{}
					m.On("ServeHTTP", mock.Anything, mock.Anything).Return()
					ts := httptest.NewServer(m)
					return ts, m
				},
				argsFunc: func(ts *httptest.Server) []string {
					return []string{"-a", ts.URL, "-r=1", "-p=1"}
				},
			},
			assertionMock: func(t *testing.T, m *mockServer) {
				m.AssertExpectations(t)
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
				mockServerFunc: func() (*httptest.Server, *mockServer) {
					m := &mockServer{}
					m.On("ServeHTTP", mock.Anything, mock.Anything).Return()
					ts := httptest.NewServer(m)
					return ts, m
				},
				argsFunc: func(ts *httptest.Server) []string {
					return []string{"-a", ts.URL, "-r=1", "-p=1"}
				},
			},
			assertionMock: func(t *testing.T, m *mockServer) {
				m.AssertNotCalled(t, "ServeHTTP")
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
					return context.WithTimeout(t.Context(), 500*time.Millisecond)
				},
				mockServerFunc: func() (*httptest.Server, *mockServer) {
					m := &mockServer{}
					m.On("ServeHTTP", mock.Anything, mock.Anything).Return()
					ts := httptest.NewServer(m)
					return ts, m
				},
				argsFunc: func(ts *httptest.Server) []string {
					return []string{"-zzz", "gibberish"}
				},
			},
			assertionMock: func(t *testing.T, m *mockServer) {
				m.AssertNotCalled(t, "ServeHTTP")
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

			ts, m := tt.setup.mockServerFunc()
			defer ts.Close()

			args := tt.setup.argsFunc(ts)

			errCh := make(chan error, 1)

			go func() {
				errCh <- run(ctx, args, tt.args.stderr)
			}()

			ticker := time.NewTicker(100 * time.Millisecond)

			var err error

		loop:
			for {
				select {
				case <-ticker.C:
					if m.NumCalls() > 0 {
						break loop
					}
				case <-ctx.Done():
					break loop
				}
			}

			tt.assertionMock(t, m)

			cancel()

			err = <-errCh
			t.Logf("run finished with %v", err)

			tt.assertionErr(t, err)
		})
	}
}
