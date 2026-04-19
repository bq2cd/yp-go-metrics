// Binary pproftest launches agent and server processes with CPU and memory profiling activated,
// waits for a given time, then performs graceful shutdown of the launched processes.
// Agent process(es) are configured to upload their collected metrics to the server process.
// Profiling results are stored under the path provided as an argument.
package main

import (
	"log"
	"os"
	"runtime/debug"
)

var exitCode int

func main() {
	t := NewTestingT()
	defer teardown(t)

	l := NewLauncher(t)
	defer l.Cleanup()

	err := l.Run()
	if err != nil {
		log.Printf("RUN ERROR: %v", err)

		exitCode = 1
	}
}

func teardown(t *TestingT) {
	err := recover()
	if err != nil {
		log.Printf("PANIC: %v", err)
		log.Println(string(debug.Stack()))

		exitCode = 2
	}

	err = t.RunCleanups()
	if err != nil {
		log.Printf("CLEANUP PANIC: %v", err)

		exitCode = 3
	}

	os.Exit(exitCode)
}
