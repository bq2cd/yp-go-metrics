package service

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type mockMetricDecoder struct {
	mock.Mock

	metrics []model.Metric
	err     error
}

func (m *mockMetricDecoder) DecodeBatch(r io.Reader) ([]model.Metric, error) {
	m.Called(r)
	return m.metrics, m.err
}

func TestNewMetricJSONDecoder(t *testing.T) {
	tests := []struct {
		name string
		want *metricJSONDecoder
	}{
		{
			name: "default",
			want: &metricJSONDecoder{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetricJSONDecoder())
		})
	}
}

func Test_metricJSONDecoder_DecodeBatch(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name      string
		args      args
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty reader",
			args: args{
				data: ``,
			},
			want:      []model.Metric{},
			assertion: assert.Error,
		},
		{
			name: "invalid json",
			args: args{
				data: `{}`,
			},
			want:      []model.Metric{},
			assertion: assert.Error,
		},
		{
			name: "empty slice",
			args: args{
				data: `[]`,
			},
			want:      []model.Metric{},
			assertion: assert.NoError,
		},
		{
			name: "single metric",
			args: args{
				data: `[{"id": "id1", "type": "counter", "delta": 123}]`,
			},
			want:      []model.Metric{model.NewCounterMetric("id1", 123)},
			assertion: assert.NoError,
		},
		{
			name: "multiple metrics",
			args: args{
				data: `[
					{"id": "id1", "type": "counter", "delta": 123},
					{"id": "id2", "type": "gauge", "value": -1.23},
					{"id": "id3", "type": "counter", "delta": -456},
					{"id": "id4", "type": "gauge", "value": 4.56}
					]`,
			},
			want: []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewGaugeMetric("id2", -1.23),
				model.NewCounterMetric("id3", -456),
				model.NewGaugeMetric("id4", 4.56),
			},
			assertion: assert.NoError,
		},
		{
			name: "multiple metrics, some invalid",
			args: args{
				data: `[
					{"id": "id1", "type": "counter", "delta": 123},
					{"id": "id2", "type": "gauge", "value": -1.23},
					{"id": "id3", "type": "counter", "value": -456},
					{"id": "id4", "type": "gauge", "value": 4.56}
					]`,
			},
			want: []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewGaugeMetric("id2", -1.23),
				func() model.Metric {
					var v float64 = -456
					return model.Metric{Type: model.MetricTypeCounter, ID: "id3", Value: &v}
				}(),
				model.NewGaugeMetric("id4", 4.56),
			},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &metricJSONDecoder{}
			r := bytes.NewBufferString(tt.args.data)
			got, err := d.DecodeBatch(r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
