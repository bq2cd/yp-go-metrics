package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	config "github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

const defaultAddress = "localhost:8080"

func runServer(cfg config.Config) error {
	storage := repository.NewMemStorage()
	svc := service.NewMetrics(storage)
	router := handler.NewRouter(svc, nil)

	log.Printf("listening on %v", cfg.ListenAddress)

	return http.ListenAndServe(cfg.ListenAddress, router)
}

func parseArgs(args []string) (config.Config, error) {
	var (
		listenAddress string
	)

	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	fs.StringVar(&listenAddress, "a", defaultAddress, "listen address in the format [HOST]:PORT")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, fmt.Errorf("invalid args: %w", err)
	}

	cfg := config.Config{}

	// validate listen address
	{
		parts := strings.Split(listenAddress, ":")
		if len(parts) > 2 {
			return config.Config{}, fmt.Errorf("invalid listen address")
		}
		cfg.ListenAddress = listenAddress
	}

	return cfg, nil
}

func run(args []string) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	return runServer(cfg)
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
