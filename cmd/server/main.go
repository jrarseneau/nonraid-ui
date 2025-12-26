package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/jrarseneau/nonraid-ui/internal/api"
	"github.com/jrarseneau/nonraid-ui/internal/nmdctl"
	"github.com/jrarseneau/nonraid-ui/internal/notifications"
	"github.com/jrarseneau/nonraid-ui/internal/settings"
	"github.com/jrarseneau/nonraid-ui/internal/smartctl"
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

	// Create and load settings
	settingsMgr := settings.NewManager("")
	if err := settingsMgr.Load(); err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}
	log.Printf("Settings loaded from: %s", settings.DefaultSettingsPath)

	// Create and start SMART cache (polls every 30 seconds)
	smartCache := smartctl.NewCache(client, 30*time.Second)
	smartCache.Start()
	defer smartCache.Stop()

	// Create and start notification manager (needs smartCache for temperature data)
	notifMgr := notifications.NewManager(client, settingsMgr, smartCache)
	if err := notifMgr.Start(); err != nil {
		log.Fatalf("Failed to start notification manager: %v", err)
	}
	defer notifMgr.Stop()

	// Create API server
	server := api.NewServer(client, settingsMgr, notifMgr, smartCache)

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
