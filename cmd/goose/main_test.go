// Custom goose binary as per https://github.com/pressly/goose/blob/main/examples/go-migrations/main.go
package main

import (
	"bytes"
	"context"
	"flag"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func Test_parseArgs(t *testing.T) {
	type args struct {
		fs        *flag.FlagSet
		args      []string
		envParser envparser.Parser
	}
	type want struct {
		got cliOptions
		err error
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseArgs(tt.args.fs, tt.args.args, tt.args.envParser)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_launchProcess(t *testing.T) {
	type args struct {
		ctx    context.Context
		logger log.Logger
		opts   cliOptions
	}
	type want struct {
		err error
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := launchProcess(tt.args.ctx, tt.args.logger, tt.args.opts)
			require.Equal(t, tt.want.err, err)
		})
	}
}

func Test_run(t *testing.T) {
	type args struct {
		ctx  context.Context
		args []string
	}
	type want struct {
		gotStderr string
		err       error
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			stderr := &bytes.Buffer{}
			err := run(tt.args.ctx, tt.args.args, stderr)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.gotStderr, stderr.String())
		})
	}
}

func Test_main(t *testing.T) {
	// should be covered by [Test_run]
	t.SkipNow()
}
