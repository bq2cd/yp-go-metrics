package handler

import (
	"fmt"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

type faultyMetricService struct{}

func (s *faultyMetricService) StoreSingle(metric model.Metric) error {
	return fmt.Errorf("faulty storage set error")
}

func (s *faultyMetricService) StoreBatch(metrics []model.Metric) error {
	return fmt.Errorf("faulty storage set error")
}

func (s *faultyMetricService) RetrieveSingle(key model.MetricKey) (model.Metric, error) {
	return model.Metric{}, fmt.Errorf("faulty storage get error")
}

func (s *faultyMetricService) RetrieveBatch(keys []model.MetricKey) ([]model.Metric, error) {
	return nil, fmt.Errorf("faulty storage get error")
}

func (s *faultyMetricService) RetrieveAll() ([]model.Metric, error) {
	return nil, fmt.Errorf("faulty storage get error")
}
