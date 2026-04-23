package memstats

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type mockMemStats struct {
	Counter1    uint64
	Counter2    int64
	Counter3    int
	Gauge1      float64
	Gauge2      float32
	BadCounter  string
	BadGauge    string
	UnknownType int
}

func (m *mockMemStats) ReadStats() reflect.Value {
	return reflect.ValueOf(m).Elem()
}

func TestGetSupportedMetrics(t *testing.T) {
	tests := []struct {
		name string
		want func() map[string]model.MetricType
	}{
		{
			name: "multiple calls return the same",
			want: func() map[string]model.MetricType {
				return GetSupportedMetrics()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want(), GetSupportedMetrics())
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *source
	}{
		{
			name: "default initialisation",
			want: &source{supportedMetrics: GetSupportedMetrics(), reader: &memStats{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New())
		})
	}
}

func Test_source_ReadMetrics(t *testing.T) {
	type args struct {
		supportedMetrics map[string]model.MetricType
		reader           memStatsReader
	}
	type want struct {
		metricIDToType map[string]model.MetricType
	}
	tests := []struct {
		name      string
		args      args
		want      want
		assertion func(assert.TestingT, want, []model.Metric)
	}{
		{
			name: "default metrics collected",
			args: args{
				supportedMetrics: GetSupportedMetrics(),
				reader:           &memStats{},
			},
			want: want{
				metricIDToType: GetSupportedMetrics(),
			},
			assertion: func(t assert.TestingT, want want, got []model.Metric) {
				assert.Len(t, got, len(want.metricIDToType))
				metricIDToType := make(map[string]model.MetricType, len(want.metricIDToType))
				for i := range got {
					m := got[i]
					metricIDToType[m.ID] = m.Type
				}
				for mID, mType := range want.metricIDToType {
					assert.Contains(t, metricIDToType, mID)
					assert.Equal(t, mType, metricIDToType[mID])
				}
			},
		},
		{
			name: "custom metrics collected",
			args: args{
				supportedMetrics: map[string]model.MetricType{
					"Counter1":    model.MetricTypeCounter,
					"Counter2":    model.MetricTypeCounter,
					"Counter3":    model.MetricTypeCounter,
					"Gauge1":      model.MetricTypeGauge,
					"Gauge2":      model.MetricTypeGauge,
					"BadCounter":  model.MetricTypeCounter,
					"BadGauge":    model.MetricTypeGauge,
					"UnknownType": model.MetricType("unknown"),
				},
				reader: &mockMemStats{Counter1: 35, Counter2: -35, Counter3: 29, Gauge1: 3.75, Gauge2: -8.21, BadCounter: "123", BadGauge: "1.23", UnknownType: -1},
			},
			want: want{
				metricIDToType: map[string]model.MetricType{
					"Counter1": model.MetricTypeCounter,
					"Counter2": model.MetricTypeCounter,
					"Counter3": model.MetricTypeCounter,
					"Gauge1":   model.MetricTypeGauge,
					"Gauge2":   model.MetricTypeGauge,
				},
			},
			assertion: func(t assert.TestingT, want want, got []model.Metric) {
				assert.Len(t, got, len(want.metricIDToType))
				metricIDToType := make(map[string]model.MetricType, len(want.metricIDToType))
				for i := range got {
					m := got[i]
					metricIDToType[m.ID] = m.Type
				}
				for mID, mType := range want.metricIDToType {
					assert.Contains(t, metricIDToType, mID)
					assert.Equal(t, mType, metricIDToType[mID])
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &source{
				supportedMetrics: tt.args.supportedMetrics,
				reader:           tt.args.reader,
			}
			got, err := s.ReadMetrics()
			require.NoError(t, err)
			tt.assertion(t, tt.want, got)
		})
	}
}

func Test_castToInt64(t *testing.T) {
	type args struct {
		v reflect.Value
	}
	tests := []struct {
		name string
		args args
		want int64
		ok   bool
	}{
		// ok
		{
			name: "int ok",
			args: args{v: reflect.ValueOf(int(-1))},
			want: int64(-1),
			ok:   true,
		},
		{
			name: "int8 ok",
			args: args{v: reflect.ValueOf(int8(-1))},
			want: int64(-1),
			ok:   true,
		},
		{
			name: "int16 ok",
			args: args{v: reflect.ValueOf(int16(-1))},
			want: int64(-1),
			ok:   true,
		},
		{
			name: "int32 ok",
			args: args{v: reflect.ValueOf(int32(-1))},
			want: int64(-1),
			ok:   true,
		},
		{
			name: "int64 ok",
			args: args{v: reflect.ValueOf(int64(-1))},
			want: int64(-1),
			ok:   true,
		},
		{
			name: "uint ok",
			args: args{v: reflect.ValueOf(uint(1))},
			want: int64(1),
			ok:   true,
		},
		{
			name: "uint8 ok",
			args: args{v: reflect.ValueOf(uint8(1))},
			want: int64(1),
			ok:   true,
		},
		{
			name: "uint16 ok",
			args: args{v: reflect.ValueOf(uint16(1))},
			want: int64(1),
			ok:   true,
		},
		{
			name: "uint32 ok",
			args: args{v: reflect.ValueOf(uint32(1))},
			want: int64(1),
			ok:   true,
		},
		{
			name: "uint64 ok",
			args: args{v: reflect.ValueOf(uint64(1))},
			want: int64(1),
			ok:   true,
		},
		// not ok
		{
			name: "float32 not ok",
			args: args{v: reflect.ValueOf(float32(1))},
			want: 0,
			ok:   false,
		},
		{
			name: "float64 not ok",
			args: args{v: reflect.ValueOf(float64(1))},
			want: 0,
			ok:   false,
		},
		{
			name: "string not ok",
			args: args{v: reflect.ValueOf("1")},
			want: 0,
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := castToInt64(tt.args.v)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func Test_castToFloat64(t *testing.T) {
	type args struct {
		v reflect.Value
	}
	tests := []struct {
		name string
		args args
		want float64
		ok   bool
	}{
		// ok
		{
			name: "int64 ok",
			args: args{v: reflect.ValueOf(int64(-1))},
			want: float64(-1),
			ok:   true,
		},
		{
			name: "uint64 ok",
			args: args{v: reflect.ValueOf(uint64(1))},
			want: float64(1),
			ok:   true,
		},
		{
			name: "float64 ok",
			args: args{v: reflect.ValueOf(float64(-3.7))},
			want: float64(-3.7),
			ok:   true,
		},
		{
			name: "float32 ok",
			args: args{v: reflect.ValueOf(float32(-3.7))},
			want: float64(float32(-3.7)),
			ok:   true,
		},
		// not ok
		{
			name: "string not ok",
			args: args{v: reflect.ValueOf("1.2")},
			want: 0,
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := castToFloat64(tt.args.v)
			if tt.want != 0 {
				assert.InEpsilon(t, tt.want, got, 1e-10)
			} else {
				assert.Zero(t, got)
			}
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func Test_memStats_ReadStats(t *testing.T) {
	tests := []struct {
		name string
		m    *memStats
		want reflect.Value
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &memStats{}
			assert.Equal(t, tt.want, m.ReadStats())
		})
	}
}
