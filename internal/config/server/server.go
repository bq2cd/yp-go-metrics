package server

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
)

var (
	// ErrInvalidConfig is returned by [Config.Validate] when config is invalid.
	ErrInvalidConfig = errors.New("invalid config")
)

// Config defines a group of options for the server part.
type Config struct {
	ListenAddress            string
	ShutdownTimeout          time.Duration
	MetricStoreInterval      time.Duration
	MetricStoreFilePath      string
	MetricStoreLoadOnStartup bool
	DatabaseURL              url.URL
	HMACSecretKey            []byte
	AuditFilePath            string
	AuditURL                 url.URL
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

		abspath, err := getAbsoluteFilePath(path)
		if err != nil {
			return err
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

// DatabaseURL sets a URL for a database connection.
func DatabaseURL(dsn string) Option {
	return func(c *Config) error {
		if dsn == "" {
			return nil
		}
		uri, err := url.Parse(dsn)
		if err != nil {
			return fmt.Errorf("cannot parse url: %w", err)
		}
		if !(uri.Scheme == "postgres" || uri.Scheme == "postgresql") {
			return fmt.Errorf("scheme must be postgres or postgresql")
		}
		if uri.Hostname() == "" {
			return fmt.Errorf("must specify host")
		}
		if uri.Port() == "" {
			return fmt.Errorf("must specify port")
		}
		c.DatabaseURL = *uri
		return nil
	}
}

// HMACSecretKey sets a secret key to be used HMAC calculation.
func HMACSecretKey(key string) Option {
	return func(c *Config) error {
		c.HMACSecretKey = hmacsigner.LoadSecretKey(key)
		return nil
	}
}

// AuditFilePath sets a path to a file on disk for writing audit events.
func AuditFilePath(path string) Option {
	return func(c *Config) error {
		if path == "" {
			return nil
		}

		abspath, err := getAbsoluteFilePath(path)
		if err != nil {
			return err
		}

		c.AuditFilePath = abspath

		return nil
	}
}

// AuditURL sets a URL for an audit event receiver.
func AuditURL(urlText string) Option {
	return func(c *Config) error {
		if urlText == "" {
			return nil
		}

		uri, err := url.Parse(urlText)
		if err != nil {
			return fmt.Errorf("cannot parse url: %w", err)
		}

		c.AuditURL = *uri

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
	if err := DatabaseURL(c.DatabaseURL.String())(tmp); err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}
	if err := AuditFilePath(c.AuditFilePath)(tmp); err != nil {
		return fmt.Errorf("%w: %w", err, ErrInvalidConfig)
	}

	return nil
}
