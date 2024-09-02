package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

var (
	port     string
	hostname string
)

func getHostName() string {
	name, hostNameError := os.Hostname()
	if hostNameError != nil {
		fmt.Println("Error: ", hostNameError)
		return ""
	}
	return name
}

func main() {
	port = os.Args[1]
	hostname = strings.Split(getHostName(), ".")[0]
	http.HandleFunc("/helloworld", helloWorldHandler)
	http.ListenAndServe(":"+port, nil)
}

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(hostname + ":" + port))
}
