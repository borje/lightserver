package main

import (
	"fmt"
	"net/http"
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	for i := eventQueue.Len(); i > 0; i-- {
		e := (*eventQueue)[i-1]
		fmt.Fprintf(w, "%d: %3s @ %s\n", e.device, e.action, e.time)
	}
}
