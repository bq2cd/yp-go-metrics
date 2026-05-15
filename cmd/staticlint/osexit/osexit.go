// Package osexit implements a static code analyzer built on top of `golang.org/x/tools/go/analysis` framework.
// The analyzer reports all calls to `os.Exit` made inside `main` function of the `main` package,
// including the calls inside other functions of the `main` package that are reachable from the `main` function.
// See [Run] method documentation for implementation details.
package osexit

import (
	"errors"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

const (
	inspectedPackage  = "main"
	inspectedFunction = "main"
	forbiddenPackage  = "os"
	forbiddenFunction = "Exit"
)

// Analyzer defines entry point for [analysis] framework.
var Analyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "check for calls of os.Exit in main function of the main package",
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
	Run:      run,
}

var (
	// ErrMissingRequiredAnalyzers is returned in an unlikely case when
	// required analyzers (aka dependencies) have not been run as part of analysis pipeline.
	ErrMissingRequiredAnalyzers = errors.New("missing required analyzers")
)

func run(pass *analysis.Pass) (any, error) {
	c, err := newChecker(pass)
	if err != nil {
		return nil, err
	}

	err = c.Run()

	return nil, err
}
