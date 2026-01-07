package notifications

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	DefaultStatePath = "/var/lib/nonraid-ui/notification-state.json"
)

// EventKey uniquely identifies a notification event
type EventKey struct {
	Type     string `json:"type"`      // "array_health", "disk_warning", "disk_critical", "disk_status"
	Resource string `json:"resource"`  // e.g., disk ID or "array"
}

// EventState tracks when a notification was last sent
type EventState struct {
	LastSent time.Time `json:"last_sent"`
	Count    int       `json:"count"` // Number of times sent
}

// State manages notification state persistence
type State struct {
	filePath string
	mu       sync.RWMutex
	events   map[string]EventState // Key is JSON-encoded EventKey
}

// NewState creates a new notification state manager
func NewState(filePath string) *State {
	if filePath == "" {
		filePath = DefaultStatePath
	}
	return &State{
		filePath: filePath,
		events:   make(map[string]EventState),
	}
}

// Load reads state from disk
func (s *State) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Check if file exists
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty state
			return s.saveUnsafe()
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse existing state
	if err := json.Unmarshal(data, &s.events); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	return nil
}

// Save writes current state to disk
func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveUnsafe()
}

// saveUnsafe writes state without acquiring lock (must be called with lock held)
func (s *State) saveUnsafe() error {
	data, err := json.MarshalIndent(s.events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write atomically using temp file + rename
	tempPath := s.filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	if err := os.Rename(tempPath, s.filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// ShouldNotify checks if a notification should be sent based on frequency
func (s *State) ShouldNotify(eventType, resource, frequency string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(eventType, resource)
	state, exists := s.events[key]

	// If never sent, should notify
	if !exists {
		return true
	}

	// If frequency is "once", never send again (unless state is cleared on restart)
	if frequency == "once" {
		return false
	}

	// Check if enough time has passed
	duration := s.frequencyToDuration(frequency)
	return time.Since(state.LastSent) >= duration
}

// MarkSent records that a notification was sent
func (s *State) MarkSent(eventType, resource string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(eventType, resource)
	state := s.events[key]
	state.LastSent = time.Now()
	state.Count++
	s.events[key] = state

	return s.saveUnsafe()
}

// Clear removes a specific event from state (e.g., when condition is resolved)
func (s *State) Clear(eventType, resource string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(eventType, resource)
	delete(s.events, key)

	return s.saveUnsafe()
}

// ClearAll removes all events from state
func (s *State) ClearAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = make(map[string]EventState)
	return s.saveUnsafe()
}

// makeKey creates a unique key for an event
func (s *State) makeKey(eventType, resource string) string {
	return fmt.Sprintf("%s:%s", eventType, resource)
}

// frequencyToDuration converts frequency string to time.Duration
func (s *State) frequencyToDuration(frequency string) time.Duration {
	switch frequency {
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "3h":
		return 3 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "24h":
		return 24 * time.Hour
	default:
		return 0 // "once" or invalid
	}
}
