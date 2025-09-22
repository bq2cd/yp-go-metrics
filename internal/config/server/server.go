package server

import "time"

// Config defines a group of options for the server part.
type Config struct {
	ListenAddress   string
	ShutdownTimeout time.Duration
}
