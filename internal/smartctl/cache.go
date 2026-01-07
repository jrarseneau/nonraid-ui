package smartctl

import (
	"log"
	"sync"
	"time"

	"github.com/jrarseneau/nonraid-ui/internal/nmdctl"
)

// Cache stores SMART temperature data for all disks
type Cache struct {
	mu           sync.RWMutex
	temperatures map[string]*int // device -> temperature in Celsius
	client       *Client
	nmdctlClient *nmdctl.Client
	pollInterval time.Duration
	stopCh       chan struct{}
}

// NewCache creates a new SMART data cache
func NewCache(nmdctlClient *nmdctl.Client, pollInterval time.Duration) *Cache {
	return &Cache{
		temperatures: make(map[string]*int),
		client:       NewClient(),
		nmdctlClient: nmdctlClient,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

// Start begins background polling of SMART data
func (c *Cache) Start() {
	// Initial poll
	c.pollAll()

	// Start background polling
	go func() {
		ticker := time.NewTicker(c.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.pollAll()
			case <-c.stopCh:
				return
			}
		}
	}()

	log.Printf("SMART cache started with %v polling interval", c.pollInterval)
}

// Stop stops the background polling
func (c *Cache) Stop() {
	close(c.stopCh)
}

// pollAll fetches SMART data for all disks concurrently
func (c *Cache) pollAll() {
	// Get current disk list from nmdctl
	status, err := c.nmdctlClient.GetStatus()
	if err != nil {
		log.Printf("Failed to get disk list for SMART polling: %v", err)
		return
	}

	// Create a wait group for concurrent polling
	var wg sync.WaitGroup
	tempCh := make(chan struct {
		device string
		temp   *int
	}, len(status.Disks))

	// Poll each disk concurrently
	for _, disk := range status.Disks {
		if disk.Device == "" {
			continue
		}

		wg.Add(1)
		go func(device string) {
			defer wg.Done()

			temp, err := c.client.GetTemperature(device)
			if err != nil {
				// Only log if it's not a "no data" error
				log.Printf("Failed to get temperature for %s: %v", device, err)
			}

			tempCh <- struct {
				device string
				temp   *int
			}{device: device, temp: temp}
		}(disk.Device)
	}

	// Wait for all polling to complete
	go func() {
		wg.Wait()
		close(tempCh)
	}()

	// Collect results
	newTemps := make(map[string]*int)
	for result := range tempCh {
		newTemps[result.device] = result.temp
	}

	// Update cache atomically
	c.mu.Lock()
	c.temperatures = newTemps
	c.mu.Unlock()

	log.Printf("SMART polling completed: %d disks polled, %d with temperature data",
		len(status.Disks), c.countValidTemperatures(newTemps))
}

// countValidTemperatures counts how many disks have valid temperature data
func (c *Cache) countValidTemperatures(temps map[string]*int) int {
	count := 0
	for _, temp := range temps {
		if temp != nil {
			count++
		}
	}
	return count
}

// GetTemperature returns the cached temperature for a device
func (c *Cache) GetTemperature(device string) *int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.temperatures[device]
}

// GetAllTemperatures returns all cached temperatures
func (c *Cache) GetAllTemperatures() map[string]*int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	temps := make(map[string]*int, len(c.temperatures))
	for device, temp := range c.temperatures {
		temps[device] = temp
	}
	return temps
}
