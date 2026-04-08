package service_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository/auditsink/auditsinktest"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type mockSinkConfig struct {
	numSinks       int
	maxCloseErrors int
	maxWriteErrors int
	maxDelay       time.Duration
}

func TestAuditEventProcessor_StartProcessing(t *testing.T) {
	testEvent := model.NewAuditEvent(
		model.NewMetricSet(
			model.NewCounterMetric("id1", 123),
			model.NewGaugeMetric("id2", -1.23),
		),
		"192.168.2.2",
	)

	tests := map[string]struct {
		sinkConfig   mockSinkConfig
		runTimeout   time.Duration
		writeTimeout time.Duration
		warmupDelay  time.Duration
		wantErrFinal error
		wantNoWrites bool
	}{
		"single sink, no delay, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 1,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 10 * time.Millisecond,
		},
		"multiple sinks, no delay, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 10 * time.Millisecond,
		},
		"multiple sinks, some slow, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
				maxDelay: 10 * time.Millisecond,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 10 * time.Millisecond,
		},
		"multiple sinks, some very slow, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
				maxDelay: 50 * time.Millisecond,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 10 * time.Millisecond,
		},
		"multiple sinks, some very slow, some write errors": {
			sinkConfig: mockSinkConfig{
				numSinks:       5,
				maxDelay:       100 * time.Millisecond,
				maxWriteErrors: 3,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 10 * time.Millisecond,
		},
		"multiple sinks, warmup delay too big, closed with no writes": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
				maxDelay: 10 * time.Millisecond,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 10 * time.Millisecond,
			warmupDelay:  30 * time.Millisecond,
			wantErrFinal: service.ErrAuditEventProcessorClosed,
			wantNoWrites: true,
		},
		"multiple sinks, no delay, no errors, write timeout": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 0,
			wantErrFinal: context.DeadlineExceeded,
			wantNoWrites: true,
		},
		"multiple sinks, no delays, all write errors, all close errors": {
			sinkConfig: mockSinkConfig{
				numSinks:       5,
				maxCloseErrors: 5,
				maxWriteErrors: 5,
			},
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 10 * time.Millisecond,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sinks := createMockSinks(ctrl, tc.sinkConfig, testEvent, tc.wantNoWrites)

			logger := log.NewTestLogger()

			ctx, cancel := context.WithTimeout(t.Context(), tc.runTimeout)
			defer cancel()

			processor := service.NewAuditEventProcessor(logger)

			// Act
			doneCh := make(chan struct{}, 1)
			go func() {
				processor.StartProcessing(ctx)
				doneCh <- struct{}{}
			}()

			for i := range sinks {
				processor.RegisterSink(fmt.Sprintf("sink-%d", i), sinks[i])
			}

			time.Sleep(tc.warmupDelay)

			ctxWrite, cancelWrite := context.WithTimeout(t.Context(), tc.writeTimeout)
			defer cancelWrite()

			err := processor.WriteEvent(ctxWrite, testEvent)

			if tc.wantErrFinal != nil {
				require.ErrorIs(t, err, tc.wantErrFinal)
			} else {
				require.NoError(t, err)
			}

			<-doneCh

			if tc.sinkConfig.maxWriteErrors > 0 || tc.sinkConfig.maxCloseErrors > 0 {
				logErrors := make([]log.TestLogEvent, 0)
				for _, logEvent := range logger.RecordedEvents() {
					if logEvent.Level() == log.LevelError {
						logErrors = append(logErrors, logEvent)
					}
				}

				assert.GreaterOrEqual(t, len(logErrors), 1)
			}
		})
	}
}

func createMockSinks(ctrl *gomock.Controller, cfg mockSinkConfig, expectedEvent model.AuditEvent, expectNoWrites bool) []*auditsinktest.MockAuditSink {
	maxDelayMs := int(cfg.maxDelay / time.Millisecond)

	sinks := make([]*auditsinktest.MockAuditSink, cfg.numSinks)

	for i := range cfg.numSinks {
		sink := auditsinktest.NewMockAuditSink(ctrl)
		sinks[i] = sink

		sink.EXPECT().Close().DoAndReturn(
			func() error {
				if rand.IntN(cfg.numSinks) <= cfg.maxCloseErrors {
					return fmt.Errorf("sink close error %d", i)
				}

				return nil
			},
		)

		if expectNoWrites {
			continue
		}

		sink.EXPECT().WriteEvent(gomock.Any(), expectedEvent).DoAndReturn(
			func(_ context.Context, _ model.AuditEvent) error {
				if maxDelayMs > 0 {
					time.Sleep(time.Duration(rand.IntN(maxDelayMs)) * time.Millisecond)
				}

				if rand.IntN(cfg.numSinks) <= cfg.maxWriteErrors {
					return fmt.Errorf("sink write error %d", i)
				}

				return nil
			},
		)
	}

	return sinks
}
