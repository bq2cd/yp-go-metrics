package main

import (
	"net/url"
	"testing"
	"time"

	config "github.com/bq2cd/yp-go-metrics/internal/config/agent"
	"github.com/stretchr/testify/assert"
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
