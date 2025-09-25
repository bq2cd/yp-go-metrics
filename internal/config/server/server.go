package server

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
)

// Config defines a group of options for the server part.
type Config struct {
	ListenAddress   string
	ShutdownTimeout time.Duration
}

// Option is function that take pointer to config as an argument,
// modifies config accordingly, and returns error if provided value
// (via a parent function) is invalid.
type Option func(*Config) error

// New create new Config and applies options to it (if any).
// If any option fails, this function return nil instead of config
// and the corresponding error.
func New(opts ...Option) (*Config, error) {
	c := &Config{}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// ListenAddress sets value for the server to listen at.
func ListenAddress(addr string) Option {
	return func(c *Config) error {
		if addr == "" {
			return fmt.Errorf("empty listen addr")
		}
		if len(strings.Split(addr, ":")) > 2 {
			return fmt.Errorf("invalid listen addr (got %v)", addr)
		}
		c.ListenAddress = addr
		return nil
	}
}

// ShutdownTimeout sets timeout (in seconds) for graceful shutdown of the server.
func ShutdownTimeout(timeoutSec uint) Option {
	return func(c *Config) error {
		if timeoutSec == 0 {
			return fmt.Errorf("timeout must > 0 (got %d)", timeoutSec)
		}
		c.ShutdownTimeout = time.Duration(timeoutSec) * time.Second
		return nil
	}
}

// Validate performs logical validation of the config and returns
// an error if some values do not make sense (logically).
func (c *Config) Validate() error {
	if c.ListenAddress == "" {
		return fmt.Errorf("missing listen address: %w", ErrInvalidConfig)
	}
	if c.ShutdownTimeout == 0 {
		return fmt.Errorf("shutdown timeout must be > 0: %w", ErrInvalidConfig)
	}
	return nil
}
