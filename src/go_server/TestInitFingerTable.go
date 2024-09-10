package main

import (
	"fmt"
	"math"
	"testing"
)

func TestInitFingerTable(t *testing.T) {

	// Define the number of bits in the identifier space
	const bit_space = 3

	// Create a list of nodes
	allNodes := []*Node{
		{1, nil, nil, 0},
		{3, nil, nil, 0},
		{5, nil, nil, 0},
		{7, nil, nil, 0},
	}

	// Initialize the finger tables for each node
	for _, node := range allNodes {
		_initFingerTable(node, allNodes)
	}

	// Assert that the finger table is initialized correctly for all nodes
	for _, node := range allNodes {

		// Check if the finger table has the correct number of entries
		if len(node.FingerTable) != bit_space {
			t.Errorf("Expected %d finger table entries, got %d", bit_space, len(node.FingerTable))
		}

		fmt.Printf("Node %d passed the finger table length test\n", node.Id)

		// Check if the finger table entries are initialized correctly
		for i, finger := range node.FingerTable {
			expectedStart := (node.Id + int(math.Pow(2, float64(i)))) % int(math.Pow(2, bit_space))
			if finger.Start != expectedStart {
				t.Errorf("Expected Start value %d for finger %d, got %d", expectedStart, i, finger.Start)
			}
		}

		fmt.Printf("Node %d passed the finger table Start values test\n", node.Id)

		// Check if the successor nodes are set correctly
		for i, finger := range node.FingerTable {
			expectedSuccessor := findSuccessor(finger.Start, allNodes)
			if finger.successor != expectedSuccessor {
				t.Errorf("Expected successor node %v for finger %d, got %v", expectedSuccessor, i, finger.successor)
			}
		}

		fmt.Printf("Node %d passed the finger table successor nodes test\n", node.Id)
		fmt.Println()
	}

	fmt.Println("All tests passed!")
}
