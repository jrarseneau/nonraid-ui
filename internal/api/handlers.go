package api

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jrarseneau/nonraid-ui/internal/nmdctl"
	"github.com/jrarseneau/nonraid-ui/internal/settings"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

// Server represents the API server
type Server struct {
	client   *nmdctl.Client
	settings *settings.Manager
	router   *mux.Router
}

// NewServer creates a new API server
func NewServer(client *nmdctl.Client, settingsMgr *settings.Manager) *Server {
	s := &Server{
		client:   client,
		settings: settingsMgr,
		router:   mux.NewRouter(),
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
