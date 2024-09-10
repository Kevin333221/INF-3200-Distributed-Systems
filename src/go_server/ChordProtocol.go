package main

import (
	"math"
)

type Node struct {
	Id          int            `json:"id"`
	FingerTable []*FingerEntry `json:"finger_table"`
	successor   *Node
	SuccessorID int `json:"successorID"`
}

type FingerEntry struct {
	Start       int `json:"start"`
	successor   *Node
	SuccessorID int `json:"successorID"`
}

// FingerTable initialization for a node
// initFingerTable initializes the finger table for a given node.
// It calculates the start and successor for each entry in the finger table using the following formulas:
//
// finger[i].start = (n.id + 2^(i-1)) % 2^m
//
// finger[i].successor = findSuccessor(start, allNodes)
//
// Parameters:
// - n: The node for which the finger table is being initialized.
// - allNodes: A slice containing all the nodes in the system.
//
// Returns: None
func _initFingerTable(n *Node, allNodes []*Node) {
	for i := 1; i <= m; i++ {
		start := (n.Id + int(math.Pow(2, float64(i-1)))) % int(math.Pow(2, float64(m))) // Start value for the finger entry
		successor := findSuccessor(start, allNodes)                            // Find the successor node for the start value

		finger := &FingerEntry{
			Start:       start,
			successor:   successor, // Find the successor node for the start value
			SuccessorID: successor.Id,
		}
		n.FingerTable = append(n.FingerTable, finger) // Add the finger entry to the finger table
	}
}

// Find the successor node for a given key
// findSuccessor finds the successor node for a given key in a list of nodes.
// It iterates through all the nodes and returns the first node whose ID is greater than or equal to the key.
// If no such node is found, it wraps around to the first node in the list.
//
// Parameters:
// - key: The key for which the successor node needs to be found.
// - allNodes: A list of nodes to search for the successor.
//
// Returns:
// - The successor node for the given key.
func findSuccessor(key int, allNodes []*Node) *Node {
	for _, node := range allNodes { // Iterate through all nodes to find the successor
		if node.Id >= key { // If the node ID is greater than or equal to the key
			return node // Return the node as the successor
		}
	}
	return allNodes[0] // Wrap around to the first node
}

// Initialize the Chord ring
// initializeChordRing initializes the Chord ring with a specified number of nodes.
// It creates a list of nodes, spreads them evenly across the identifier space, and links them in a simple circle.

// Parameters: None
// Returns: None
func initializeChordRing() []*Node {

	// Initialize nodes in the Chord ring
	allNodes := make([]*Node, amount_nodes)

	// Need to spread the nodes across the identifier space evenly
	interval := int(math.Floor(math.Pow(2, float64(m)) / float64(amount_nodes)))
	for i := 0; i < amount_nodes; i++ {
		allNodes[i] = &Node{
			Id: i * interval,
		}
	}

	// Link nodes in a simple circle (successor)
	for i, node := range allNodes {
		node.successor = allNodes[(i+1)%len(allNodes)]
		node.SuccessorID = node.successor.Id
	}

	// Initialize finger tables for each node
	for _, node := range allNodes {
		_initFingerTable(node, allNodes)
	}

	return allNodes
}
