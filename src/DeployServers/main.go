package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

var m int                 // Number of bits in the identifier space
var amount_nodes int      // Number of nodes in the Chord ring
var address_list []string // List of addresses for the nodes

func main() {

	m, _ = strconv.Atoi(os.Args[1])
	address_list_string := os.Args[2]
	amount_nodes, _ = strconv.Atoi(os.Args[3])

	// Split the address list into individual addresses
	err := json.Unmarshal([]byte(address_list_string), &address_list)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Initialize the Chord ring
	allnodes := initializeChordRing()

	// Marshal the allnodes array into JSON
	jsonData, err := json.Marshal(allnodes)

	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

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
