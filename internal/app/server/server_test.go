package server

import (
	"context"
	"testing"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	type args struct {
		ctx    context.Context
		logger log.Logger
		cfg    config.Config
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
			err := Run(tt.args.ctx, tt.args.logger, tt.args.cfg)
			require.Equal(t, tt.want.err, err)
		})
	}
}
