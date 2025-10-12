package service

import (
	"io"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricJSONDecoder(t *testing.T) {
	tests := []struct {
		name string
		want *metricJSONDecoder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetricJSONDecoder())
		})
	}
}

func Test_metricJSONDecoder_DecodeBatch(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name      string
		d         *metricJSONDecoder
		args      args
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &metricJSONDecoder{}
			got, err := d.DecodeBatch(tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
