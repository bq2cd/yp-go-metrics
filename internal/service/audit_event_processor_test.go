package service_test

import (
	"context"
	"errors"
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
	"github.com/bq2cd/yp-go-metrics/pkg/option"
)

type mockSinkConfig struct {
	numSinks       int
	maxCloseErrors int
	maxWriteErrors int
	minDelay       time.Duration
	maxDelay       time.Duration
}

func TestAuditEventProcessor_StartProcessing(t *testing.T) {
	testTimeout := 5 * time.Second

	tests := map[string]struct {
		sinkConfig    mockSinkConfig
		testEvents    []model.AuditEvent
		runTimeout    time.Duration
		writeTimeout  time.Duration
		warmupDelay   time.Duration
		opts          []option.Option[service.AuditEventProcessorConfig]
		wantErrFinal  error
		wantNumWrites int
	}{
		"single sink, no delay, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 1,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  10 * time.Millisecond,
			wantNumWrites: 1,
		},
		"multiple sinks, no delay, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  10 * time.Millisecond,
			wantNumWrites: 1,
		},
		"multiple sinks, some slow, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
				maxDelay: 10 * time.Millisecond,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  10 * time.Millisecond,
			wantNumWrites: 1,
		},
		"multiple sinks, some very slow, no errors": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
				maxDelay: 50 * time.Millisecond,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  10 * time.Millisecond,
			wantNumWrites: 1,
		},
		"multiple sinks, some very slow, some write errors": {
			sinkConfig: mockSinkConfig{
				numSinks:       5,
				maxDelay:       100 * time.Millisecond,
				maxWriteErrors: 3,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  10 * time.Millisecond,
			wantNumWrites: 1,
		},
		"multiple sinks, warmup delay too big, closed with no writes": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
				maxDelay: 10 * time.Millisecond,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  10 * time.Millisecond,
			warmupDelay:   30 * time.Millisecond,
			wantErrFinal:  service.ErrAuditEventProcessorClosed,
			wantNumWrites: 0,
		},
		"multiple sinks, no delay, no errors, write timeout": {
			sinkConfig: mockSinkConfig{
				numSinks: 5,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  0,
			wantErrFinal:  context.DeadlineExceeded,
			wantNumWrites: 0,
		},
		"multiple sinks, no delays, all write errors, all close errors": {
			sinkConfig: mockSinkConfig{
				numSinks:       5,
				maxCloseErrors: 5,
				maxWriteErrors: 5,
			},
			testEvents:    generateAuditEvents(1),
			runTimeout:    20 * time.Millisecond,
			writeTimeout:  10 * time.Millisecond,
			wantNumWrites: 1,
		},
		"number of incoming events is greater than channel buffer size, concurrency less than number of sinks, slow sinks": {
			sinkConfig: mockSinkConfig{
				numSinks: 4,
				minDelay: 15 * time.Millisecond,
				maxDelay: 20 * time.Millisecond,
			},
			testEvents:   generateAuditEvents(10),
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 100 * time.Millisecond,
			opts: []option.Option[service.AuditEventProcessorConfig]{
				service.WithAuditEventProcessorBufferSize(2),
				service.WithAuditEventProcessorConcurrency(2),
			},
			wantNumWrites: 1,
			wantErrFinal:  service.ErrAuditEventProcessorClosed,
		},
		"number of incoming events is greater than channel buffer size, concurrency less than number of sinks, delayed sinks": {
			sinkConfig: mockSinkConfig{
				numSinks: 4,
				minDelay: 5 * time.Millisecond,
				maxDelay: 5 * time.Millisecond,
			},
			testEvents:   generateAuditEvents(10),
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 100 * time.Millisecond,
			opts: []option.Option[service.AuditEventProcessorConfig]{
				service.WithAuditEventProcessorBufferSize(2),
				service.WithAuditEventProcessorConcurrency(2),
			},
			wantNumWrites: 2,
			wantErrFinal:  service.ErrAuditEventProcessorClosed,
		},
		"number of incoming events is greater than channel buffer size, unlimited concurrency, slow sinks": {
			sinkConfig: mockSinkConfig{
				numSinks: 4,
				minDelay: 15 * time.Millisecond,
				maxDelay: 20 * time.Millisecond,
			},
			testEvents:   generateAuditEvents(10),
			runTimeout:   20 * time.Millisecond,
			writeTimeout: 100 * time.Millisecond,
			opts: []option.Option[service.AuditEventProcessorConfig]{
				service.WithAuditEventProcessorBufferSize(2),
				service.WithAuditEventProcessorConcurrency(0),
			},
			wantNumWrites: 2,
			wantErrFinal:  service.ErrAuditEventProcessorClosed,
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sinks := createMockSinks(ctrl, tc.sinkConfig, tc.testEvents[:tc.wantNumWrites])

			logger := log.NewTestLogger()

			processor := service.NewAuditEventProcessor(logger, tc.opts...)

			ctxRun, cancelRun := context.WithTimeout(t.Context(), tc.runTimeout)
			defer cancelRun()

			// Act
			doneCh := make(chan struct{}, 1)
			go func() {
				processor.StartProcessing(ctxRun)
				doneCh <- struct{}{}
			}()

			errCh := make(chan error, 1)
			go func() {
				for i := range sinks {
					processor.RegisterSink(fmt.Sprintf("sink-%d", i), sinks[i])
				}

				time.Sleep(tc.warmupDelay)

				ctxWrite, cancelWrite := context.WithTimeout(t.Context(), tc.writeTimeout)
				defer cancelWrite()

				var errFinal error
				for _, event := range tc.testEvents {
					err := processor.WriteEvent(ctxWrite, event)
					errFinal = errors.Join(errFinal, err)
				}

				errCh <- errFinal
			}()

			// ensure we are never deadlocked
			timer := time.NewTimer(testTimeout)
			select {
			case <-doneCh:
				timer.Stop()
			case <-timer.C:
				t.Fatalf("test timeout (%v) exceeded!", testTimeout)
			}

			// Assert
			errFinal := <-errCh
			if tc.wantErrFinal != nil {
				require.ErrorIs(t, errFinal, tc.wantErrFinal)
			} else {
				require.NoError(t, errFinal)
			}

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

func createMockSinks(ctrl *gomock.Controller, cfg mockSinkConfig, expectedEvents []model.AuditEvent) []*auditsinktest.MockAuditSink {
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

		for _, event := range expectedEvents {
			sink.EXPECT().WriteEvent(gomock.Any(), event).DoAndReturn(
				func(_ context.Context, _ model.AuditEvent) error {
					delay := cfg.minDelay
					if maxDelayMs > 0 {
						delay = max(delay, time.Duration(rand.IntN(maxDelayMs))*time.Millisecond)
					}

					time.Sleep(delay)

					if rand.IntN(cfg.numSinks) <= cfg.maxWriteErrors {
						return fmt.Errorf("sink write error %d", i)
					}

					return nil
				},
			)
		}
	}

	return sinks
}

func generateAuditEvents(n int) []model.AuditEvent {
	events := make([]model.AuditEvent, n)

	for i := 1; i <= len(events); i++ {
		events[i-1] = model.NewAuditEvent(
			model.NewMetricSet(
				model.NewCounterMetric(fmt.Sprintf("counter-%d", i), int64(i)*100),
				model.NewGaugeMetric(fmt.Sprintf("gauge-%d", i), -1*float64(i)/100),
			),
			fmt.Sprintf("192.168.0.%d", i),
		)
	}

	return events
}
