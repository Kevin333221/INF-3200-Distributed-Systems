package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func periodicUpdateFingerTable() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stabilize()
		checkPredecessor()
		updateFingerTable()
	}
}

func stabilize() {

	// Psudo code
	// 1. x = successor.predecessor
	// 2. if x is between current node and successor
	// 3. 	successor = x
	// 4. notify successor

	s := serverInstance
	successor := s.node.SuccessorID

	// Get the predecessor of the successor node
	request := fmt.Sprintf("http://%s/node-info?successor=%d", successor.Address, s.node.Id)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(request)

	if err != nil {
		return
	}

	var data map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)

	if err != nil {
		return
	}

	// Get info about the successor node
	request = fmt.Sprintf("http://%s/node-info", data["address"].(string))
	client = &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Get(request)

	if err != nil {
		return
	}

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)

	if err != nil {
		return
	}

	if data["predecessor"] == nil {
		notify(successor.Address)
		return
	}

	predecessorData := data["predecessor"].(map[string]interface{})
	predecessorAddress := predecessorData["address"].(string)

	// Get the predecessor of the successor node
	request = fmt.Sprintf("http://%s/node-info", predecessorAddress)
	client = &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Get(request)

	if err != nil {
		return
	}

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)

	if err != nil {
		return
	}

	predecessor := data

	// Check if the predecessor of the successor node is between the current node and the successor
	if isBetween(s.node.Id, int(predecessor["id"].(float64)), successor.Id) {
		s.node.SuccessorID = &NodeAddress{
			Id:      int(predecessor["id"].(float64)),
			Address: predecessor["address"].(string),
		}
	}

	// Notify the successor node
	notify(successor.Address)
}

func updateFingerTable() {
	// Psudo code
	// next = next + 1
	// if next > m
	// 	next = 1
	// finger[next].node = find_successor(n + 2^(next-1))

	s := serverInstance

	for i := 0; i < keyIdentifierSpace; i++ {

		// Calculate the next finger entry
		next := (s.node.Id + 1<<i) % (1 << keyIdentifierSpace)
		finger := s.node.FingerTable[i]

		successor := s.findSuccessor(next)

		// Get the successor node for the next finger entry
		url := fmt.Sprintf("http://%s/node-info?successor=%d", successor.Address, next)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)

		if err != nil || resp.StatusCode != http.StatusOK {
			return
		}

		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&data)

		if err != nil {
			return
		}

		url = fmt.Sprintf("http://%s/node-info", data["address"].(string))
		client = &http.Client{Timeout: 10 * time.Second}
		resp, err = client.Get(url)

		if err != nil || resp.StatusCode != http.StatusOK {
			return
		}

		decoder = json.NewDecoder(resp.Body)
		err = decoder.Decode(&data)

		if err != nil {
			return
		}

		finger.SuccessorID = &NodeAddress{
			Id:      int(data["id"].(float64)),
			Address: data["address"].(string),
		}
	}
}

func checkPredecessor() {
	// Psudo code
	// if predecessor has failed
	// 	predecessor = nil

	s := serverInstance

	if s.node.PredecessorID == nil {
		return
	}

	request := fmt.Sprintf("http://%s/node-info", s.node.PredecessorID.Address)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(request)

	if err != nil {
		s.node.PredecessorID = nil
		return
	}

	// If the predecessor node has crashed, set the predecessor to nil
	if resp.StatusCode != http.StatusOK {
		s.node.PredecessorID = nil
		return
	}

	var data map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)

	if err != nil {
		s.node.PredecessorID = nil
		return
	}

	successorData := data["successor"].(map[string]interface{})
	successorAddress := successorData["address"].(string)

	if successorAddress != s.node.Address {
		s.node.PredecessorID = nil
	}
}

func notify(address string) {
	// Psudo code
	// if predecessor is nil or n' is between predecessor and n
	// 	predecessor = n'

	s := serverInstance

	request := fmt.Sprintf("http://%s/node-info", address)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(request)

	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return
	}

	var data map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)

	if err != nil {
		return
	}

	if s.node.PredecessorID == nil || isBetween(s.node.PredecessorID.Id, int(data["id"].(float64)), s.node.Id) {
		s.node.PredecessorID = &NodeAddress{
			Id:      int(data["id"].(float64)),
			Address: data["address"].(string),
		}
	}
}
