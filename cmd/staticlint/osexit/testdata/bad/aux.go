package main

import (
	"os"
)

func auxNeverCalled() {
	os.Exit(99)
}

func aux1() {
	if false {
		os.Exit(21)
	}
}

func aux2(fn func()) {
	fn()
}

func aux3(fn func() (func(int), int)) {
	tricky, code := fn()

	tricky(code)
}

func aux4(fn func(int)) {
	fn(24)
}
