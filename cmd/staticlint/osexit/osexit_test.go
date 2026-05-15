package osexit_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/bq2cd/yp-go-metrics/cmd/staticlint/osexit"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), osexit.Analyzer, "./...")
}
