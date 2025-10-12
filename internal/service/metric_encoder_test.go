package service

import (
	"bytes"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricJSONEncoder(t *testing.T) {
	tests := []struct {
		name string
		want *metricJSONEncoder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetricJSONEncoder())
		})
	}
}

func Test_metricJSONEncoder_EncodeBatch(t *testing.T) {
	type args struct {
		metrics []model.Metric
	}
	tests := []struct {
		name      string
		d         *metricJSONEncoder
		args      args
		wantW     string
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &metricJSONEncoder{}
			w := &bytes.Buffer{}
			tt.assertion(t, d.EncodeBatch(w, tt.args.metrics))
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}
