package cli

import "io"

// TerminalConfig defines terminal configuration, in particular stdout and stderr streams.
type TerminalConfig struct {
	Stdout io.Writer
	Stderr io.Writer
}
