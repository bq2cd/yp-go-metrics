package logger

import (
	"github.com/bq2cd/yp-go-metrics/internal/log"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
)

// NewProduction returns pre-configured logger for a production environment.
func NewProduction() log.Logger {
	l, err := log.NewZapLogger(zap.NewProductionConfig())
	if err != nil {
		panic(err)
	}
	return l
}

// NewDevelopment returns pre-configured logger for a development environment.
func NewDevelopment() log.Logger {
	l, err := log.NewZeroLogger(zerolog.NewConsoleWriter())
	if err != nil {
		panic(err)
	}
	return l
}
