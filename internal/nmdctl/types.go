package nmdctl

import "time"

// Status represents the complete output from nmdctl status -o json
type Status struct {
	Array  Array   `json:"array"`
	Resync *Resync `json:"resync,omitempty"`
	Disks  []Disk  `json:"disks"`
}

// Array represents the overall array information
type Array struct {
	Label            string   `json:"label"`
	State            string   `json:"state"`
	Superblock       string   `json:"superblock"`
	DisksPresent     int      `json:"disks_present"`
	DisksImported    int      `json:"disks_imported"`
	DisksUnassigned  int      `json:"disks_unassigned"`
	TotalSlots       int      `json:"total_slots"`
	Health           Health   `json:"health"`
	Size             Size     `json:"size"`
	Counters         Counters `json:"counters"`
	LastSync         LastSync `json:"last_sync"`
}

// Health represents the array health status
type Health struct {
	Status  string `json:"status"`
	Details string `json:"details"`
	Code    int    `json:"code"`
}

// Size represents the array size information
type Size struct {
	DataGB              int  `json:"data_gb"`
	DataDiskCount       int  `json:"data_disk_count"`
	HasParity           bool `json:"has_parity"`
	HasSecondParity     bool `json:"has_second_parity"`
	ParitySizeGB        int  `json:"parity_size_gb"`
	SecondParitySizeGB  int  `json:"second_parity_size_gb"`
}

// Counters represents disk counters
type Counters struct {
	Missing    int `json:"missing"`
	Invalid    int `json:"invalid"`
	Wrong      int `json:"wrong"`
	Disabled   int `json:"disabled"`
	Replaced   int `json:"replaced"`
	New        int `json:"new"`
	SyncErrors int `json:"sync_errors"`
	DiskErrors int `json:"disk_errors"`
}

// LastSync represents the last sync operation
type LastSync struct {
	Timestamp      int64  `json:"timestamp"`
	AgeSeconds     int    `json:"age_seconds"`
	ElapsedSeconds int    `json:"elapsed_seconds"`
	Status         string `json:"status"`
}

// Resync represents an active resync operation
type Resync struct {
	Active          bool    `json:"active"`
	Paused          bool    `json:"paused"`
	Pending         bool    `json:"pending"`
	Action          string  `json:"action"`
	ProgressPercent int     `json:"progress_percent"`
	PositionGB      float64 `json:"position_gb"`
	SizeGB          int     `json:"size_gb"`
	RateMBPS        float64 `json:"rate_mb_s"`
	ElapsedSeconds  int     `json:"elapsed_seconds"`
	ETASeconds      int     `json:"eta_seconds"`
}

// Disk represents a single disk in the array
type Disk struct {
	Slot           int         `json:"slot"`
	Type           string      `json:"type"`
	SizeKB         int64       `json:"size_kb"`
	SizeGB         int         `json:"size_gb"`
	Device         string      `json:"device"`
	Status         string      `json:"status"`
	Errors         int         `json:"errors"`
	Reads          int64       `json:"reads"`
	Writes         int64       `json:"writes"`
	DiskID         string      `json:"disk_id"`
	DiskName       string      `json:"disk_name"`
	Filesystem     *Filesystem `json:"filesystem,omitempty"`
	Temperature    *int        `json:"temperature,omitempty"`     // Temperature in Celsius from SMART data
	Note           *string     `json:"note,omitempty"`            // User note for this disk
	NoteUpdatedAt  *time.Time  `json:"note_updated_at,omitempty"` // When note was last updated
}

// Filesystem represents the filesystem information for a disk
type Filesystem struct {
	Type       string `json:"type"`
	Mountpoint string `json:"mountpoint"`
	Usage      string `json:"usage"`
}
