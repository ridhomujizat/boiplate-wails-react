package dto

import "time"

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
	Uptime    string    `json:"uptime,omitempty"`
}

// DetailedHealthResponse represents a detailed health check response
type DetailedHealthResponse struct {
	Status     string            `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	Version    string            `json:"version"`
	Uptime     string            `json:"uptime"`
	System     SystemInfo        `json:"system"`
	Components map[string]string `json:"components"`
}

// SystemInfo represents system information
type SystemInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	NumCPU     int    `json:"numCpu"`
	GoRoutines int    `json:"goRoutines"`
	GoVersion  string `json:"goVersion"`
}
