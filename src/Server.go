package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Node struct {
	Id            int            `json:"id"`
	FingerTable   []*FingerEntry `json:"finger_table"`
	SuccessorID   *NodeAddress   `json:"successorID"`
	PredecessorID *NodeAddress   `json:"predecessorID"`
	Address       string         `json:"address"`
}

type FingerEntry struct {
	Start       int          `json:"start"`
	SuccessorID *NodeAddress `json:"successorID"`
}

type NodeAddress struct {
	Id      int    `json:"id"`
	Address string `json:"address"`
}

type Server struct {
	hostname string
	port     string
	node     *Node
	server   *http.Server
	storage  map[string]string
}

var serverInstance *Server
var keyIdentifierSpace int

func InitServer(node *Node) {

	addressParts := strings.Split(node.Address, ":")

	keyIdentifierSpace = len(node.FingerTable)

	// Create a new server instance
	serverInstance = &Server{
		hostname: addressParts[0],
		port:     addressParts[1],
		node:     node,
		storage:  make(map[string]string),
	}

	serverInstance.server = &http.Server{
		Addr:    ":" + serverInstance.port,
		Handler: initMux(),
	}

	if serverInstance.node.Id == 0 {
		//serverInstance.storage["hei"] = "Hello, World!"
		//serverInstance.storage["hello"] = "Hello, World! 2"
	}

	fmt.Printf("\nServer initialized at: %s and node ID %d\n", serverInstance.hostname+":"+serverInstance.port, serverInstance.node.Id)

	// // Looping through the finger table of the node
	// for _, finger := range serverInstance.node.FingerTable {
	// 	fmt.Printf("Finger start: %d, Finger successor ID: %d\n", finger.Start, finger.SuccessorID.Id)
	// }

	// fmt.Printf("\n")

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

func hash(input string) (uint64) {

	// Hash the input using SHA-256
	hash := sha256.Sum256([]byte(input))

	// Convert the first 8 bytes of the hash to a uint64
	hashedValue := binary.BigEndian.Uint64(hash[:8])

	// Apply modulo 2^n to restrict the result between 0 and 2^n - 1
	maxValue := uint64(1<<keyIdentifierSpace) - 1
	return hashedValue % maxValue
}

func initMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/helloworld", helloworldHandler)
	mux.HandleFunc("/storage/", storageHandler)
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

func (s *Server) findSuccessor(key int) *NodeAddress {

	// Check if the key is in the finger table
	for _, finger := range s.node.FingerTable {
		if s.node.Id > finger.SuccessorID.Id {
			if key > s.node.Id || key <= finger.SuccessorID.Id {
				return finger.SuccessorID
			}
		} else if key <= finger.SuccessorID.Id {
			return finger.SuccessorID
		}

	}

	// If the key is not in the finger table, return the last entry
	return s.node.FingerTable[len(s.node.FingerTable)-1].SuccessorID
}

func httpReq(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Get(url)
}

func forwardGetStorageRequest(w http.ResponseWriter, address string, key string, value string) {
	// Format the URL
	url := fmt.Sprintf("http://%s/storage/%s", address, key)

	// Forward the request to the given node
	req, err := http.NewRequest("PUT", url, strings.NewReader(value))
    if err != nil {
        http.Error(w, "Error creating request", http.StatusInternalServerError)
        return
    }

	client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
        w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Error forwarding request to successor node", http.StatusInternalServerError)
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// GET: Returns HTTP code 200, with value, if <key> exists in the DHT. Returns HTTP code 404, if <key> does not exist in the DHT.
// PUT: Returns HTTP code 200. Assumed that <value> is persisted
func storageHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		key := strings.TrimPrefix(r.URL.Path, "/storage/")
		hashedKey := int(hash(key))

		if serverInstance.node.Id > serverInstance.node.PredecessorID.Id {
			if hashedKey <= serverInstance.node.Id || hashedKey > serverInstance.node.PredecessorID.Id {
				// Key must exist in my storage DHT
				value, ok := serverInstance.storage[key]
				if ok {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "text/plain")
					w.Write([]byte(value))
					return
				} else {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
		}

		// Find the successor node for the key
		successor := serverInstance.findSuccessor(hashedKey)
		url := fmt.Sprintf("http://%s/storage/%d", successor.Address, hashedKey)

		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Redirecting to: " + url))

		resp, err := httpReq(url)
		if err != nil {
			http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
			return
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Error reading response from successor node", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("\n" + string(body)))
			return
		} else if resp.StatusCode == http.StatusNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

	} else if r.Method == "PUT" {
		key := strings.TrimPrefix(r.URL.Path, "/storage/")
		hashedKey := hash(key)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading body:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		value := string(body)
		
		// Check if the hashed key is within the range of the current node
		if int(hashedKey) <= serverInstance.node.Id && int(hashedKey) > serverInstance.node.PredecessorID.Id {
			// If the key already exists in the storage, return 403 Forbidden
			if _, exists := serverInstance.storage[key]; exists {
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("Key already exists in system"))
				return
			}

			// If not, add the key to the storage
			serverInstance.storage[key] = value
			w.WriteHeader(http.StatusOK)
			return
		}












		// Add key to DHT
		// 	body, err := io.ReadAll(r.Body)
		// 	if err != nil {
		// 		fmt.Println("Error reading body:", err)
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		return
		// 	}
		// 	defer r.Body.Close()

		// 	// Assume body contains SuccessorID as plain text
		// 	var successorID int
		// 	if err := json.Unmarshal(body, &successorID); err != nil {
		// 		http.Error(w, "Invalid body format", http.StatusBadRequest)
		// 		return
		// 	}

		// 	if finger.SuccessorID == successorID {
		// 		w.WriteHeader(http.StatusOK)
		// 		w.Header().Set("Content-Type", "text/plain")
		// 		w.Write([]byte(fmt.Sprintf("Found: %+v", finger)))
		// 		return
		// 	}
		// }
		// w.WriteHeader(http.StatusNotFound)
	}
}

// Returns HTTP code 200, with list of known nodes, as JSON.
func networkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return list of known nodes
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(
			"Known nodes:\n",
		))
		for _, node := range serverInstance.node.FingerTable {
			w.Write([]byte(fmt.Sprintf("NodeID: %d\n", node.SuccessorID.Id)))
		}

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

	nodeID, err := strconv.Atoi(os.Args[1])

	if err != nil {
		fmt.Println("Error parsing node ID:", err)
		return
	}

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
