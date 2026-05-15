// Binary staticlint implements custom static code analyzer using `multichecker` framework.
//
// Included analyzers:
//
// - standard analyzers from `golang.org/x/tools/go/analysis/passes` (for concrete list see [passes] subpackage).
//
// - staticcheck analyzers of `SA` category + the following analyzers from other categories:
//   - S1000 Use plain channel send or receive instead of single-case select.
//   - S1030 Use [bytes.Buffer.String] or [bytes.Buffer.Bytes].
//   - ST1005 Incorrectly formatted error string.
//   - ST1015 A switch's default case should be the first or last case.
//   - U1000 Unused code.
//
// - `durationcheck` analyzer (https://github.com/charithe/durationcheck);
//
// - `embeddedstructfieldcheck` analyzer (https://github.com/manuelarte/embeddedstructfieldcheck);
//
// - custom analyzer prohibiting usage of `os.Exit` in `main` function of `main` package;
package main

import (
	"strings"

	"github.com/charithe/durationcheck"
	embeddedstructfield "github.com/manuelarte/embeddedstructfieldcheck/analyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"

	"github.com/bq2cd/yp-go-metrics/cmd/staticlint/osexit"
	"github.com/bq2cd/yp-go-metrics/cmd/staticlint/passes"
)

func main() {
	multichecker.Main(
		buildAnalyzers()...,
	)
}

func buildAnalyzers() []*analysis.Analyzer {
	analyzers := make([]*analysis.Analyzer, 0)

	// analysis/passes
	analyzers = append(analyzers, passes.Analyzers[:]...)

	// staticcheck
	analyzers = append(analyzers, getStaticcheckAnalyzers()...)

	// extra public analyzers
	analyzers = append(analyzers, durationcheck.Analyzer)
	analyzers = append(analyzers, embeddedstructfield.NewAnalyzer())

	// custom analyzer (`os.Exit` detection)
	analyzers = append(analyzers, osexit.Analyzer)

	return analyzers
}

func getStaticcheckAnalyzers() []*analysis.Analyzer {
	analyzers := make([]*analysis.Analyzer, 0)

	extra := map[string]bool{
		"S1000":  true, // Use plain channel send or receive instead of single-case select
		"S1030":  true, // Use bytes.Buffer.String or bytes.Buffer.Bytes
		"ST1005": true, // Incorrectly formatted error string
		"ST1015": true, // A switch's default case should be the first or last case
		"U1000":  true, // Unused code
	}

	for _, check := range staticcheck.Analyzers {
		if strings.HasPrefix(check.Analyzer.Name, "SA") {
			analyzers = append(analyzers, check.Analyzer)
		}

		if extra[check.Analyzer.Name] {
			analyzers = append(analyzers, check.Analyzer)
		}
	}

	return analyzers
}
