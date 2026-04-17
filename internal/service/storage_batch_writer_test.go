package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

func Test_metricBatchTx_Result(t *testing.T) {
	type fields struct {
		ctx   context.Context
		batch MetricBatch
		errCh chan error
	}
	type want struct {
		got <-chan error
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &metricBatchTx{
				ctx:   tt.fields.ctx,
				batch: tt.fields.batch,
				errCh: tt.fields.errCh,
			}
			got := b.Result()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestNewStorageBatchWriter(t *testing.T) {
	type args struct {
		storage repository.StorageMulti
	}
	type want struct {
		got *storageBatchWriter
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
			got := NewStorageBatchWriter(tt.args.storage)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_storageBatchWriter_WriteBatch(t *testing.T) {
	type fields struct {
		storage    repository.StorageMulti
		incomingCh chan *metricBatchTx
	}
	type args struct {
		ctx   context.Context
		batch MetricBatch
	}
	type want struct {
		got MetricBatchTx
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w := &storageBatchWriter{
				storage:    tt.fields.storage,
				incomingCh: tt.fields.incomingCh,
			}
			got := w.WriteBatch(tt.args.ctx, tt.args.batch)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_storageBatchWriter_StartProcessing(t *testing.T) {
	type fields struct {
		storage    repository.StorageMulti
		incomingCh chan *metricBatchTx
	}
	type args struct {
		ctx context.Context
	}
	type testcase struct {
		fields fields
		args   args
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w := &storageBatchWriter{
				storage:    tt.fields.storage,
				incomingCh: tt.fields.incomingCh,
			}
			w.StartProcessing(tt.args.ctx)
		})
	}
}
