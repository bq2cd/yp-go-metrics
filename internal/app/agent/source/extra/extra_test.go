package extra

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/model"
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
	type fields struct {
		pollCounter int
	}
	type want struct {
		metricIDToType map[string]model.MetricType
		pollCounter    int
	}
	tests := []struct {
		name      string
		fields    fields
		want      want
		assertion func(assert.TestingT, want, []model.Metric, *source)
	}{
		{
			name:   "all metrics collected",
			fields: fields{pollCounter: 1},
			want: want{
				metricIDToType: supportedMetrics,
				pollCounter:    2,
			},
			assertion: func(t assert.TestingT, want want, got []model.Metric, s *source) {
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
				assert.Equal(t, want.pollCounter, s.pollCounter)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &source{
				pollCounter: tt.fields.pollCounter,
			}
			got, err := s.ReadMetrics()
			require.NoError(t, err)
			tt.assertion(t, tt.want, got, s)
		})
	}
}
