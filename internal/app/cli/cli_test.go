package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_Run(t *testing.T) {
	type args struct {
		ctx    context.Context
		logger log.Logger
		args   []string
	}
	type want struct {
		gotStderr string
		err       error
	}
	type testcase struct {
		a    App[any]
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			stderr := &bytes.Buffer{}
			err := tt.a.Run(tt.args.ctx, tt.args.logger, tt.args.args, stderr)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.gotStderr, stderr.String())
		})
	}
}
