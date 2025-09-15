package main

import (
	"log"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/config/server"
	"github.com/bq2cd/yp-go-metrics/internal/handler"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

const defaultAddress = ":8080"

func run(config server.Config) error {
	storage := repository.NewMemStorage()
	svc := service.NewMetrics(storage)
	router := handler.NewRouter(svc, nil)

	log.Printf("listening on %v", config.ListenAddress)

	return http.ListenAndServe(config.ListenAddress, router)
}

func main() {
	config := server.Config{ListenAddress: defaultAddress}

	err := run(config)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
