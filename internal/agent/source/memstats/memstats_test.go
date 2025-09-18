package memstats

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

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
			want: &source{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New())
		})
	}
}

func Test_source_ReadMetrics(t *testing.T) {
	type want struct {
		metricIDToType map[string]model.MetricType
	}
	tests := []struct {
		name      string
		s         *source
		want      want
		assertion func(assert.TestingT, want, []model.Metric)
	}{
		{
			name: "all metrics collected",
			want: want{
				metricIDToType: supportedMetrics,
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
			s := &source{}
			got, err := s.ReadMetrics()
			assert.NoError(t, err)
			tt.assertion(t, tt.want, got)
		})
	}
}
