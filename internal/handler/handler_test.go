package handler

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricFromURLPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    func() model.Metric
		wantErr error
	}{
		{
			name:    "empty path",
			args:    args{path: ""},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "wrong operation",
			args:    args{path: "/badOperation"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "missing operation",
			args:    args{path: "/ /someType/someID/someValue"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "missing type",
			args:    args{path: "/ / /someID/someValue"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "missing id",
			args:    args{path: "/ / / /someValue"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update without parameters",
			args:    args{path: "/update"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update wrong type without id",
			args:    args{path: "/update/badType"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update wrong type without value",
			args:    args{path: "/update/badType/someID"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update wrong type with bad value",
			args:    args{path: "/update/badType/someID/badValue"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update wrong type with correct value",
			args:    args{path: "/update/badType/someID/15.1"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update missing type without value",
			args:    args{path: "/update/ /myCounter"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update correct type without id",
			args:    args{path: "/update/counter"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update correct type without value",
			args:    args{path: "/update/counter/myCounter"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update correct type with bad value",
			args:    args{path: "/update/counter/myCounter/badValue"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update missing type with correct value",
			args:    args{path: "/update/ /myCounter/123"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update missing id with correct value",
			args:    args{path: "/update/counter/ /123"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name:    "update missing id with correct value (2)",
			args:    args{path: "/update/counter//123"},
			want:    func() model.Metric { return model.Metric{} },
			wantErr: ErrPathConversionError,
		},
		{
			name: "update counter type with correct positive value",
			args: args{path: "/update/counter/myCounter/123"},
			want: func() model.Metric {
				var value int64 = 123
				return model.Metric{ID: "myCounter", Type: model.MetricTypeCounter, Delta: &value, Value: nil, Hash: model.MetricHash("counter/myCounter")}
			},
			wantErr: nil,
		},
		{
			name: "update counter type with correct positive value (no leading slash)",
			args: args{path: "update/counter/myCounter/123"},
			want: func() model.Metric {
				var value int64 = 123
				return model.Metric{ID: "myCounter", Type: model.MetricTypeCounter, Delta: &value, Value: nil, Hash: model.MetricHash("counter/myCounter")}
			},
			wantErr: nil,
		},
		{
			name: "update counter type with correct negative value",
			args: args{path: "/update/counter/myCounter/-456"},
			want: func() model.Metric {
				var value int64 = -456
				return model.Metric{ID: "myCounter", Type: model.MetricTypeCounter, Delta: &value, Value: nil, Hash: model.MetricHash("counter/myCounter")}
			},
			wantErr: nil,
		},
		{
			name: "update gauge type with correct positive value",
			args: args{path: "/update/gauge/myGauge/1.23"},
			want: func() model.Metric {
				var value = 1.23
				return model.Metric{ID: "myGauge", Type: model.MetricTypeGauge, Delta: nil, Value: &value, Hash: model.MetricHash("gauge/myGauge")}
			},
			wantErr: nil,
		},
		{
			name: "update gauge type with correct negative value",
			args: args{path: "/update/gauge/myGauge/-4.56"},
			want: func() model.Metric {
				var value = -4.56
				return model.Metric{ID: "myGauge", Type: model.MetricTypeGauge, Delta: nil, Value: &value, Hash: model.MetricHash("gauge/myGauge")}
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetricFromURLPath(tt.args.path)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.want(), got)
		})
	}
}
