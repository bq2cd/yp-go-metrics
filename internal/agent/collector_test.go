package agent

import (
	"maps"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/agent/source"
	"github.com/bq2cd/yp-go-metrics/internal/agent/source/extra"
	"github.com/bq2cd/yp-go-metrics/internal/agent/source/memstats"
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
	type args struct {
		collector Collector
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
			name: "default metrics",
			args: args{collector: &defaultCollector{sources: []source.Source{memstats.New(), extra.New()}}},
			want: want{
				metricIDToType: func() map[string]model.MetricType {
					m := make(map[string]model.MetricType)
					maps.Copy(m, memstats.GetSupportedMetrics())
					maps.Copy(m, extra.GetSupportedMetrics())
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
			got, err := tt.args.collector.Collect()
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
			want: &defaultCollector{sources: source.DefaultSources()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewDefaultCollector())
		})
	}
}
