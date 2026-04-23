package testutil

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// SkipTestInGithubActions skips current test if GITHUB_ACTIONS env var is present and not empty.
func SkipTestInGithubActions(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("not supported in Github Actions")
	}
}

// AssertNumberInBetween verifies that provided number satisfies `wantLow <= actual <= wantHigh`.
func AssertNumberInBetween(t *testing.T, wantLow, wantHigh, actual int) {
	assert.GreaterOrEqual(t, actual, wantLow)
	assert.LessOrEqual(t, actual, wantHigh)
}

// Mock is wrapper on top of [mock.Mock].
type Mock struct {
	mock.Mock

	mu sync.RWMutex
}

// Called is a wrapper on top of [mock.Called] to add additional
// layer of synchronization for safe access to [mock.Calls] slice.
func (m *Mock) Called(arguments ...any) mock.Arguments {
	// simplified code copied from [mock.Colled] to get caller information.
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("Couldn't get the caller information")
	}
	functionPath := runtime.FuncForPC(pc).Name()
	parts := strings.Split(functionPath, ".")
	functionName := parts[len(parts)-1]

	// We need this mutex to protect [mock.Calls] slice.
	// While it is protected internally for writes with [mock.mutex],
	// there is no public API to get number of calls per given method
	// without running into a race condition.
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.MethodCalled(functionName, arguments...)
}

// GetNumCalls provides a safe access to underlying [mock.Calls] slice.
// This is a missing piece from the official testify API, unfortunately.
func (m *Mock) GetNumCalls(methodName string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var actualCalls int
	for _, call := range m.Calls {
		if call.Method == methodName {
			actualCalls++
		}
	}

	return actualCalls
}
