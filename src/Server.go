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

	fmt.Printf("\nServer initialized at: %s and node ID %d\n", serverInstance.hostname+":"+serverInstance.port, serverInstance.node.Id)

	// Channel to listen for shutdown signal (interrupts or timer)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start the server
	go startServer()

	// Start the server shutdown timer
	go startServerShutdownTimer(shutdownChan)

	// Start the periodic finger table update
	// go periodicUpdateFingerTable()

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
	mux.HandleFunc("/getkeys", transferKeysHandler)
	mux.HandleFunc("/network", networkHandler)
	mux.HandleFunc("/node-info", nodeInfoHandler)
	mux.HandleFunc("/leave", leaveHandler)
	mux.HandleFunc("/sim-crash", simulateCrash)
	mux.HandleFunc("/sim-recover", simulateRecover)
	mux.HandleFunc("/join", joinRingHandler)

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
	// First, check if the key falls between the current node and its immediate successor
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
	return s.node.FingerTable[len(s.node.FingerTable)-1].SuccessorID
}

func (s *Server) findClosestPredecessor(key int) *NodeAddress {

	// Iterate through the finger table in reverse order
	for i := len(s.node.FingerTable) - 1; i >= 0; i-- {
		finger := s.node.FingerTable[i]

		// fmt.Printf("Checking finger %d: %d\n", i, finger.SuccessorID.Id)

		// Check if the finger points to a node that is a valid predecessor of the key
		// and that the finger node is closer to the key than the current node
		if isBetween(s.node.Id, finger.SuccessorID.Id, key) {
			// fmt.Printf("Found closest predecessor: %d\n", finger.SuccessorID.Id)
			return finger.SuccessorID
		}
	}

	// Return the closest valid predecessor found
	return s.node.FingerTable[len(s.node.FingerTable)-1].SuccessorID
}

// Helper function to check if 'key' is in the interval (n1, n2] with wraparound handling
func isBetweenInclusive(n1, key, n2 int) bool {
	if n1 < n2 {
		return key > n1 && key <= n2
	}
	return key > n1 || key <= n2
}

// Helper function to check if 'key' is in the interval (n1, n2) with wraparound handling
func isBetween(n1, key, n2 int) bool {
	if n1 < n2 {
		return key > n1 && key < n2
	}
	return key > n1 || key < n2
}

func get_keys_in_range(start, end int) []string {
	keys := make([]string, 0)
	for key, _ := range serverInstance.storage {
		keyInt := hash(key)
		fmt.Printf("Key: %s Hash: %d Start: %d End: %d\n", key, keyInt, start, end)
		if isBetween(start, keyInt, end) {
			keys = append(keys, key)
		}
	}
	return keys
}

func transferKeysHandler(w http.ResponseWriter, r *http.Request) {

	// Incomming node ID
	nodeID := r.URL.Query().Get("nodeID")

	if nodeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeIDInt, err := strconv.Atoi(nodeID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("Transferring keys to node %d\n", nodeIDInt)

	// Get the keys in the range of the given node
	keys := get_keys_in_range(serverInstance.node.PredecessorID.Id, nodeIDInt)

	fmt.Printf("Keys to transfer to %d: %v\n", nodeIDInt, keys)
}

// GET: Returns HTTP code 200, with value, if <key> exists in the DHT. Returns HTTP code 404, if <key> does not exist in the DHT.
// PUT: Returns HTTP code 200. Assumed that <value> is persisted
func storageHandler(w http.ResponseWriter, r *http.Request) {

	s := serverInstance

	if r.Method == "GET" {

		key := strings.TrimPrefix(r.URL.Path, "/storage/")
		keyInt := hash(key)

		// Check if the key is within the valid range
		if keyInt < 0 || keyInt >= 1<<keyIdentifierSpace || fmt.Sprintf("%T", keyInt) != "int" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		curr_node := s.node.Id
		prev_node := s.node.PredecessorID.Id

		if curr_node == prev_node {
			_, ok := s.storage[key]
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(s.storage[key]))
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}

		// Checking for wrap-around in the ring
		if prev_node > curr_node {
			if keyInt <= curr_node || keyInt > prev_node {

				// Check local storage
				value, ok := s.storage[key]
				if ok {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(value))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
				return
			}
		} else if keyInt > prev_node && keyInt <= curr_node {
			// If the key falls between the current node and its predecessor, return the value

			value, ok := s.storage[key]
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(value))
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}

		// Find the successor node for the given key
		successor := s.findSuccessor(keyInt)

		// If the successor is the current node, return the value
		if successor.Address == s.node.Address {
			value, ok := s.storage[key]
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(value))
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}

		// Forward the request to the successor node
		url := fmt.Sprintf("http://%s/storage/%s", successor.Address, key)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)

		if err != nil {
			http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
			return
		}

		// Handle the response
		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)

			if err != nil {
				http.Error(w, "Error reading response from successor node", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(body)

		} else if resp.StatusCode == http.StatusNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
		}
		return

	} else if r.Method == "PUT" {

		key := strings.TrimPrefix(r.URL.Path, "/storage/")
		keyInt := hash(key)

		if keyInt < 0 || keyInt >= 1<<keyIdentifierSpace || fmt.Sprintf("%T", keyInt) != "int" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading body:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		value := string(body)

		curr_node := s.node.Id
		prev_node := s.node.PredecessorID.Id

		// If the current node is the only node in the ring, store the value
		if prev_node == curr_node {
			// Check local storage
			_, ok := s.storage[key]
			if ok {
				w.WriteHeader(http.StatusForbidden)
			} else {
				// Store the value if the key is not already present
				s.storage[key] = value
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		// Checking for wrap-around in the ring
		if prev_node > curr_node {
			if keyInt <= curr_node || keyInt > prev_node {

				// Check local storage
				_, ok := s.storage[key]
				if ok {
					w.WriteHeader(http.StatusForbidden)
				} else {
					// Store the value if the key is not already present
					s.storage[key] = value
					w.WriteHeader(http.StatusOK)
				}
				return
			}
		} else if keyInt > prev_node && keyInt <= curr_node {
			// If the key falls between the current node and its predecessor, store the value
			_, ok := s.storage[key]
			// Store the value only if the key is not already present
			if ok {
				w.WriteHeader(http.StatusForbidden)
			} else {
				s.storage[key] = value
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		// Find the successor node for the given key
		successor := s.findSuccessor(keyInt)

		// If the successor is the current node, store the value
		if successor.Address == s.node.Address {
			_, ok := s.storage[key]
			// Store the value only if the key is not already present
			if ok {
				w.WriteHeader(http.StatusForbidden)
			} else {
				s.storage[key] = value
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		// Forward the request to the successor node
		url := fmt.Sprintf("http://%s/storage/%s", successor.Address, key)

		// Forward the request to the given node
		req, err := http.NewRequest("PUT", url, strings.NewReader(value))
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		// Set the content type and length
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Handle the response
		if resp.StatusCode == http.StatusOK {
			w.WriteHeader(http.StatusOK)
			serverInstance.storage[key] = value
		} else {
			http.Error(w, "Error forwarding request to successor node", http.StatusInternalServerError)
		}
		return
	}
}

// Returns HTTP code 200, with list of known nodes, as a JSON array of strings.
func networkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Collect known node addresses into a list
		nodes := make([]string, 0)
		for _, node := range serverInstance.node.FingerTable {
			nodes = append(nodes, node.SuccessorID.Address)
		}

		// Convert the list of node addresses to JSON
		jsonData, err := json.Marshal(nodes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error encoding JSON"))
			return
		}

		// Set content type and return the JSON data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

func (s *Server) create_info_interface() map[string]interface{} {
	data := make(map[string]interface{})
	data["node_hash"] = s.node.Id
	data["address"] = s.node.Address

	if s.node.SuccessorID == nil {
		data["successor"] = "nil"
	} else {
		data["successor"] = s.node.SuccessorID.Address
	}

	others := make([]string, 0)
	for _, node := range s.node.FingerTable {
		if node.SuccessorID == nil {
			others = append(others, "nil")
		} else {
			others = append(others, node.SuccessorID.Address)
		}
	}

	data["others"] = others
	return data
}

func nodeInfoHandler(w http.ResponseWriter, r *http.Request) {

	s := serverInstance

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return

	} else if r.Method == http.MethodGet {

		data := s.create_info_interface()

		askingId := r.URL.Query().Get("successor")
		if askingId != "" {
			keyInt, err := strconv.Atoi(askingId)
			key := "successor_of_" + strconv.Itoa(keyInt)

			if err != nil || keyInt < 0 || keyInt >= 1<<keyIdentifierSpace {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			curr_node := s.node.Id
			prev_node := s.node.PredecessorID.Id
			data[key] = s.node.Address

			// If the current node is the only node in the ring, return it self
			if curr_node == prev_node {
				send_node_info(w, data)
				return
			}

			// Checking for wrap-around in the ring
			if prev_node > curr_node {
				if keyInt <= curr_node || keyInt > prev_node {
					send_node_info(w, data)
					return
				}
			} else if keyInt > prev_node && keyInt <= curr_node {
				// If the key falls between the current node and its predecessor, return the value
				send_node_info(w, data)
				return
			}

			successor := s.findSuccessor(keyInt)

			// If the successor is the current node, return it self
			if successor.Address == s.node.Address {
				send_node_info(w, data)
				return

			} else {

				// Forward the request to the successor node
				url := fmt.Sprintf("http://%s/node-info?successor=%d", successor.Address, keyInt)

				client := &http.Client{Timeout: 10 * time.Second}
				resp, err := client.Get(url)

				if err != nil {
					http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
					return
				}

				// Handle the response
				if resp.StatusCode == http.StatusOK {

					var successorData map[string]interface{}
					decoder := json.NewDecoder(resp.Body)
					err = decoder.Decode(&successorData)

					if err != nil {
						http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
						return
					}

					data["address"] = successorData["address"]
					data["successor"] = successorData["successor"]
					data["others"] = successorData["others"]
					data[key] = successorData[key]
					send_node_info(w, data)

				} else if resp.StatusCode == http.StatusNotFound {
					w.WriteHeader(http.StatusNotFound)
				} else {
					http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
				}
			}
			return
		}

		send_node_info(w, data)
		return
	}
}

func send_node_info(w http.ResponseWriter, data map[string]interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error encoding JSON"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func get_response(w http.ResponseWriter, url string) *http.Response {

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		http.Error(w, "Error connecting to successor node", http.StatusInternalServerError)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Error getting node info", http.StatusInternalServerError)
		return nil
	}

	return resp
}

func joinRingHandler(w http.ResponseWriter, r *http.Request) {

	s := serverInstance

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return

	} else if r.Method == http.MethodPost {

		successorID := r.URL.Query().Get("nprime")
		if successorID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		successorInfo := fmt.Sprintf("http://%s/node-info", successorID)
		resp := get_response(w, successorInfo)
		if resp == nil {
			return
		}

		var successor map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&successor)

		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
			return
		}

		// Sending a request to the successor node to get the node info
		nodeInfo := fmt.Sprintf("http://%s/node-info?successor=%d", successorID, s.node.Id)
		resp = get_response(w, nodeInfo)
		if resp == nil {
			return
		}

		// Decode the JSON response
		var data map[string]interface{}
		decoder = json.NewDecoder(resp.Body)
		err = decoder.Decode(&data)

		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
			return
		}

		// Update the successor of the current node
		key := "successor_of_" + strconv.Itoa(s.node.Id)
		request := fmt.Sprintf("http://%s/node-info", data[key].(string))
		resp = get_response(w, request)
		if resp == nil {
			return
		}

		var successorData map[string]interface{}
		decoder = json.NewDecoder(resp.Body)
		err = decoder.Decode(&successorData)

		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
			return
		}

		s.node.SuccessorID = &NodeAddress{
			Id:      int(successorData["node_hash"].(float64)),
			Address: successorData["address"].(string),
		}

		fmt.Printf("SUCCESSOR: %v\n", s.node.SuccessorID)

		// Get keys from the storage from the successor node that should be transferred to the current node
		// keys := make([]string, 0)

		// Iterate through the storage of the successor node
		request = fmt.Sprintf("http://%s/storage", s.node.SuccessorID.Address)
		resp = get_response(w, request)
		if resp == nil {
			return
		}

		var storage map[string]string
		decoder = json.NewDecoder(resp.Body)
		err = decoder.Decode(&storage)

		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
			return
		}

		fmt.Printf("STORAGE: %v\n", storage)

		// // Iterate through the storage of the successor node
		// for key, _ := range storage {
		// 	keyInt := hash(key)
		// 	if isBetween(s.node.Id, keyInt, s.node.SuccessorID.Id) {
		// 		keys = append(keys, key)
		// 	}
		// }
	}
}

// func periodicUpdateFingerTable() {
// 	ticker := time.NewTicker(5 * time.Second)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		fmt.Println("Updating Finger Table...")
// 		updateFingerTable()
// 	}
// }

// func updateFingerTable() {
// 	// s := serverInstance
// }

// func (s *Server) updateFingerTable(node *NodeAddress, i int) {

// 	// Check if the node is the immediate successor of the current node
// 	if isBetween(s.node.Id, node.Id, s.node.FingerTable[i].SuccessorID.Id) {
// 		s.node.FingerTable[i].SuccessorID = node

// 		// Update the successor of the current node
// 		if i == 0 {
// 			s.node.SuccessorID = node
// 		}

// 		// Update the predecessor of the successor node
// 		if i == len(s.node.FingerTable)-1 {
// 			s.updatePredecessor(node)
// 		}
// 	}
// }

// func (s *Server) updatePredecessor(node *NodeAddress) {
// 	// Check if the node is the immediate predecessor of the current node
// 	if isBetween(s.node.PredecessorID.Id, node.Id, s.node.Id) {
// 		s.node.PredecessorID = node
// 	}
// }

func leaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	} else if r.Method == http.MethodPost {
		// TODO: Implement leave handler
	}
}

func simulateCrash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	} else if r.Method == http.MethodPost {
		// TODO: Implement crash simulation
	}
}

func simulateRecover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	} else if r.Method == http.MethodPost {
		// TODO: Implement recovery simulation
	}
}

func createNewNode() {

	// Creates a new id by hashing a random number
	id := hash(strconv.Itoa(int(time.Now().UnixNano())))
	for id == 0 || id == 4 || id == 8 || id == 12 {
		id = hash(strconv.Itoa(int(time.Now().UnixNano())))
	}

	address := os.Args[3]
	fingerTable := make([]*FingerEntry, keyIdentifierSpace)

	for i := 0; i < keyIdentifierSpace; i++ {
		fingerTable[i] = &FingerEntry{
			Start:       -1,
			SuccessorID: nil,
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
