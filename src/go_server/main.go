package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
)

var m int            // Number of bits in the identifier space
var amount_nodes int // Number of nodes in the Chord ring

func main() {

	// Run the test
	// run_test()

	amount_nodes, _ = strconv.Atoi(os.Args[1])
	m, _ = strconv.Atoi(os.Args[2])
	allnodes := initializeChordRing()

	// Convert the nodes to JSON
	jsonData := []byte("[")
	for i, node := range allnodes {
		nodeJSON, _ := json.Marshal(node)
		jsonData = append(jsonData, nodeJSON...)
		if i != len(allnodes)-1 {
			jsonData = append(jsonData, []byte(",\n")...)
		}
	}
	jsonData = append(jsonData, []byte("\n]")...)

	// Create or open the file
	file, err := os.Create("Nodes.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func run_test() {
	// Create a new testing object
	t := new(testing.T)

	// Run the tests
	TestInitFingerTable(t)
}
