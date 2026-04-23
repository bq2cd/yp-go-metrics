package main

import (
	"errors"
	"log"
	"sync"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
)

// ErrTestingFailNow is used in `panic` in [TestingT.FailNow] to allow distinguishing
// between different kinds of panic.
var ErrTestingFailNow = errors.New("FAIL NOW")

var _ servertest.TestingT = (*TestingT)(nil)

// TestingT is a dummy implementation of [servertest.TestingT] interface to
// make it possible to use [servertest] module outside of tests.
type TestingT struct {
	mu       sync.Mutex
	cleanups []func()
}

// NewTestingT creates an instance of [TestingT].
func NewTestingT() *TestingT {
	return &TestingT{
		cleanups: make([]func(), 0),
	}
}

// Errorf logs an error using [log.Printf].
func (t *TestingT) Errorf(format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
}

// FailNow panics with [ErrTestingFailNow] error to abort the execution flow.
// It simulates [testing.T.FailNow] method.
func (t *TestingT) FailNow() {
	panic(ErrTestingFailNow)
}

// Helper is a noop function required for implementation of [servertest.TestingT] interface.
func (t *TestingT) Helper() {
	// [testing.T.Helper] performs a complex logic of recording the caller's stack,
	// but we do not need this outside of tests.
}

// Cleanup registers a callback function that will be called in the end of the execution
// (typically at the end of [main] function).
// All registered functions are executed in reverse order, starting from the last registered function.
func (t *TestingT) Cleanup(fn func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cleanups = append(t.cleanups, fn)
}

// Logf logs a message using [log.Printf].
func (t *TestingT) Logf(format string, args ...any) {
	log.Printf("DEBUG: "+format, args...)
}

// RunCleanups performs execution of all registered cleanup functions.
// It attempts to run every cleanup function even when some cleanup functions panic.
// If one or more panics have occurred, the value of the last panic will be returned, and
// the resulting stacktrace will contain all panic calls.
func (t *TestingT) RunCleanups() any {
	return t.runCleanups(true)
}

func (t *TestingT) runCleanups(catchPanic bool) (panicVal any) {
	// shamelessly "inspired" by [testing.runCleanup] :)

	if catchPanic {
		defer func() {
			panicVal = recover()
		}()
	}

	defer func() {
		t.mu.Lock()
		more := len(t.cleanups) > 0
		t.mu.Unlock()

		if more {
			t.runCleanups(false) // let first invocation catch panics
		}
	}()

	for {
		var cleanup func()

		t.mu.Lock()
		if len(t.cleanups) > 0 {
			last := len(t.cleanups) - 1
			cleanup = t.cleanups[last]
			t.cleanups = t.cleanups[:last]
		}
		t.mu.Unlock()

		if cleanup == nil {
			return
		}

		cleanup()
	}
}
