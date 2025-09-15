package service

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	type args struct {
		storage repository.Storage
	}
	tests := []struct {
		name string
		args args
		want *metricService
	}{
		{
			name: "new service",
			args: args{storage: repository.NewMemStorage()},
			want: &metricService{storage: repository.NewMemStorage()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMetrics(tt.args.storage))
		})
	}
}

func Test_metricService_Update(t *testing.T) {
	type fields struct {
		storage func() repository.Storage
	}
	type args struct {
		metric model.Metric
	}
	type want struct {
		metric model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *want
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty storage",
			fields: fields{storage: func() repository.Storage { return repository.NewMemStorage() }},
			args:   args{metric: model.NewCounterMetric("id1", 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "new counter",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewCounterMetric("id1", 5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewCounterMetric("id2", 10)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "existing counter",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewCounterMetric("id1", 5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewCounterMetric("id1", 10)},
			want: &want{metric: model.NewCounterMetric("id1", 15)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "new gauge",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewGaugeMetric("id1", 1.5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewGaugeMetric("id2", 5.1)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "existing gauge",
			fields: fields{storage: func() repository.Storage {
				storage := repository.NewMemStorage()
				err := storage.Set(model.NewGaugeMetric("id1", 1.5))
				require.NoError(t, err)
				return storage
			}},
			args: args{metric: model.NewGaugeMetric("id1", -1.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				storage: tt.fields.storage(),
			}
			tt.assertion(t, s.Update(tt.args.metric))
			got, err := s.storage.Get(tt.args.metric.Hash)
			require.NoError(t, err)
			if tt.want != nil {
				assert.Equal(t, tt.want.metric, got)
			} else {
				assert.Equal(t, tt.args.metric, got)
			}
		})
	}
}
