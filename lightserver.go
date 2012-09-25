package main

import (
	"fmt"
	"net/http"
	"os/exec"
	/*"log"*/
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("tdtool", "--list")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	buffer := make([]byte, 1024)
	stdout.Read(buffer)
	fmt.Fprintf(w, "status: %s\n", buffer)
}

func main() {
	fmt.Printf("Hello world\n")
	http.HandleFunc("/status", statusHandler)
	http.ListenAndServe(":8081", nil)
}
