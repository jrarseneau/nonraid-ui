package nmdctl

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// Client handles interactions with nmdctl
type Client struct {
	command string
}

// NewClient creates a new nmdctl client
func NewClient() *Client {
	return &Client{
		command: "nmdctl",
	}
}

// GetStatus executes nmdctl status -o json and returns the parsed result
// Note: nmdctl returns exit code 1 when array is degraded/unhealthy, but still outputs valid JSON
func (c *Client) GetStatus() (*Status, error) {
	cmd := exec.Command(c.command, "status", "-o", "json")

	// Use CombinedOutput to capture stdout even if exit code is non-zero
	output, err := cmd.CombinedOutput()

	// nmdctl returns exit code 1 when array is degraded/unhealthy, but JSON is still valid
	// Only fail if we got no output at all
	if err != nil && len(output) == 0 {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nmdctl failed: %s - stderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute nmdctl: %w", err)
	}

	var status Status
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse nmdctl output: %w (output: %s)", err, string(output))
	}

	return &status, nil
}
