package sqlstorage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sqlHandlerImpl_Select(t *testing.T) {
	type args struct {
		ctx      context.Context
		selector sqlSelector
		itemIds  []string
	}
	type want struct {
		got []sqlItem
		err error
	}
	type testcase struct {
		h    sqlHandler
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.h.Select(tt.args.ctx, tt.args.selector, tt.args.itemIds...)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlHandlerImpl_Insert(t *testing.T) {
	type args struct {
		ctx    context.Context
		execer sqlExecer
		items  []sqlItem
	}
	type want struct {
		got sql.Result
		err error
	}
	type testcase struct {
		h    sqlHandler
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.h.Insert(tt.args.ctx, tt.args.execer, tt.args.items...)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlHandlerImpl_ConvertMetrics(t *testing.T) {
	type args struct {
		metrics model.MetricSet
	}
	type want struct {
		got []sqlItem
	}
	type testcase struct {
		h    sqlHandler
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.h.ConvertMetrics(tt.args.metrics)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
