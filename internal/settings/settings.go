package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultSettingsPath = "/var/lib/nonraid-ui/settings.json"
	DefaultAppearance   = "auto"
	DefaultWarningPct   = 95
	DefaultCriticalPct  = 98
)

// Settings represents the application configuration
type Settings struct {
	Appearance string     `json:"appearance"` // "light", "dark", or "auto"
	Thresholds Thresholds `json:"thresholds"`
}

// Thresholds defines disk usage warning levels
type Thresholds struct {
	WarningPct  int `json:"warning_pct"`  // Warning threshold percentage (e.g., 95)
	CriticalPct int `json:"critical_pct"` // Critical threshold percentage (e.g., 98)
}

// Manager handles settings persistence
type Manager struct {
	filePath string
	mu       sync.RWMutex
	settings Settings
}

// NewManager creates a new settings manager
func NewManager(filePath string) *Manager {
	if filePath == "" {
		filePath = DefaultSettingsPath
	}
	return &Manager{
		filePath: filePath,
		settings: GetDefaults(),
	}
}

// GetDefaults returns default settings
func GetDefaults() Settings {
	return Settings{
		Appearance: DefaultAppearance,
		Thresholds: Thresholds{
			WarningPct:  DefaultWarningPct,
			CriticalPct: DefaultCriticalPct,
		},
	}
}

// Load reads settings from disk, creating defaults if file doesn't exist
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Check if file exists
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults and save them
			m.settings = GetDefaults()
			return m.saveUnsafe()
		}
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Parse existing settings
	if err := json.Unmarshal(data, &m.settings); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Validate settings
	m.validateAndFixUnsafe()

	return nil
}

// Save writes current settings to disk
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveUnsafe()
}

// saveUnsafe writes settings without acquiring lock (must be called with lock held)
func (m *Manager) saveUnsafe() error {
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write atomically using temp file + rename
	tempPath := m.filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	if err := os.Rename(tempPath, m.filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename settings file: %w", err)
	}

	return nil
}

// Get returns a copy of current settings
func (m *Manager) Get() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

// Update updates settings and saves to disk
func (m *Manager) Update(newSettings Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings = newSettings
	m.validateAndFixUnsafe()
	return m.saveUnsafe()
}

// validateAndFixUnsafe ensures settings are valid, fixing invalid values
// Must be called with lock held
func (m *Manager) validateAndFixUnsafe() {
	// Validate appearance
	if m.settings.Appearance != "light" && m.settings.Appearance != "dark" && m.settings.Appearance != "auto" {
		m.settings.Appearance = DefaultAppearance
	}

	// Validate thresholds
	if m.settings.Thresholds.WarningPct < 0 || m.settings.Thresholds.WarningPct > 100 {
		m.settings.Thresholds.WarningPct = DefaultWarningPct
	}
	if m.settings.Thresholds.CriticalPct < 0 || m.settings.Thresholds.CriticalPct > 100 {
		m.settings.Thresholds.CriticalPct = DefaultCriticalPct
	}

	// Ensure warning < critical
	if m.settings.Thresholds.WarningPct >= m.settings.Thresholds.CriticalPct {
		m.settings.Thresholds.WarningPct = DefaultWarningPct
		m.settings.Thresholds.CriticalPct = DefaultCriticalPct
	}
}
