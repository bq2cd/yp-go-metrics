package service

import "context"

//go:generate go tool mockgen -destination=servicetest/mock_storage_pinger.go -package servicetest github.com/bq2cd/yp-go-metrics/internal/service StoragePinger

// StoragePinger abstracts a way to check storage health.
type StoragePinger interface {
	Ping(ctx context.Context) error
}
