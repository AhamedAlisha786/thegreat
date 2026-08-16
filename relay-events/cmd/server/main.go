// Command server runs the Relay campaign events HTTP service.
package main

import (
	"log"
	"net/http"

	"relay-events/internal/events"
)

func main() {
	srv := events.NewServer()

	addr := ":8080"
	log.Printf("relay-events listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, srv.Routes()))
}
