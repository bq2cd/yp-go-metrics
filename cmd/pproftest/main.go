// Binary pproftest launches agent and server processes with CPU and memory profiling activated,
// waits for a given time, then performs graceful shutdown of the launched processes.
// Agent process(es) are configured to upload their collected metrics to the server process.
// Profiling results are stored under the path provided as an argument.
package main

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
)

func main() {
	var err error

	t := NewTestingT()

	defer func() {
		teardown(t, err) // panics are handled here
	}()

	err = run(t)
}

func run(t *TestingT) error {
	l, err := NewLauncher(t)
	if err != nil {
		return fmt.Errorf("launcher error: %w", err)
	}

	defer l.Cleanup()

	err = l.Run()
	if err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	return nil
}

func teardown(t *TestingT, errRun error) {
	errFinal := errRun

	result := recover()
	if result != nil {
		errFinal = errors.Join(errFinal, fmt.Errorf("run panic: %v", result))

		log.Printf("RUN PANIC: %v", result)
		log.Println(string(debug.Stack()))

	}

	result = t.RunCleanups()
	if result != nil {
		errFinal = errors.Join(errFinal, fmt.Errorf("cleanup panic: %v", result))

		log.Printf("CLEANUP PANIC: %v", result)
		log.Println(string(debug.Stack()))
	}

	if errFinal != nil {
		log.Fatalf("FINAL ERROR: %v", errFinal)
	}
}
