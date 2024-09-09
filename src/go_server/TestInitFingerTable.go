package main

import (
	"fmt"
	"testing"
)

func TestInitFingerTable(t *testing.T) {

	// Create a list of nodes
	allNodes := []*Node{
		{1, nil, nil},
		{3, nil, nil},
		{5, nil, nil},
		{7, nil, nil},
	}

	for _, node := range allNodes {
		// Call the initFingerTable function
		initFingerTable(node, allNodes)
	}

	fmt.Printf("Node 0: %d\n", len(allNodes[0].fingerTable))

	// // Assert that the finger table is initialized correctly for all nodes
	// for _, node := range allNodes {
	// 	// Check if the finger table has the correct number of entries
	// 	if len(node.fingerTable) != m {
	// 		t.Errorf("Expected %d finger table entries, got %d", m, len(node.fingerTable))
	// 	} else {
	// 		t.Logf("Node %d has %d finger table entries", node.id, len(node.fingerTable))
	// 	}

	// 	// Check if the finger table entries are initialized correctly
	// 	for i, finger := range node.fingerTable {
	// 		expectedStart := (node.id + int(math.Pow(2, float64(i)))) % int(math.Pow(2, m))
	// 		if finger.start != expectedStart {
	// 			t.Errorf("Expected start value %d for finger %d, got %d", expectedStart, i, finger.start)
	// 		} else {
	// 			t.Logf("Node %d has finger %d with start value %d", node.id, i, finger.start)
	// 		}
	// 	}

	// 	// Check if the successor nodes are set correctly
	// 	for i, finger := range node.fingerTable {
	// 		expectedSuccessor := findSuccessor(finger.start, allNodes)
	// 		if finger.successor != expectedSuccessor {
	// 			t.Errorf("Expected successor node %v for finger %d, got %v", expectedSuccessor, i, finger.successor)
	// 		} else {
	// 			t.Logf("Node %d has finger %d with successor node %v", node.id, i, finger.successor)
	// 		}
	// 	}
	// }
}
