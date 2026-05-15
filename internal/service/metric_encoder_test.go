package service

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type mockMetricEncoder struct {
	mock.Mock

	err error
}

func (m *mockMetricEncoder) EncodeBatch(w io.Writer, metrics []model.Metric) error {
	m.Called(w, metrics)
	return m.err
}

func TestNewMetricJSONEncoder(t *testing.T) {
	tests := []struct {
		name string
		want *metricJSONEncoder
	}{
		{
			name: "default",
			want: &metricJSONEncoder{},
		},
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
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "empty metrics (nil)",
			args:      args{metrics: nil},
			want:      `[]`,
			assertion: assert.NoError,
		},
		{
			name:      "empty metrics",
			args:      args{metrics: []model.Metric{}},
			want:      `[]`,
			assertion: assert.NoError,
		},
		{
			name: "single metric",
			args: args{metrics: []model.Metric{
				model.NewCounterMetric("id1", 123),
			}},
			want: `[
				{"id": "id1", "type": "counter", "delta": 123}
				]`,
			assertion: assert.NoError,
		},
		{
			name: "multiple metrics",
			args: args{metrics: []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewGaugeMetric("id2", -1.23),
				model.NewCounterMetric("id3", -456),
				model.NewGaugeMetric("id4", 4.56),
			}},
			want: `[
				{"id": "id1", "type": "counter", "delta": 123},
				{"id": "id2", "type": "gauge", "value": -1.23},
				{"id": "id3", "type": "counter", "delta": -456},
				{"id": "id4", "type": "gauge", "value": 4.56}
				]`,
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &metricJSONEncoder{}
			w := &bytes.Buffer{}
			tt.assertion(t, d.EncodeBatch(w, tt.args.metrics))
			assert.JSONEq(t, tt.want, w.String())
		})
	}
}
