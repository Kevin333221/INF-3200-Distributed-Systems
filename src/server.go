package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Node struct {
	Id          int
	FingerTable []*FingerEntry
	SuccessorID int
}

type FingerEntry struct {
	Start       int
	SuccessorID int
}

type Server struct {
	port     string
	node     *Node
	hostname string
	server   *http.Server
}

var serverInstance *Server

func InitServer(port string, node *Node) {

	// Create a new server instance
	serverInstance = &Server{
		port:     port,
		node:     node,
		hostname: strings.Split(getHostName(), ".")[0],
		server: &http.Server{
			Addr:    ":" + port,
			Handler: initMux(),
		},
	}

	// fmt.Printf("\nServer initialized with port %s and node ID %d Server hostname: %s\n", serverInstance.port, serverInstance.node.Id, serverInstance.hostname)
	fmt.Printf("\nNode: %d\n", serverInstance.node)

	// // Looping through the finger table of the node
	// for _, finger := range serverInstance.node.FingerTable {
	// 	fmt.Printf("Finger start: %d, Finger successor ID: %d\n", finger.Start, finger.SuccessorID)
	// }

	// fmt.Printf("Successor ID: %d\n", serverInstance.node.SuccessorID)

	// Channel to listen for shutdown signal (interrupts or timer)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start the server
	go startServer()

	// Start the server shutdown timer
	go startServerShutdownTimer(shutdownChan)

	// Wait for the shutdown signal
	<-shutdownChan

	// Shutdown the server
	shutdownServer()

	fmt.Println("Server exiting")
}

func getHostName() string {
	name, hostNameError := os.Hostname()
	if hostNameError != nil {
		fmt.Println("Error: ", hostNameError)
		return ""
	}
	return name
}

func initMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/helloworld", helloworldHandler)
	mux.HandleFunc("/storage/<key>", storageHandler)
	mux.HandleFunc("/network", networkHandler)
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

// GET: Returns HTTP code 200, with value, if <key> exists in the DHT. Returns HTTP code 404, if <key> does not exist in the DHT.
// PUT: Returns HTTP code 200. Assumed that <value> is persisted
func storageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Check if key exists in the DHT
	} else if r.Method == "PUT" {
		// Add key to DHT
	}
}

// Returns HTTP code 200, with list of known nodes, as JSON.
func networkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return list of known nodes
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// json.NewEncoder(w).Encode(all_known_nodes)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func helloworldHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	} else if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(serverInstance.hostname + ":" + serverInstance.port))
	}
}

func main() {

	port := os.Args[1]
	nodeID := os.Args[2]

	// Read data from "Nodes.json"
	file, err := os.Open("go_server/Nodes.json")
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
		NodeID, _ := strconv.Atoi(nodeID)
		if node.Id == NodeID {
			foundNode = node
			break
		}
	}

	if foundNode == nil {
		fmt.Println("Node not found")
		return
	}

	InitServer(port, foundNode)
}
