package main

import (
	"fmt"
	custom "os"
)

var globalTricky = custom.Exit

func main() {
	custom.Exit(1) // want "os.Exit is not allowed in main function"

	if false {
		custom.Exit(2) // want "os.Exit is not allowed in main function"
	}

	{
		custom.Exit(3) // want "os.Exit is not allowed in main function"
	}

	for i := range 10 {
		custom.Exit(i + 100) // want "os.Exit is not allowed in main function"
	}

	lambda := func() (func(int), error) {
		tricky, err := custom.Exit, fmt.Errorf("never called")

		return tricky, err
	}

	defer func() { // want "os.Exit is not allowed in main function"
		tricky, _ := lambda()

		tricky(3)
	}()

	go func() { // want "os.Exit is not allowed in main function"
		globalTricky(7)
	}()

	tricky, code := custom.Exit, 5

	tricky(code) // want "os.Exit is not allowed in main function"
}

func runNeverCalled() {
	custom.Exit(1)
}
