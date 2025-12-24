package notifications

import (
	"fmt"
	"log"
	"time"

	"github.com/jrarseneau/nonraid-ui/internal/nmdctl"
	"github.com/jrarseneau/nonraid-ui/internal/settings"
)

// Manager handles notification polling and dispatch
type Manager struct {
	client   *nmdctl.Client
	settings *settings.Manager
	state    *State
	stopChan chan struct{}
}

// NewManager creates a new notification manager
func NewManager(client *nmdctl.Client, settingsMgr *settings.Manager) *Manager {
	return &Manager{
		client:   client,
		settings: settingsMgr,
		state:    NewState(""),
		stopChan: make(chan struct{}),
	}
}

// Start begins the notification polling loop
func (m *Manager) Start() error {
	// Load state
	if err := m.state.Load(); err != nil {
		return fmt.Errorf("failed to load notification state: %w", err)
	}

	log.Println("Starting notification manager...")

	// Start polling in background
	go m.pollLoop()

	return nil
}

// Stop stops the notification manager
func (m *Manager) Stop() {
	close(m.stopChan)
	log.Println("Notification manager stopped")
}

// pollLoop polls nmdctl status and checks for events
func (m *Manager) pollLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run immediately on start
	m.checkAndNotify()

	for {
		select {
		case <-ticker.C:
			m.checkAndNotify()
		case <-m.stopChan:
			return
		}
	}
}

// checkAndNotify checks for events and sends notifications
func (m *Manager) checkAndNotify() {
	// Get current settings
	cfg := m.settings.Get()

	// Skip if notifications disabled
	if !cfg.Notifications.Enabled {
		return
	}

	// Get current status
	status, err := m.client.GetStatus()
	if err != nil {
		log.Printf("Failed to get status for notifications: %v", err)
		return
	}

	// Check for events
	m.checkArrayHealth(status, cfg)
	m.checkDiskUsage(status, cfg)
	m.checkDiskStatus(status, cfg)
}

// checkArrayHealth checks if array is healthy
func (m *Manager) checkArrayHealth(status *nmdctl.Status, cfg settings.Settings) {
	if !cfg.Notifications.Events.ArrayHealth {
		return
	}

	// Check if array is not healthy
	if status.Array.State != "healthy" {
		eventType := "array_health"
		resource := "array"

		// Check if we should notify
		if !m.state.ShouldNotify(eventType, resource, cfg.Notifications.Frequency) {
			return
		}

		// Send notification
		title := "⚠️ Array Health Alert"
		description := fmt.Sprintf("Array state is **%s** (not healthy)", status.Array.State)
		m.sendNotification(title, description, ColorYellow, cfg)

		// Mark as sent
		m.state.MarkSent(eventType, resource)
	} else {
		// Array is healthy, clear state so we can notify again if it becomes unhealthy
		m.state.Clear("array_health", "array")
	}
}

// checkDiskUsage checks disk usage against thresholds
func (m *Manager) checkDiskUsage(status *nmdctl.Status, cfg settings.Settings) {
	for _, disk := range status.Disks {
		// Skip parity disks (they don't have usage data)
		if disk.Type == "P" || disk.Type == "Q" {
			continue
		}

		// Skip if no filesystem data
		if disk.Filesystem == nil || disk.Filesystem.Usage == "" {
			continue
		}

		// Parse usage percentage
		var usagePct int
		fmt.Sscanf(disk.Filesystem.Usage, "%d%%", &usagePct)

		// Check critical threshold
		if cfg.Notifications.Events.DiskCritical && usagePct >= cfg.Thresholds.CriticalPct {
			eventType := "disk_critical"
			resource := disk.DiskID

			if m.state.ShouldNotify(eventType, resource, cfg.Notifications.Frequency) {
				title := "🚨 Disk Critical Alert"
				description := fmt.Sprintf("Disk **%s** (slot %d) has reached **%d%%** usage (critical threshold: %d%%)",
					disk.DiskID, disk.Slot, usagePct, cfg.Thresholds.CriticalPct)
				m.sendNotification(title, description, ColorRed, cfg)
				m.state.MarkSent(eventType, resource)
			}
		} else if cfg.Notifications.Events.DiskWarning && usagePct >= cfg.Thresholds.WarningPct {
			// Check warning threshold (only if not critical)
			eventType := "disk_warning"
			resource := disk.DiskID

			if m.state.ShouldNotify(eventType, resource, cfg.Notifications.Frequency) {
				title := "⚠️ Disk Warning Alert"
				description := fmt.Sprintf("Disk **%s** (slot %d) has reached **%d%%** usage (warning threshold: %d%%)",
					disk.DiskID, disk.Slot, usagePct, cfg.Thresholds.WarningPct)
				m.sendNotification(title, description, ColorYellow, cfg)
				m.state.MarkSent(eventType, resource)
			}
		} else {
			// Usage is below thresholds, clear states
			m.state.Clear("disk_critical", disk.DiskID)
			m.state.Clear("disk_warning", disk.DiskID)
		}
	}
}

// checkDiskStatus checks if any disk status is not OK
func (m *Manager) checkDiskStatus(status *nmdctl.Status, cfg settings.Settings) {
	if !cfg.Notifications.Events.DiskStatusNotOK {
		return
	}

	for _, disk := range status.Disks {
		if disk.Status != "OK" {
			eventType := "disk_status"
			resource := disk.DiskID

			if m.state.ShouldNotify(eventType, resource, cfg.Notifications.Frequency) {
				title := "⚠️ Disk Status Alert"
				description := fmt.Sprintf("Disk **%s** (slot %d) status is **%s** (not OK)",
					disk.DiskID, disk.Slot, disk.Status)
				m.sendNotification(title, description, ColorRed, cfg)
				m.state.MarkSent(eventType, resource)
			}
		} else {
			// Disk is OK, clear state
			m.state.Clear("disk_status", disk.DiskID)
		}
	}
}

// sendNotification sends a notification via all enabled channels
func (m *Manager) sendNotification(title, description string, color int, cfg settings.Settings) {
	// Send via email if enabled
	if cfg.Notifications.Email.Enabled {
		emailService := NewEmailService(
			cfg.Notifications.Email.SMTPServer,
			cfg.Notifications.Email.SMTPPort,
			cfg.Notifications.Email.Username,
			cfg.Notifications.Email.PasswordMD5,
			cfg.Notifications.Email.FromAddress,
		)

		// Build HTML email body
		body := fmt.Sprintf(`
<html>
<body>
	<h2>%s</h2>
	<p>%s</p>
	<hr>
	<p><small>Sent from nonraidUI at %s</small></p>
</body>
</html>
`, title, description, time.Now().Format("2006-01-02 15:04:05"))

		// Note: This is a simplified version - in production you'd need to handle password properly
		err := emailService.Send(cfg.Notifications.Email.ToAddress, title, body)
		if err != nil {
			log.Printf("Failed to send email notification: %v", err)
		} else {
			log.Printf("Sent email notification: %s", title)
		}
	}

	// Send via Discord if enabled
	if cfg.Notifications.Discord.Enabled {
		discordService := NewDiscordService(cfg.Notifications.Discord.WebhookURL)

		err := discordService.Send(title, description, color, nil)
		if err != nil {
			log.Printf("Failed to send Discord notification: %v", err)
		} else {
			log.Printf("Sent Discord notification: %s", title)
		}
	}
}

// TestEmail sends a test email notification
func (m *Manager) TestEmail(cfg settings.EmailConfig, password string) error {
	service := NewEmailService(
		cfg.SMTPServer,
		cfg.SMTPPort,
		cfg.Username,
		cfg.PasswordMD5,
		cfg.FromAddress,
	)

	return service.TestConnection(cfg.ToAddress, password)
}

// TestDiscord sends a test Discord notification
func (m *Manager) TestDiscord(webhookURL string) error {
	service := NewDiscordService(webhookURL)
	return service.TestConnection()
}
