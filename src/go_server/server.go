package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	port     string
	hostname string
)

func InitServer() {
	port = os.Args[1]
	hostname = strings.Split(getHostName(), ".")[0]

	server := &http.Server{
		Addr:    ":" + port,
		Handler: initMux(),
	}

	// Channel to listen for shutdown signal (interrupts or timer)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start the server
	go startServer(server, port)

	// Start the server shutdown timer
	go startServerShutdownTimer(shutdownChan)

	// Wait for the shutdown signal
	<-shutdownChan

	// Shutdown the server
	shutdownServer(server)

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

func shutdownServer(server *http.Server) {
	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Server forced to shutdown:", err)
	}
}

func startServer(server *http.Server, port string) {
	// Start the server in a separate goroutine
	fmt.Println("Server is running on port", port)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Could not listen on %s: %v\n", port, err)
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
		// Get list of known nodes
	}
}

func helloworldHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	} else if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(hostname + ":" + port))
	}
}
