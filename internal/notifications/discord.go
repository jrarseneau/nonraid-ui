package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DiscordService handles Discord webhook notifications
type DiscordService struct {
	webhookURL string
}

// NewDiscordService creates a new Discord notification service
func NewDiscordService(webhookURL string) *DiscordService {
	return &DiscordService{
		webhookURL: webhookURL,
	}
}

// DiscordMessage represents a Discord webhook message
type DiscordMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordEmbedFooter represents a footer in a Discord embed
type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

// Send sends a Discord notification
func (d *DiscordService) Send(title, description string, color int, fields []DiscordEmbedField) error {
	embed := DiscordEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &DiscordEmbedFooter{
			Text: "nonraidUI",
		},
	}

	message := DiscordMessage{
		Embeds: []DiscordEmbed{embed},
	}

	return d.sendMessage(message)
}

// SendSimple sends a simple text notification
func (d *DiscordService) SendSimple(content string) error {
	message := DiscordMessage{
		Content: content,
	}

	return d.sendMessage(message)
}

// sendMessage sends a Discord webhook message
func (d *DiscordService) sendMessage(message DiscordMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	resp, err := http.Post(d.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// TestConnection tests the Discord webhook
func (d *DiscordService) TestConnection() error {
	embed := DiscordEmbed{
		Title:       "🧪 Test Notification",
		Description: "This is a test notification from nonraidUI to verify your Discord webhook is configured correctly.",
		Color:       3447003, // Blue color
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &DiscordEmbedFooter{
			Text: "nonraidUI",
		},
	}

	message := DiscordMessage{
		Embeds: []DiscordEmbed{embed},
	}

	return d.sendMessage(message)
}

// Color constants for different notification types
const (
	ColorGreen  = 3066993  // Success/OK
	ColorYellow = 16776960 // Warning
	ColorRed    = 15158332 // Critical/Error
	ColorBlue   = 3447003  // Info
)
