package db

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	DriverNone = Driver("")
	DriverPgx  = Driver("pgx")
)

var (
	ErrUnsupportedDBType = errors.New("unsupported database type")

	supportedDBTypes = map[string]bool{
		"postgres":   true,
		"postgresql": true,
	}
)

type Driver string

// Config represents database connection details.
// Fields are not exported to prevent mutation from outside.
type Config struct {
	driver Driver
	dsn    string
}

// Enabled returns true if database url is not empty.
func (c Config) Enabled() bool {
	return c.dsn != ""
}

// Driver returns database driver.
func (c Config) Driver() Driver {
	return c.driver
}

// DSN returns database connection string.
func (c Config) DSN() string {
	return c.dsn
}

// New creates an instance of [Config] from the given URL.
func New(dbURL url.URL) (Config, error) {
	c := Config{
		driver: DriverPgx,
		dsn:    dbURL.String(),
	}
	if c.dsn == "" {
		// empty DSN is allowed - this means database is not enabled
		return c, nil
	}
	if !supportedDBTypes[dbURL.Scheme] {
		return Config{}, fmt.Errorf("%w: %s", ErrUnsupportedDBType, dbURL.Scheme)
	}
	return c, nil
}
