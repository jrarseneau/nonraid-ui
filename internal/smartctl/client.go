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

// SmartData represents the parsed smartctl output
type SmartData struct {
	Temperature *Temperature `json:"temperature"`
}

// Temperature contains temperature readings
type Temperature struct {
	Current int `json:"current"` // Temperature in Celsius
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
