package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

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
	api.HandleFunc("/disk/{slot}", s.handleDiskDetails).Methods("GET")
	api.HandleFunc("/disks/{diskid}/note", s.handleUpdateDiskNote).Methods("PUT")
	api.HandleFunc("/disks/{diskid}/note", s.handleDeleteDiskNote).Methods("DELETE")
	api.HandleFunc("/settings", s.handleGetSettings).Methods("GET")
	api.HandleFunc("/settings", s.handleUpdateSettings).Methods("PUT")
	api.HandleFunc("/notifications/test/email", s.handleTestEmail).Methods("POST")
	api.HandleFunc("/notifications/test/discord", s.handleTestDiscord).Methods("POST")

	// Serve embedded frontend with SPA fallback
	frontendSubFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Printf("Warning: frontend assets not embedded, serving will fail: %v", err)
	} else {
		s.router.PathPrefix("/").HandlerFunc(s.handleSPA(frontendSubFS))
	}
}

// handleSPA serves the SPA with fallback to index.html for client-side routing
func (s *Server) handleSPA(fsys fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(fsys))

	return func(w http.ResponseWriter, r *http.Request) {
		// Try to open the requested file
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		file, err := fsys.Open(path[1:]) // Remove leading slash
		if err == nil {
			file.Close()
			// File exists, serve it normally
			fileServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist, serve index.html for client-side routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.client.GetStatus()
	if err != nil {
		log.Printf("Error getting status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get current settings for disk notes
	currentSettings := s.settings.Get()

	// Merge SMART temperature data and notes into disk info
	for i := range status.Disks {
		if status.Disks[i].Device != "" {
			temp := s.smartCache.GetTemperature(status.Disks[i].Device)
			status.Disks[i].Temperature = temp
		}

		// Add note if exists
		if diskNote, exists := currentSettings.DiskNotes[status.Disks[i].DiskID]; exists {
			status.Disks[i].Note = &diskNote.Note
			status.Disks[i].NoteUpdatedAt = &diskNote.UpdatedAt
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// DiskDetails represents the combined disk information
type DiskDetails struct {
	Disk      nmdctl.Disk          `json:"disk"`
	SmartData *smartctl.FullSmartData `json:"smart_data,omitempty"`
	AllDisks  []DiskNavigationInfo `json:"all_disks"` // For navigation
}

// DiskNavigationInfo contains minimal info for navigation
type DiskNavigationInfo struct {
	Slot int    `json:"slot"`
	Type string `json:"type"`
}

func (s *Server) handleDiskDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slotStr := vars["slot"]

	slot, err := strconv.Atoi(slotStr)
	if err != nil {
		http.Error(w, "Invalid slot number", http.StatusBadRequest)
		return
	}

	// Get current status from nmdctl
	status, err := s.client.GetStatus()
	if err != nil {
		log.Printf("Error getting status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the disk with the matching slot
	var targetDisk *nmdctl.Disk
	for i := range status.Disks {
		if status.Disks[i].Slot == slot {
			targetDisk = &status.Disks[i]
			break
		}
	}

	if targetDisk == nil {
		http.Error(w, "Disk not found", http.StatusNotFound)
		return
	}

	// Add temperature from cache
	if targetDisk.Device != "" {
		temp := s.smartCache.GetTemperature(targetDisk.Device)
		targetDisk.Temperature = temp
	}

	// Add note if exists
	currentSettings := s.settings.Get()
	if diskNote, exists := currentSettings.DiskNotes[targetDisk.DiskID]; exists {
		targetDisk.Note = &diskNote.Note
		targetDisk.NoteUpdatedAt = &diskNote.UpdatedAt
	}

	// Get full SMART data
	var smartData *smartctl.FullSmartData
	if targetDisk.Device != "" {
		smartClient := smartctl.NewClient()
		smartData, err = smartClient.GetFullData(targetDisk.Device)
		if err != nil {
			log.Printf("Warning: Failed to get SMART data for %s: %v", targetDisk.Device, err)
			// Don't fail the request, just continue without SMART data
		}
	}

	// Build navigation info (all disks sorted: P, Q, then 1-N)
	allDisksNav := make([]DiskNavigationInfo, len(status.Disks))
	for i, disk := range status.Disks {
		allDisksNav[i] = DiskNavigationInfo{
			Slot: disk.Slot,
			Type: disk.Type,
		}
	}

	// Sort disks: P first, Q second, then numerical order
	sort.Slice(allDisksNav, func(i, j int) bool {
		// P comes first
		if allDisksNav[i].Type == "P" {
			return true
		}
		if allDisksNav[j].Type == "P" {
			return false
		}
		// Q comes second
		if allDisksNav[i].Type == "Q" {
			return true
		}
		if allDisksNav[j].Type == "Q" {
			return false
		}
		// Both are data disks, sort by slot number
		return allDisksNav[i].Slot < allDisksNav[j].Slot
	})

	response := DiskDetails{
		Disk:      *targetDisk,
		SmartData: smartData,
		AllDisks:  allDisksNav,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

func (s *Server) handleUpdateDiskNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	diskID := vars["diskid"]

	if diskID == "" {
		http.Error(w, "Disk ID is required", http.StatusBadRequest)
		return
	}

	// Limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	// Parse request body
	var req struct {
		Note string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Trim whitespace
	note := strings.TrimSpace(req.Note)

	// Validate note length
	if len(note) > 500 {
		http.Error(w, "Note too long (max 500 characters)", http.StatusBadRequest)
		return
	}

	// Get current settings
	currentSettings := s.settings.Get()

	// If note is empty, delete it
	if note == "" {
		delete(currentSettings.DiskNotes, diskID)
	} else {
		// Update or create note
		currentSettings.DiskNotes[diskID] = settings.DiskNote{
			Note:      note,
			UpdatedAt: time.Now(),
		}
	}

	// Save settings
	if err := s.settings.Update(currentSettings); err != nil {
		log.Printf("Error saving disk note: %v", err)
		http.Error(w, "Failed to save note", http.StatusInternalServerError)
		return
	}

	// Return the updated note
	response := map[string]interface{}{
		"disk_id": diskID,
	}
	if diskNote, exists := currentSettings.DiskNotes[diskID]; exists {
		response["note"] = diskNote.Note
		response["updated_at"] = diskNote.UpdatedAt
	} else {
		response["note"] = ""
		response["updated_at"] = nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleDeleteDiskNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	diskID := vars["diskid"]

	if diskID == "" {
		http.Error(w, "Disk ID is required", http.StatusBadRequest)
		return
	}

	// Get current settings
	currentSettings := s.settings.Get()

	// Delete the note
	delete(currentSettings.DiskNotes, diskID)

	// Save settings
	if err := s.settings.Update(currentSettings); err != nil {
		log.Printf("Error deleting disk note: %v", err)
		http.Error(w, "Failed to delete note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
