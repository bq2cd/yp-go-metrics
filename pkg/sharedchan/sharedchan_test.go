package sharedchan_test

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/pkg/sharedchan"
)

func TestSendAndClose(t *testing.T) {
	testTimeout := 5 * time.Second

	const (
		defaultReceivers = 1
		defaultClosers   = 3
	)
	type workers struct {
		senders   int
		receivers int
		closers   int
	}
	type delays struct {
		close   time.Duration
		receive time.Duration
	}
	type want struct {
		sent     int
		received int
	}
	tests := map[string]struct {
		chanSize uint
		workers  workers
		delays   delays
		want     want
	}{
		"unbuffered chan, many workers, slow receiver": {
			chanSize: 0,
			workers:  workers{senders: 10, receivers: defaultReceivers, closers: defaultClosers},
			delays:   delays{close: 10 * time.Millisecond, receive: 50 * time.Millisecond},
			want:     want{sent: 0, received: 0},
		},
		"unbuffered chan, many workers, delayed receiver": {
			chanSize: 0,
			workers:  workers{senders: 10, receivers: defaultReceivers, closers: defaultClosers},
			delays:   delays{close: 15 * time.Millisecond, receive: 10 * time.Millisecond},
			want:     want{sent: 1, received: 1},
		},
		"unbuffered chan, many workers, multiple delayed receivers": {
			chanSize: 0,
			workers:  workers{senders: 10, receivers: 5, closers: defaultClosers},
			delays:   delays{close: 15 * time.Millisecond, receive: 10 * time.Millisecond},
			want:     want{sent: 5, received: 5},
		},
		"chan capacity is less than number of workers, slow receiver": {
			chanSize: 5,
			workers:  workers{senders: 15, receivers: defaultReceivers, closers: defaultClosers},
			delays:   delays{close: 10 * time.Millisecond, receive: 50 * time.Millisecond},
			want:     want{sent: 5, received: 5},
		},
		"chan capacity is less than number of workers, delayed receiver": {
			chanSize: 5,
			workers:  workers{senders: 15, receivers: defaultReceivers, closers: defaultClosers},
			delays:   delays{close: 15 * time.Millisecond, receive: 10 * time.Millisecond},
			want:     want{sent: 6, received: 6},
		},
		"chan capacity is less than number of workers, fast receiver": {
			chanSize: 5,
			workers:  workers{senders: 15, receivers: defaultReceivers, closers: defaultClosers},
			delays:   delays{close: 15 * time.Millisecond, receive: 0 * time.Millisecond},
			want:     want{sent: 15, received: 15},
		},
		"chan capacity is less than number of workers, multiple delayed receivers": {
			chanSize: 5,
			workers:  workers{senders: 15, receivers: 5, closers: defaultClosers},
			delays:   delays{close: 15 * time.Millisecond, receive: 10 * time.Millisecond},
			want:     want{sent: 10, received: 10},
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			var numSent, numReceived atomic.Int64

			ch := sharedchan.NewChannel[int](tc.chanSize)

			grp := new(errgroup.Group)

			// senders
			for range tc.workers.senders {
				grp.Go(func() error {
					v := rand.IntN(100)
					ok := ch.Send(v)
					t.Logf("sent: %d (%v)", v, ok)
					if ok {
						numSent.Add(1)
					}

					return nil
				})
			}

			// receivers
			for range tc.workers.receivers {
				grp.Go(func() error {
					for {
						<-time.After(tc.delays.receive)
						v, ok := <-ch.Receive()
						if !ok {
							break
						}
						t.Logf("received: %d", v)
						numReceived.Add(1)
					}

					return nil
				})
			}

			// closers (simulating concurrent closing)
			for range tc.workers.closers {
				grp.Go(func() error {
					time.Sleep(tc.delays.close)
					ch.Close()

					return nil
				})
			}

			done := make(chan struct{})

			go func() {
				grp.Wait()
				done <- struct{}{}
			}()

			// ensure we catch any deadlocks
			timer := time.NewTimer(testTimeout)
			select {
			case <-done:
				timer.Stop()
			case <-timer.C:
				t.Fatalf("test timeout (%v) exceeded!", testTimeout)
			}

			// assert
			assert.Equal(t, tc.want.sent, int(numSent.Load()), "num sent")
			assert.Equal(t, tc.want.received, int(numReceived.Load()), "num received")
		})
	}
}

func BenchmarkSend(b *testing.B) {
	type config struct {
		numEvents int
		chanSize  int
		senders   int
	}

	for _, cfg := range []config{
		{numEvents: 10, chanSize: 0, senders: 1},
		{numEvents: 10, chanSize: 0, senders: 10},
		{numEvents: 100, chanSize: 0, senders: 1},
		{numEvents: 100, chanSize: 0, senders: 10},
		{numEvents: 100, chanSize: 10, senders: 1},
		{numEvents: 100, chanSize: 10, senders: 10},
		{numEvents: 100, chanSize: 100, senders: 1},
		{numEvents: 100, chanSize: 100, senders: 10},
		{numEvents: 1000, chanSize: 100, senders: 1},
		{numEvents: 1000, chanSize: 100, senders: 10},
		{numEvents: 1000, chanSize: 100, senders: 25},
		{numEvents: 10_000, chanSize: 100, senders: 1},
		{numEvents: 10_000, chanSize: 100, senders: 10},
		{numEvents: 10_000, chanSize: 100, senders: 25},
	} {
		suffix := fmt.Sprintf(" size=%d events=%d senders=%d", cfg.chanSize, cfg.numEvents, cfg.senders)

		events := generateEvents(cfg.numEvents, func() int { return rand.IntN(cfg.numEvents) })
		bcfg := benchmarkSendConfig{chanSize: cfg.chanSize, senders: cfg.senders}

		b.Run("wrapped channel"+suffix, func(b *testing.B) {
			benchmarkSendWrappedChannel(b, bcfg, events)
		})

		b.Run("plain channel"+suffix, func(b *testing.B) {
			benchmarkSendPlainChannel(b, bcfg, events)
		})
	}
}

type benchmarkSendConfig struct {
	chanSize int
	senders  int
}

func generateEvents[T any](numEvents int, generateFn func() T) []T {
	events := make([]T, numEvents)

	for i := range events {
		events[i] = generateFn()
	}

	return events
}

func benchmarkSendWrappedChannel[T any](b *testing.B, cfg benchmarkSendConfig, events []T) {
	ch := sharedchan.NewChannel[T](uint(cfg.chanSize))
	defer ch.Close()

	go drainChannel(ch.Receive())

	sendToChannelInParallel(b, cfg, events, func(v T) { ch.Send(v) })
}

func benchmarkSendPlainChannel[T any](b *testing.B, cfg benchmarkSendConfig, events []T) {
	ch := make(chan T, cfg.chanSize)
	defer close(ch)

	go drainChannel(ch)

	sendToChannelInParallel(b, cfg, events, func(v T) { ch <- v })
}

func sendToChannelInParallel[T any](b *testing.B, cfg benchmarkSendConfig, events []T, sendFn func(T)) {
	wg := new(sync.WaitGroup)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		wg.Add(cfg.senders)

		for range cfg.senders {
			go func() {
				defer wg.Done()
				for i := range events {
					sendFn(events[i])
				}
			}()
		}

		wg.Wait()
	}
}

func drainChannel[T any](ch <-chan T) {
	for range ch {
		// drain channel
	}
}
