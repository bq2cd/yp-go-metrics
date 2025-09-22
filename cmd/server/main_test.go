package main

import (
	"testing"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
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
			got, err := parseArgs(tt.args.args)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
