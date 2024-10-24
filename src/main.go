package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func main() {

	nodeID, err := strconv.Atoi(os.Args[1])
	newNode := os.Args[2]

	if err != nil {
		fmt.Println("Error parsing node ID:", err)
		return
	}

	if newNode == "true" {
		fmt.Println("Created new node")
		keyIdentifierSpace, err = strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Error parsing key identifier space:", err)
			return
		}

		createNewNode()

	} else {

		// Read data from "Nodes.json"
		file, err := os.Open("DeployServers/Nodes.json")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		var nodes []*Node
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&nodes)

		if err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		var foundNode *Node
		for _, node := range nodes {
			if node.Id == nodeID {
				foundNode = node
				break
			}
		}

		if foundNode == nil {
			fmt.Println("Node not found")
			return
		}

		InitServer(foundNode)
	}
}
