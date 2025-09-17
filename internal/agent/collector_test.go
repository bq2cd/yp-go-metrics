package agent

import (
	"maps"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCollector struct {
	mock.Mock
	metrics []model.Metric
}

func (m *mockCollector) Collect() ([]model.Metric, error) {
	m.Called()
	return m.metrics, nil
}

func Test_defaultCollector_Collect(t *testing.T) {
	type want struct {
		metricIDToType map[string]model.MetricType
	}
	tests := []struct {
		name      string
		want      want
		assertion func(assert.TestingT, want, []model.Metric)
	}{
		{
			name: "default metrics",
			want: want{
				metricIDToType: func() map[string]model.MetricType {
					m := make(map[string]model.MetricType, len(defaultRuntimeMetrics)+len(defaultExtraMetrics))
					maps.Copy(m, defaultRuntimeMetrics)
					maps.Copy(m, defaultExtraMetrics)
					return m
				}(),
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
			c := &defaultCollector{}
			got, err := c.Collect()
			assert.NoError(t, err)
			tt.assertion(t, tt.want, got)
		})
	}
}

func TestNewDefaultCollector(t *testing.T) {
	tests := []struct {
		name string
		want *defaultCollector
	}{
		{
			name: "default initialisation",
			want: &defaultCollector{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewDefaultCollector())
		})
	}
}
