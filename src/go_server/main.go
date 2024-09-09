package main

import (
	"testing"
)

const m = 3            // Number of bits in the identifier space
const amount_nodes = 4 // Number of nodes in the Chord ring

func main() {

	// Create a new testing object
	t := new(testing.T)

	// Run the tests
	TestInitFingerTable(t)

}
