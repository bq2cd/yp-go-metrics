package main

import (
	"log"
	. "os"
)

func aux() (func(), func(), func()) {
	a := func() {
		log.Println("a dummy!")
	}

	b := func() func() {
		return func() {
			Exit(5)
		}
	}

	c := func() func() {
		return func() {
			log.Fatal("hidden exit is fine!")
		}
	}

	return a, b(), c()
}
