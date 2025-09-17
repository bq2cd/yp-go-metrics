package agent

import (
	"net/url"
	"time"
)

// Config represents configuration for the agent process.
type Config struct {
	UpstreamURL    url.URL
	PollInterval   time.Duration
	ReportInterval time.Duration
}
