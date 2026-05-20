package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bq2cd/yp-go-metrics/internal/app/envparser"
)

// Parser implements common logic of parsing CLI flags, environment variables and JSON config file.
type Parser[Options, Config any] struct {
	DefineArgs        func(*flag.FlagSet, *Options)
	GetConfigFilePath func(*Options) string
	CreateConfig      func(*Options) (Config, error)
}

// Parse parses CLI flags, environment variables and JSON config file in the following order of precedence:
// config file -> CLI flags -> environment variables.
func (p Parser[Options, Config]) Parse(fs *flag.FlagSet, args []string, envParser envparser.Parser) (Config, error) {
	var (
		cfg  Config
		opts Options
		err  error
	)

	p.DefineArgs(fs, &opts)

	err = p.populateOptions(fs, args, envParser, &opts)
	if err != nil {
		return cfg, err
	}

	err = p.loadConfigFile(&opts)
	if err != nil {
		return cfg, err
	}

	// Override options from config file with options from CLI flags and environment variables.
	// Precedence order: config file -> CLI flags -> environment variables
	err = p.populateOptions(fs, args, envParser, &opts)
	if err != nil {
		return cfg, err
	}

	cfg, err = p.CreateConfig(&opts)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (p Parser[Options, Config]) populateOptions(fs *flag.FlagSet, args []string, envParser envparser.Parser, opts *Options) error {
	// parse flags
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("invalid args: %w", err)
	}

	// parse env vars (take precedence over flags)
	if err := envParser.Parse(opts); err != nil {
		return fmt.Errorf("invalid env vars: %w", err)
	}

	return nil
}

func (p Parser[Options, Config]) loadConfigFile(opts *Options) error {
	path := p.GetConfigFilePath(opts)
	if path == "" {
		return nil
	}

	cfg, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open config file at %s: %w", path, err)
	}
	defer cfg.Close()

	err = json.NewDecoder(cfg).Decode(opts)
	if err != nil {
		return fmt.Errorf("cannot JSON-decode config file: %w", err)
	}

	return nil
}
