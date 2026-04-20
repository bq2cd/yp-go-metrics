package main

import (
	"log"
	"os"
)

func main() {
	os.Environ()

	log.Fatalln("oops")
}
