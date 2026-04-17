package psutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

func TestGetSupportedMetrics(t *testing.T) {
	type want struct {
		got func() map[string]model.MetricType
	}
	type testcase struct {
		want want
	}
	tests := map[string]testcase{
		"multiple calls return the same value": {
			want: want{
				got: func() map[string]model.MetricType {
					return GetSupportedMetrics()
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := GetSupportedMetrics()
			assert.Equal(t, tt.want.got(), got)
		})
	}
}

func TestNew(t *testing.T) {
	type want struct {
		got *source
	}
	type testcase struct {
		want want
	}
	tests := map[string]testcase{
		"default": {
			want: want{got: &source{}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := New()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_source_ReadMetrics(t *testing.T) {
	type want struct {
		got     map[string]model.MetricType
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		want want
	}
	tests := map[string]testcase{
		"default metric collected": {
			want: want{
				got: GetSupportedMetrics(),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := New()
			got, err := s.ReadMetrics()
			tt.want.wantErr(t, err)
			metricIDToType := make(map[string]model.MetricType)
			for i := range got {
				m := got[i]
				metricIDToType[m.ID] = m.Type
			}
			for mID, mType := range tt.want.got {
				assert.Contains(t, metricIDToType, mID)
				assert.Equal(t, mType, metricIDToType[mID])
			}
		})
	}
}
