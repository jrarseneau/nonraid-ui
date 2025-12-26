package smartctl

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Client wraps smartctl command execution
type Client struct {
	command string
	timeout time.Duration
}

// NewClient creates a new smartctl client
func NewClient() *Client {
	return &Client{
		command: "smartctl",
		timeout: 5 * time.Second, // 5 second timeout per disk
	}
}

// SmartData represents the parsed smartctl output (minimal for temperature)
type SmartData struct {
	Temperature *Temperature `json:"temperature"`
}

// Temperature contains temperature readings
type Temperature struct {
	Current int `json:"current"` // Temperature in Celsius
}

// FullSmartData represents the complete smartctl JSON output
type FullSmartData struct {
	Device         DeviceInfo       `json:"device"`
	ModelFamily    string           `json:"model_family,omitempty"`
	ModelName      string           `json:"model_name,omitempty"`
	SerialNumber   string           `json:"serial_number,omitempty"`
	FirmwareVersion string          `json:"firmware_version,omitempty"`
	RotationRate   int              `json:"rotation_rate,omitempty"` // RPM, 0 for SSD
	FormFactor     *FormFactor      `json:"form_factor,omitempty"`
	SmartStatus    *SmartStatus     `json:"smart_status,omitempty"`
	Temperature    *Temperature     `json:"temperature,omitempty"`
	PowerOnTime    *PowerOnTime     `json:"power_on_time,omitempty"`
	PowerCycleCount int             `json:"power_cycle_count,omitempty"`
	AtaSmartAttributes *AtaSmartAttributes `json:"ata_smart_attributes,omitempty"`
	SmartctlInfo   SmartctlInfo     `json:"smartctl"`
}

// DeviceInfo contains device information
type DeviceInfo struct {
	Name     string `json:"name"`
	InfoName string `json:"info_name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
}

// SmartStatus contains the overall SMART health status
type SmartStatus struct {
	Passed bool `json:"passed"`
}

// PowerOnTime contains power-on hours
type PowerOnTime struct {
	Hours int `json:"hours"`
}

// FormFactor contains form factor information
type FormFactor struct {
	AtaValue int    `json:"ata_value"`
	Name     string `json:"name"`
}

// AtaSmartAttributes contains the table of SMART attributes
type AtaSmartAttributes struct {
	Table []SmartAttribute `json:"table"`
}

// SmartAttribute represents a single SMART attribute
type SmartAttribute struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Value      int    `json:"value"`
	Worst      int    `json:"worst"`
	Thresh     int    `json:"thresh"`
	WhenFailed string `json:"when_failed"`
	Flags      struct {
		Value         int    `json:"value"`
		String        string `json:"string"`
		Prefailure    bool   `json:"prefailure"`
		UpdatedOnline bool   `json:"updated_online"`
		Performance   bool   `json:"performance"`
		ErrorRate     bool   `json:"error_rate"`
		EventCount    bool   `json:"event_count"`
		AutoKeep      bool   `json:"auto_keep"`
	} `json:"flags"`
	Raw struct {
		Value  int64  `json:"value"`
		String string `json:"string"`
	} `json:"raw"`
}

// SmartctlInfo contains smartctl version and execution info
type SmartctlInfo struct {
	Version    []int    `json:"version"`
	ExitStatus int      `json:"exit_status"`
	Messages   []Message `json:"messages,omitempty"`
}

// Message represents a smartctl message
type Message struct {
	String   string `json:"string"`
	Severity string `json:"severity"`
}

// normalizeDevicePath ensures the device path has /dev/ prefix
// smartctl works fine with partition paths, so we just add the prefix if missing
// Examples: sdp1 -> /dev/sdp1, /dev/sdp1 -> /dev/sdp1
func normalizeDevicePath(device string) string {
	// If it already has /dev/ prefix, return as-is
	if strings.HasPrefix(device, "/dev/") {
		return device
	}
	// Add /dev/ prefix
	return "/dev/" + device
}

// GetTemperature fetches the current temperature for a device
// Returns nil if temperature data is not available
func (c *Client) GetTemperature(device string) (*int, error) {
	if device == "" {
		return nil, fmt.Errorf("device cannot be empty")
	}

	// Normalize device path (strip partition numbers, ensure /dev/ prefix)
	normalizedDevice := normalizeDevicePath(device)

	// Execute smartctl with JSON output
	// -A shows all SMART attributes
	// -j outputs JSON format
	// --nocheck=standby prevents waking up sleeping drives
	cmd := exec.Command(c.command, "-A", "-j", "--nocheck=standby", normalizedDevice)

	// Create a channel for timeout
	done := make(chan error, 1)
	var output []byte
	var err error

	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	// Wait for command to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			// Check if the error is because drive is in standby
			// In that case, we just return nil (no temperature available)
			if len(output) > 0 {
				// Try to parse anyway - sometimes smartctl returns data even with non-zero exit
				var data SmartData
				if jsonErr := json.Unmarshal(output, &data); jsonErr == nil {
					if data.Temperature != nil && data.Temperature.Current > 0 {
						temp := data.Temperature.Current
						return &temp, nil
					}
				}
			}
			return nil, fmt.Errorf("smartctl command failed for %s: %w (output: %s)", normalizedDevice, err, string(output))
		}
	case <-time.After(c.timeout):
		// Kill the process if it's still running
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("smartctl timeout for device %s", normalizedDevice)
	}

	// Parse the JSON output
	var data SmartData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse smartctl output for %s: %w", normalizedDevice, err)
	}

	// Extract temperature
	if data.Temperature == nil || data.Temperature.Current <= 0 {
		// No valid temperature data
		return nil, nil
	}

	temp := data.Temperature.Current
	return &temp, nil
}

// GetFullData fetches complete SMART data for a device
// Returns the full smartctl JSON output including attributes, health status, etc.
func (c *Client) GetFullData(device string) (*FullSmartData, error) {
	if device == "" {
		return nil, fmt.Errorf("device cannot be empty")
	}

	// Normalize device path
	normalizedDevice := normalizeDevicePath(device)

	// Execute smartctl with full data output
	// -a shows all SMART information
	// -j outputs JSON format
	cmd := exec.Command(c.command, "-a", "-j", normalizedDevice)

	// Create a channel for timeout
	done := make(chan error, 1)
	var output []byte
	var err error

	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	// Wait for command to complete or timeout
	select {
	case <-done:
		// smartctl can return non-zero exit codes even with valid data
		// (e.g., if there are previous errors logged), so we try to parse anyway
		if len(output) == 0 {
			return nil, fmt.Errorf("smartctl returned no output for %s", normalizedDevice)
		}
	case <-time.After(c.timeout):
		// Kill the process if it's still running
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("smartctl timeout for device %s", normalizedDevice)
	}

	// Parse the JSON output
	var data FullSmartData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse smartctl output for %s: %w (output: %s)", normalizedDevice, err, string(output))
	}

	return &data, nil
}
