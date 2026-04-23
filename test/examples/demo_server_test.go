package examples_test

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-metrics/internal/app/server"
	srvconfig "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func initLogger() log.Logger {
	logger, err := log.NewZeroLogger(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		// avoid messing up with example output
		w.Out = os.Stderr
	}))
	if err != nil {
		panic(err)
	}

	return logger

}

func getRandomLocalhostAddress() string {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	addr := l.Addr().String()

	err = l.Close()
	if err != nil {
		panic(err)
	}

	return addr
}

func startDemoServer(addr string) func() error {
	logger := initLogger()

	ctx, cancel := context.WithCancel(context.TODO())

	grp := new(errgroup.Group)
	grp.Go(func() error {
		return server.Run(ctx, logger, srvconfig.Config{
			ListenAddress: addr,
		})
	})

	stop := func() error {
		cancel()

		return grp.Wait()
	}

	client := resty.New().SetBaseURL("http://" + addr)
	ticker := time.NewTicker(10 * time.Millisecond)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			resp, err := client.R().SetContext(ctx).Get("/")
			if err == nil && resp.IsSuccess() {
				break loop
			}
		}
	}

	return stop
}
