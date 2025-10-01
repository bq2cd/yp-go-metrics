package envparser

import (
	"github.com/caarlos0/env/v11"
)

// Parser provides a thin wrapper over `github.com/caarlos0/env` package
// to parse environment variables.
type Parser interface {
	// Parse accepts a pointer to a struct and populates any fields marked with `env` tags.
	// Returns error if parsing of environment variables fails.
	Parse(v any) error
}

type parser struct {
	options *env.Options
}

// NewParser creates an instance of environment parser without custom options.
func NewParser() *parser {
	return &parser{options: nil}
}

// NewParserWithOptions creates an instance of environment parser with custom options.
func NewParserWithOptions(opts env.Options) *parser {
	return &parser{options: &opts}
}

// Parse calls corresponding function from github.com/caarlos0/env package
// depending on the presence of custom options.
// If no custom options are present, `env.Parse(v)` is called, otherwise
// `env.ParseWithOptions(v, opts)` is called.
func (p *parser) Parse(v any) error {
	if p.options == nil {
		return env.Parse(v)
	}
	return env.ParseWithOptions(v, *p.options)
}
