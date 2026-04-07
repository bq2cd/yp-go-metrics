package testutil

import (
	"os"
	"testing"
)

// SkipTestInGithubActions skips current test if GITHUB_ACTIONS env var is present and not empty.
func SkipTestInGithubActions(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("not supported in Github Actions")
	}
}
