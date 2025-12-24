package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/jrarseneau/nonraid-ui/internal/api"
	"github.com/jrarseneau/nonraid-ui/internal/nmdctl"
)

func main() {
	port := flag.String("port", "3000", "Port to listen on")
	host := flag.String("host", "0.0.0.0", "Host to bind to")
	flag.Parse()

	log.Println("Starting nonraid-ui server...")

	// Create nmdctl client
	client := nmdctl.NewClient()

	// Test nmdctl access
	if _, err := client.GetStatus(); err != nil {
		log.Printf("Warning: Failed to get initial status from nmdctl: %v", err)
		log.Printf("Make sure nmdctl is installed and you have sufficient privileges")
	}

	// Create API server
	server := api.NewServer(client)

	// Configure HTTP server
	addr := *host + ":" + *port
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server listening on http://%s", addr)
	log.Printf("Access the UI at http://%s", addr)

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
