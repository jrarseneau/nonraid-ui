package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jrarseneau/nonraid-ui/internal/nmdctl"
	"github.com/jrarseneau/nonraid-ui/internal/notifications"
	"github.com/jrarseneau/nonraid-ui/internal/settings"
	"github.com/jrarseneau/nonraid-ui/internal/smartctl"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

// Server represents the API server
type Server struct {
	client        *nmdctl.Client
	settings      *settings.Manager
	notifications *notifications.Manager
	smartCache    *smartctl.Cache
	router        *mux.Router
}

// NewServer creates a new API server
func NewServer(client *nmdctl.Client, settingsMgr *settings.Manager, notifMgr *notifications.Manager, smartCache *smartctl.Cache) *Server {
	s := &Server{
		client:        client,
		settings:      settingsMgr,
		notifications: notifMgr,
		smartCache:    smartCache,
		router:        mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/status", s.handleStatus).Methods("GET")
	api.HandleFunc("/settings", s.handleGetSettings).Methods("GET")
	api.HandleFunc("/settings", s.handleUpdateSettings).Methods("PUT")
	api.HandleFunc("/notifications/test/email", s.handleTestEmail).Methods("POST")
	api.HandleFunc("/notifications/test/discord", s.handleTestDiscord).Methods("POST")

	// Serve embedded frontend
	frontendSubFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Printf("Warning: frontend assets not embedded, serving will fail: %v", err)
	} else {
		s.router.PathPrefix("/").Handler(http.FileServer(http.FS(frontendSubFS)))
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.client.GetStatus()
	if err != nil {
		log.Printf("Error getting status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Merge SMART temperature data into disk info
	for i := range status.Disks {
		if status.Disks[i].Device != "" {
			temp := s.smartCache.GetTemperature(status.Disks[i].Device)
			status.Disks[i].Temperature = temp
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Router returns the configured router
func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	currentSettings := s.settings.Get()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(currentSettings); err != nil {
		log.Printf("Error encoding settings: %v", err)
		http.Error(w, "Failed to encode settings", http.StatusInternalServerError)
	}
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var newSettings settings.Settings
	if err := json.Unmarshal(body, &newSettings); err != nil {
		log.Printf("Error parsing settings: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Encode the email password if it's provided and not already encoded
	if newSettings.Notifications.Email.PasswordEncoded != "" {
		// Check if it's already base64 encoded by trying to decode it
		if _, err := notifications.DecodePassword(newSettings.Notifications.Email.PasswordEncoded); err != nil {
			// Not valid base64, so encode it
			newSettings.Notifications.Email.PasswordEncoded = notifications.EncodePassword(newSettings.Notifications.Email.PasswordEncoded)
		}
	}

	// Update and save settings
	if err := s.settings.Update(newSettings); err != nil {
		log.Printf("Error saving settings: %v", err)
		http.Error(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}

	// Return updated settings
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(newSettings); err != nil {
		log.Printf("Error encoding settings: %v", err)
		http.Error(w, "Failed to encode settings", http.StatusInternalServerError)
	}
}

func (s *Server) handleTestEmail(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	// Parse request body
	var req struct {
		Config   settings.EmailConfig `json:"config"`
		Password string                `json:"password"` // Plaintext password for testing
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Test email
	if err := s.notifications.TestEmail(req.Config, req.Password); err != nil {
		log.Printf("Email test failed: %v", err)
		http.Error(w, fmt.Sprintf("Email test failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test email sent successfully"})
}

func (s *Server) handleTestDiscord(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	// Parse request body
	var req struct {
		WebhookURL string `json:"webhook_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Test Discord
	if err := s.notifications.TestDiscord(req.WebhookURL); err != nil {
		log.Printf("Discord test failed: %v", err)
		http.Error(w, fmt.Sprintf("Discord test failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test Discord notification sent successfully"})
}
