package main

import (
	"os"
)

var globalTricky = os.Exit

func main() {
	code, tricky2 := 7, os.Exit

	os.Exit(1) // want "os.Exit is not allowed in main function"

	aux1() // want "os.Exit is not allowed in main function"

	{
		os.Exit(2) // want "os.Exit is not allowed in main function"
	}

	for i := 100; i < 110; i++ {
		os.Exit(i) // want "os.Exit is not allowed in main function"
	}

	lambda1 := func() {
		os.Exit(11)
	}

	lambda2 := func() (func(int), int) {
		tricky1, code := os.Exit, 12

		return tricky1, code
	}

	defer aux2(lambda1) // want "os.Exit is not allowed in main function"

	defer aux3(lambda2) // want "os.Exit is not allowed in main function"

	go aux4(tricky2) // want "os.Exit is not allowed in main function"

	go func() { // want "os.Exit is not allowed in main function"
		globalTricky(code)
	}()
}
