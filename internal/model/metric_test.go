package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCounterMetric(t *testing.T) {
	type args struct {
		mID   string
		value int64
	}
	tests := []struct {
		name string
		args args
		want func() Metric
	}{
		{
			name: "zero",
			args: args{mID: "id1", value: 0},
			want: func() Metric {
				var value int64 = 0
				return Metric{
					ID:    "id1",
					Type:  MetricTypeCounter,
					Delta: &value,
					Value: nil,
					Hash:  MetricHash("counter/id1"),
				}
			},
		},
		{
			name: "positive value",
			args: args{mID: "id1", value: 10},
			want: func() Metric {
				var value int64 = 10
				return Metric{
					ID:    "id1",
					Type:  MetricTypeCounter,
					Delta: &value,
					Value: nil,
					Hash:  MetricHash("counter/id1"),
				}
			},
		},
		{
			name: "negative value",
			args: args{mID: "id1", value: -10},
			want: func() Metric {
				var value int64 = -10
				return Metric{
					ID:    "id1",
					Type:  MetricTypeCounter,
					Delta: &value,
					Value: nil,
					Hash:  MetricHash("counter/id1"),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want(), NewCounterMetric(tt.args.mID, tt.args.value))
		})
	}
}

func TestNewGaugeMetric(t *testing.T) {
	type args struct {
		mID   string
		value float64
	}
	tests := []struct {
		name string
		args args
		want func() Metric
	}{
		{
			name: "zero",
			args: args{mID: "id1", value: 0.0},
			want: func() Metric {
				var value = 0.0
				return Metric{
					ID:    "id1",
					Type:  MetricTypeGauge,
					Delta: nil,
					Value: &value,
					Hash:  MetricHash("gauge/id1"),
				}
			},
		},
		{
			name: "positive value",
			args: args{mID: "id1", value: 9.99},
			want: func() Metric {
				var value = 9.99
				return Metric{
					ID:    "id1",
					Type:  MetricTypeGauge,
					Delta: nil,
					Value: &value,
					Hash:  MetricHash("gauge/id1"),
				}
			},
		},
		{
			name: "negative value",
			args: args{mID: "id1", value: -9.99},
			want: func() Metric {
				var value = -9.99
				return Metric{
					ID:    "id1",
					Type:  MetricTypeGauge,
					Delta: nil,
					Value: &value,
					Hash:  MetricHash("gauge/id1"),
				}
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want(), NewGaugeMetric(tt.args.mID, tt.args.value))
		})
	}
}

func TestMetric_updateHash(t *testing.T) {
	type fields struct {
		ID   string
		Type MetricType
		Hash MetricHash
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty initial hash",
			fields: fields{
				ID:   "id1",
				Type: MetricTypeCounter,
				Hash: "",
			},
			want: "counter/id1",
		},
		{
			name: "incorrect initial hash",
			fields: fields{
				ID:   "id1",
				Type: MetricTypeCounter,
				Hash: "badHash",
			},
			want: "counter/id1",
		},
		{
			name: "correct initial hash",
			fields: fields{
				ID:   "id1",
				Type: MetricTypeCounter,
				Hash: "counter/id1",
			},
			want: "counter/id1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				ID:   tt.fields.ID,
				Type: tt.fields.Type,
				Hash: tt.fields.Hash,
			}
			m.updateHash()
			assert.Equal(t, MetricHash(tt.want), m.Hash)
		})
	}
}
