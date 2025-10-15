package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
)

// Config defines a group of options for the server part.
type Config struct {
	ListenAddress            string
	ShutdownTimeout          time.Duration
	MetricStoreInterval      time.Duration
	MetricStoreFilePath      string
	MetricStoreLoadOnStartup bool
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

// MetricStoreInterval sets an interval (in seconds) for the server to dump in-memory metrics to disk.
func MetricStoreInterval(intervalSec uint) Option {
	return func(c *Config) error {
		c.MetricStoreInterval = time.Duration(intervalSec) * time.Second
		return nil
	}
}

// MetricStoreFilePath sets a path to a file on disk where the server would dump in-memory metrics.
func MetricStoreFilePath(path string) Option {
	return func(c *Config) error {
		if path == "" {
			c.MetricStoreFilePath = path
			return nil
		}
		stat, err := os.Stat(path)
		if err == nil && stat.IsDir() {
			return fmt.Errorf("dir is not allowed")
		}
		abspath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("cannot determine absolute path: %w", err)
		}
		c.MetricStoreFilePath = abspath
		return nil
	}
}

// MetricStoreLoadOnStartup instructs the server to load metrics from a file on startup.
func MetricStoreLoadOnStartup(action bool) Option {
	return func(c *Config) error {
		c.MetricStoreLoadOnStartup = action
		return nil
	}
}

// Validate performs logical validation of the config and returns
// an error if some values do not make sense (logically).
func (c *Config) Validate() error {
	tmp := &Config{}
	if err := ListenAddress(c.ListenAddress)(tmp); err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}
	if err := ShutdownTimeout(uint(c.ShutdownTimeout / time.Second))(tmp); err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}
	if err := MetricStoreInterval(uint(c.MetricStoreInterval / time.Second))(tmp); err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}
	if err := MetricStoreFilePath(c.MetricStoreFilePath)(tmp); err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}
	if err := MetricStoreLoadOnStartup(c.MetricStoreLoadOnStartup)(tmp); err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}
	if c.MetricStoreLoadOnStartup && c.MetricStoreFilePath == "" {
		return fmt.Errorf("must specify metric store path when loading on startup: %w", ErrInvalidConfig)
	}
	return nil
}
