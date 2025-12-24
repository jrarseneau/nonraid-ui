package api

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jrarseneau/nonraid-ui/internal/nmdctl"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

// Server represents the API server
type Server struct {
	client *nmdctl.Client
	router *mux.Router
}

// NewServer creates a new API server
func NewServer(client *nmdctl.Client) *Server {
	s := &Server{
		client: client,
		router: mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/status", s.handleStatus).Methods("GET")

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
