package service

import (
	"io"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricSnapshotter(t *testing.T) {
	type args struct {
		storer  MetricStorer
		encoder MetricEncoder
		decoder MetricDecoder
	}
	tests := []struct {
		name string
		args args
		want *metricSnapshotter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetricSnapshotter(tt.args.storer, tt.args.encoder, tt.args.decoder))
		})
	}
}

func Test_metricSnapshotter_markDirty(t *testing.T) {
	type fields struct {
		MetricStorer MetricStorer
		encoder      MetricEncoder
		decoder      MetricDecoder
		notifyCh     chan struct{}
	}
	type args struct {
		numWrites int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				encoder:      tt.fields.encoder,
				decoder:      tt.fields.decoder,
				notifyCh:     tt.fields.notifyCh,
			}
			p.markDirty(tt.args.numWrites)
		})
	}
}

func Test_metricSnapshotter_StoreSingle(t *testing.T) {
	type fields struct {
		MetricStorer MetricStorer
		encoder      MetricEncoder
		decoder      MetricDecoder
		notifyCh     chan struct{}
	}
	type args struct {
		m model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				encoder:      tt.fields.encoder,
				decoder:      tt.fields.decoder,
				notifyCh:     tt.fields.notifyCh,
			}
			tt.assertion(t, p.StoreSingle(tt.args.m))
		})
	}
}

func Test_metricSnapshotter_StoreBatch(t *testing.T) {
	type fields struct {
		MetricStorer MetricStorer
		encoder      MetricEncoder
		decoder      MetricDecoder
		notifyCh     chan struct{}
	}
	type args struct {
		metrics []model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				encoder:      tt.fields.encoder,
				decoder:      tt.fields.decoder,
				notifyCh:     tt.fields.notifyCh,
			}
			tt.assertion(t, p.StoreBatch(tt.args.metrics))
		})
	}
}

func Test_metricSnapshotter_DumpClose(t *testing.T) {
	type fields struct {
		MetricStorer MetricStorer
		encoder      MetricEncoder
		decoder      MetricDecoder
		notifyCh     chan struct{}
	}
	type args struct {
		w io.WriteCloser
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				encoder:      tt.fields.encoder,
				decoder:      tt.fields.decoder,
				notifyCh:     tt.fields.notifyCh,
			}
			tt.assertion(t, p.DumpClose(tt.args.w))
		})
	}
}

func Test_metricSnapshotter_LoadClose(t *testing.T) {
	type fields struct {
		MetricStorer MetricStorer
		encoder      MetricEncoder
		decoder      MetricDecoder
		notifyCh     chan struct{}
	}
	type args struct {
		r io.ReadCloser
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				encoder:      tt.fields.encoder,
				decoder:      tt.fields.decoder,
				notifyCh:     tt.fields.notifyCh,
			}
			tt.assertion(t, p.LoadClose(tt.args.r))
		})
	}
}

func Test_metricSnapshotter_C(t *testing.T) {
	type fields struct {
		MetricStorer MetricStorer
		encoder      MetricEncoder
		decoder      MetricDecoder
		notifyCh     chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   <-chan struct{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				encoder:      tt.fields.encoder,
				decoder:      tt.fields.decoder,
				notifyCh:     tt.fields.notifyCh,
			}
			assert.Equal(t, tt.want, p.C())
		})
	}
}
