package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultSettingsPath       = "/var/lib/nonraid-ui/settings.json"
	DefaultAppearance         = "auto"
	DefaultWarningPct         = 95
	DefaultCriticalPct        = 98
	DefaultTempWarningC       = 45
	DefaultTempCriticalC      = 55
	DefaultNotifyFrequency    = "once"
)

// Settings represents the application configuration
type Settings struct {
	Appearance    string        `json:"appearance"` // "light", "dark", or "auto"
	Thresholds    Thresholds    `json:"thresholds"`
	Notifications Notifications `json:"notifications"`
}

// Thresholds defines disk usage and temperature warning levels
type Thresholds struct {
	WarningPct      int `json:"warning_pct"`       // Disk usage warning threshold percentage (e.g., 95)
	CriticalPct     int `json:"critical_pct"`      // Disk usage critical threshold percentage (e.g., 98)
	TempWarningC    int `json:"temp_warning_c"`    // Temperature warning threshold in Celsius (e.g., 45)
	TempCriticalC   int `json:"temp_critical_c"`   // Temperature critical threshold in Celsius (e.g., 55)
}

// Notifications defines notification settings
type Notifications struct {
	Enabled          bool              `json:"enabled"`
	DefaultFrequency string            `json:"default_frequency"` // "once", "15m", "30m", "1h", "3h", "6h", "12h", "24h"
	Events           NotificationEvent `json:"events"`
	Email            EmailConfig       `json:"email"`
	Discord          DiscordConfig     `json:"discord"`
}

// NotificationEventConfig defines configuration for a single notification event
type NotificationEventConfig struct {
	Enabled   bool   `json:"enabled"`
	Frequency string `json:"frequency"` // "default", "once", "15m", "30m", "1h", "3h", "6h", "12h", "24h"
}

// NotificationEvent defines which events trigger notifications
type NotificationEvent struct {
	ArrayHealth      NotificationEventConfig `json:"array_health"`
	DiskWarning      NotificationEventConfig `json:"disk_warning"`
	DiskCritical     NotificationEventConfig `json:"disk_critical"`
	DiskTempWarning  NotificationEventConfig `json:"disk_temp_warning"`
	DiskTempCritical NotificationEventConfig `json:"disk_temp_critical"`
	DiskStatusNotOK  NotificationEventConfig `json:"disk_status_not_ok"`
}

// EmailConfig defines email notification settings
type EmailConfig struct {
	Enabled     bool   `json:"enabled"`
	SMTPServer  string `json:"smtp_server"`
	SMTPPort    int    `json:"smtp_port"`
	Username    string `json:"username"`
	Password    string `json:"password"` // Plain text password
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
}

// DiscordConfig defines Discord webhook notification settings
type DiscordConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
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
			WarningPct:    DefaultWarningPct,
			CriticalPct:   DefaultCriticalPct,
			TempWarningC:  DefaultTempWarningC,
			TempCriticalC: DefaultTempCriticalC,
		},
		Notifications: Notifications{
			Enabled:          false,
			DefaultFrequency: DefaultNotifyFrequency,
			Events: NotificationEvent{
				ArrayHealth:      NotificationEventConfig{Enabled: false, Frequency: "default"},
				DiskWarning:      NotificationEventConfig{Enabled: false, Frequency: "default"},
				DiskCritical:     NotificationEventConfig{Enabled: false, Frequency: "default"},
				DiskTempWarning:  NotificationEventConfig{Enabled: false, Frequency: "default"},
				DiskTempCritical: NotificationEventConfig{Enabled: false, Frequency: "default"},
				DiskStatusNotOK:  NotificationEventConfig{Enabled: false, Frequency: "default"},
			},
			Email: EmailConfig{
				Enabled:  false,
				SMTPPort: 587,
			},
			Discord: DiscordConfig{
				Enabled: false,
			},
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

	// Validate temperature thresholds (reasonable range: 1-100°C, 0 is invalid)
	if m.settings.Thresholds.TempWarningC <= 0 || m.settings.Thresholds.TempWarningC > 100 {
		m.settings.Thresholds.TempWarningC = DefaultTempWarningC
	}
	if m.settings.Thresholds.TempCriticalC <= 0 || m.settings.Thresholds.TempCriticalC > 100 {
		m.settings.Thresholds.TempCriticalC = DefaultTempCriticalC
	}

	// Ensure temp warning < temp critical
	if m.settings.Thresholds.TempWarningC >= m.settings.Thresholds.TempCriticalC {
		m.settings.Thresholds.TempWarningC = DefaultTempWarningC
		m.settings.Thresholds.TempCriticalC = DefaultTempCriticalC
	}

	// Validate notification frequencies
	validDefaultFrequencies := map[string]bool{
		"once": true, "15m": true, "30m": true, "1h": true,
		"3h": true, "6h": true, "12h": true, "24h": true,
	}

	// Validate default frequency (cannot be "default")
	if !validDefaultFrequencies[m.settings.Notifications.DefaultFrequency] {
		m.settings.Notifications.DefaultFrequency = DefaultNotifyFrequency
	}

	// Validate and migrate event configurations
	m.settings.Notifications.Events.ArrayHealth = m.validateEventConfig(m.settings.Notifications.Events.ArrayHealth)
	m.settings.Notifications.Events.DiskWarning = m.validateEventConfig(m.settings.Notifications.Events.DiskWarning)
	m.settings.Notifications.Events.DiskCritical = m.validateEventConfig(m.settings.Notifications.Events.DiskCritical)
	m.settings.Notifications.Events.DiskTempWarning = m.validateEventConfig(m.settings.Notifications.Events.DiskTempWarning)
	m.settings.Notifications.Events.DiskTempCritical = m.validateEventConfig(m.settings.Notifications.Events.DiskTempCritical)
	m.settings.Notifications.Events.DiskStatusNotOK = m.validateEventConfig(m.settings.Notifications.Events.DiskStatusNotOK)

	// Validate email port
	if m.settings.Notifications.Email.SMTPPort < 1 || m.settings.Notifications.Email.SMTPPort > 65535 {
		m.settings.Notifications.Email.SMTPPort = 587
	}
}

// validateEventConfig validates and ensures a notification event config has valid values
func (m *Manager) validateEventConfig(config NotificationEventConfig) NotificationEventConfig {
	// Validate frequency
	validFrequencies := map[string]bool{
		"default": true, "once": true, "15m": true, "30m": true, "1h": true,
		"3h": true, "6h": true, "12h": true, "24h": true,
	}
	if !validFrequencies[config.Frequency] {
		config.Frequency = "default"
	}
	return config
}
