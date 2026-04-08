package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/auditsink/auditsinktest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func TestMetricAuditor_RecordMetricsUploaded(t *testing.T) {
	uploadedMetrics := model.NewMetricSet(model.NewCounterMetric("id1", 123), model.NewGaugeMetric("id2", -3.4))
	clientInfo := model.ClientInfo{IPAddress: "192.168.1.1"}

	refEvent := model.NewAuditEvent(uploadedMetrics, clientInfo.IPAddress)

	tests := map[string]struct {
		wantErr bool
	}{
		"sink writes event successfully": {
			wantErr: false,
		},
		"sink fails during event writing": {
			wantErr: true,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sink := auditsinktest.NewMockAuditSink(ctrl)

			m := sink.EXPECT().WriteEvent(gomock.Any(), gomock.Cond(func(event model.AuditEvent) bool {
				return gomock.InAnyOrder(event.MetricNames).Matches(refEvent.MetricNames) &&
					event.IPAddress == refEvent.IPAddress &&
					event.Timestamp >= refEvent.Timestamp
			}))

			if tc.wantErr {
				m.Return(errors.New("oops, sink failure"))
			} else {
				m.Return(nil)
			}

			logger := log.NewTestLogger()

			auditor := service.NewMetricAuditor(logger, sink)

			// Act
			auditor.RecordMetricsUploaded(t.Context(), uploadedMetrics, clientInfo)

			// Assert
			events := logger.RecordedEvents()

			if tc.wantErr {
				assert.Len(t, events.FindMatchingEvents(log.LevelError, "cannot write audit event"), 1)
			} else {
				assert.Empty(t, events)
			}
		})
	}
}
