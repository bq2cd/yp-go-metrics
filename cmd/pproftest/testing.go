package main

import (
	"errors"
	"log"
	"sync"
)

var ErrTestingFailNow = errors.New("FAIL NOW")

type TestingT struct {
	mu       sync.Mutex
	cleanups []func()
}

func NewTestingT() *TestingT {
	return &TestingT{
		cleanups: make([]func(), 0),
	}
}

func (t *TestingT) Errorf(format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
}

func (t *TestingT) FailNow() {
	panic(ErrTestingFailNow)
}

func (t *TestingT) Helper() {}

func (t *TestingT) Cleanup(fn func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cleanups = append(t.cleanups, fn)
}

func (t *TestingT) Logf(format string, args ...any) {
	log.Printf("DEBUG: "+format, args...)
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
