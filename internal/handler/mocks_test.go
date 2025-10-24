package handler

import (
	"context"
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type faultyMetricService struct{}

func (s *faultyMetricService) StoreSingle(ctx context.Context, metric model.Metric) error {
	return fmt.Errorf("faulty storage set error")
}

func (s *faultyMetricService) StoreBatch(ctx context.Context, metrics []model.Metric) error {
	return fmt.Errorf("faulty storage set error")
}

func (s *faultyMetricService) RetrieveSingle(ctx context.Context, key model.MetricKey) (model.Metric, error) {
	return model.Metric{}, fmt.Errorf("faulty storage get error")
}

func (s *faultyMetricService) RetrieveBatch(ctx context.Context, keys []model.MetricKey) ([]model.Metric, error) {
	return nil, fmt.Errorf("faulty storage get error")
}

func (s *faultyMetricService) RetrieveAll(ctx context.Context) ([]model.Metric, error) {
	return nil, fmt.Errorf("faulty storage get error")
}
