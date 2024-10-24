package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func InitServer(node *Node) {

	addressParts := strings.Split(node.Address, ":")

	keyIdentifierSpace = len(node.FingerTable)

	// Create a new server instance
	serverInstance = &Server{
		hostname: addressParts[0],
		port:     addressParts[1],
		node:     node,
		storage:  make(map[string]string),
		crashed:  false,
	}

	serverInstance.server = &http.Server{
		Addr:    ":" + serverInstance.port,
		Handler: initMux(),
	}

	fmt.Printf("\nServer initialized at: %s and node ID %d\n", serverInstance.hostname+":"+serverInstance.port, serverInstance.node.Id)

	// Channel to listen for shutdown signal (interrupts or timer)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start the server
	go startServer()

	// Start the server shutdown timer
	go startServerShutdownTimer(shutdownChan)

	// Start the periodic finger table update
	go periodicUpdateFingerTable()

	// Wait for the shutdown signal
	<-shutdownChan

	// Shutdown the server
	shutdownServer()

	fmt.Println("Server exiting")
}

func hash(input string) int {

	// Hash the input using SHA-256
	hash := sha256.Sum256([]byte(input))

	// Convert the first 8 bytes of the hash to a uint64
	hashedValue := binary.BigEndian.Uint64(hash[:8])

	// Apply modulo 2^n to restrict the result between 0 and 2^n - 1
	return int(hashedValue % uint64(1<<keyIdentifierSpace))
}

func initMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/helloworld", helloworldHandler)
	mux.HandleFunc("/storage/", storageHandler)
	mux.HandleFunc("/network", networkHandler)
	mux.HandleFunc("/node-info", nodeInfoHandler)
	mux.HandleFunc("/leave", leaveHandler)
	mux.HandleFunc("/sim-crash", simulateCrashHandler)
	mux.HandleFunc("/sim-recover", simulateRecoverHandler)
	mux.HandleFunc("/join", joinRingHandler)
	mux.HandleFunc("/update-successor", updateSuccessorHandler)
	mux.HandleFunc("/update-predecessor", updatePredecessorHandler)

	return mux
}

func shutdownServer() {
	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serverInstance.server.Shutdown(ctx); err != nil {
		fmt.Println("Server forced to shutdown:", err)
	}
}

func startServer() {
	// Start the server in a separate goroutine
	err := serverInstance.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Could not listen on %s: %v\n", serverInstance.port, err)
	}
}

func startServerShutdownTimer(shutdownChan chan os.Signal) {
	// Timer to shut down the server after 10 minutes
	time.Sleep(10 * time.Minute)
	fmt.Println("Shutting down the server after 10 minutes...")
	shutdownChan <- os.Interrupt
}

func (s *Server) findSuccessor(key int) *NodeAddress {

	// First, check if the key falls between the current node and its immediate successor (me, successor]
	if isBetweenInclusive(s.node.Id, key, s.node.SuccessorID.Id) {
		return s.node.SuccessorID
	}

	// Otherwise, look in the finger table for the closest predecessor
	closestPredecessor := s.findClosestPredecessor(key)

	// Recursively call findSuccessor on the closest predecessor if it's not nil
	if closestPredecessor != nil {
		return closestPredecessor
	}

	// If no closer predecessor is found, return the successor as fallback
	return s.node.SuccessorID
}

func (s *Server) findClosestPredecessor(key int) *NodeAddress {

	is_nil := false
	for _, finger := range s.node.FingerTable {
		if finger.SuccessorID == nil {
			is_nil = true
		}
	}

	if is_nil {
		return s.node.SuccessorID
	}

	// Iterate through the finger table in reverse order
	for i := len(s.node.FingerTable) - 1; i >= 0; i-- {
		finger := s.node.FingerTable[i]

		// Check if the finger points to a node that is a valid predecessor of the key
		// and that the finger node is closer to the key than the current node
		if isBetween(s.node.Id, finger.SuccessorID.Id, key) {
			return finger.SuccessorID
		}
	}

	// Return myself
	return s.node.SuccessorID

	// return s.node.FingerTable[len(s.node.FingerTable)-1].SuccessorID
}

func get_response(url string) *http.Response {
	client := &http.Client{Timeout: 10 * time.Second}
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, _ := client.Do(request)
	return resp
}

func put_request(url string, jsonData []byte) *http.Response {
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest(http.MethodPut, url, strings.NewReader(string(jsonData)))

	req.Header.Set("Content-Type", "application/json")
	resp, _ := client.Do(req)
	return resp
}

// Additional functions
func updateSuccessor(address_from NodeAddress, address_to *NodeAddress) {
	request := fmt.Sprintf("http://%s/update-successor", address_from.Address)
	jsonData, _ := json.Marshal(address_to)
	put_request(request, jsonData)
}

func updatePredecessor(address_from NodeAddress, address_to *NodeAddress) {
	request := fmt.Sprintf("http://%s/update-predecessor", address_from.Address)
	jsonData, _ := json.Marshal(address_to)
	put_request(request, jsonData)
}

func getNode(address string) map[string]interface{} {

	fmt.Printf("Fetchin info from %s\n", address)

	request := fmt.Sprintf("http://%s/node-info", address)
	resp := get_response(request)

	var data map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&data)

	return data
}

func createNewNode() {

	// Creates a new id by hashing a random number
	id := hash(strconv.Itoa(int(time.Now().UnixNano())))

	address := os.Args[3]
	fingerTable := make([]*FingerEntry, keyIdentifierSpace)

	for i := 0; i < keyIdentifierSpace; i++ {
		fingerTable[i] = &FingerEntry{
			Start:       int(math.Pow(2, float64(i))),
			SuccessorID: &NodeAddress{Id: id, Address: address},
		}
	}

	newNode := &Node{
		Id:            id,
		FingerTable:   fingerTable,
		SuccessorID:   &NodeAddress{Id: id, Address: address},
		PredecessorID: nil,
		Address:       address,
	}

	InitServer(newNode)
}

// Helper function to check if 'key' is in the interval (n1, n2] with wraparound handling
func isBetweenInclusive(n1, key, n2 int) bool {
	if n1 < n2 {
		return n1 < key && key <= n2
	}
	return n1 < key || key <= n2
}

// Helper function to check if 'key' is in the interval (n1, n2) with wraparound handling
func isBetween(n1, key, n2 int) bool {
	if n1 < n2 {
		return n1 < key && key < n2
	}
	return n1 < key || key < n2
}
