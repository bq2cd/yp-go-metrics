package main

import (
	. "os"
)

func main() {
	Exit(3) // want "os.Exit is not allowed in main function"

	c := &cleanuper{}

	c.Add(aux())

	defer c.Run() // want "os.Exit is not allowed in main function"

	aux2() // want "os.Exit is not allowed in main function"
}
