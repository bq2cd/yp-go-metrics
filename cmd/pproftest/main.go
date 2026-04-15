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

	err = t.runCleanups(true)
	if err != nil {
		log.Printf("CLEANUP PANIC: %v", err)

		exitCode = 3
	}

	os.Exit(exitCode)
}
