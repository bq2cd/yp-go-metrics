package agent

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
)

var (
	reURLCheck = regexp.MustCompile("^[^:/]+://.+")

	ErrInvalidConfig = errors.New("invalid config")
)

// Config represents configuration for the agent process.
type Config struct {
	UpstreamURL    url.URL
	PollInterval   time.Duration
	ReportInterval time.Duration
	HMACSecretKey  []byte
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

// UpstreamURL sets value for upstream url for the agent to report
// metrics to.
func UpstreamURL(upstreamURL string) Option {
	return func(c *Config) error {
		if upstreamURL == "" {
			return fmt.Errorf("empty upstream url")
		}
		if !reURLCheck.MatchString(upstreamURL) {
			upstreamURL = "http://" + upstreamURL
		}

		parsed, err := url.Parse(upstreamURL)
		if err != nil {
			return fmt.Errorf("invalid upstream url (%v)", upstreamURL)
		}
		c.UpstreamURL = *parsed
		return nil
	}
}

func setInterval(intervalSec uint, target *time.Duration) error {
	if intervalSec == 0 {
		return fmt.Errorf("interval must be > 0 (got %d)", intervalSec)
	}
	*target = time.Duration(intervalSec) * time.Second
	return nil
}

// PollInterval sets interval for metrics polling process.
func PollInterval(intervalSec uint) Option {
	return func(c *Config) error {
		return setInterval(intervalSec, &c.PollInterval)
	}
}

// ReportInterval sets interval for metrics reporting process.
func ReportInterval(intervalSec uint) Option {
	return func(c *Config) error {
		return setInterval(intervalSec, &c.ReportInterval)
	}
}

// HMACSecretKey sets a secret key to be used HMAC calculation.
func HMACSecretKey(key string) Option {
	return func(c *Config) error {
		c.HMACSecretKey = hmacsigner.LoadSecretKey(key)
		return nil
	}
}

// Validate performs logical validation of the config and returns
// an error if some values do not make sense (logically).
func (c *Config) Validate() error {
	if c.UpstreamURL.String() == "" {
		return fmt.Errorf("missing upstream url: %w", ErrInvalidConfig)
	}
	if c.PollInterval == 0 {
		return fmt.Errorf("poll interval must be > 0: %w", ErrInvalidConfig)
	}
	if c.ReportInterval == 0 {
		return fmt.Errorf("report interval must be > 0: %w", ErrInvalidConfig)
	}
	if c.ReportInterval < c.PollInterval {
		return fmt.Errorf("report interval must be >= poll interval (got %v < %v): %w", c.ReportInterval, c.PollInterval, ErrInvalidConfig)
	}
	return nil
}
