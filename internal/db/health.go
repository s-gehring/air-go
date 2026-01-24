package db

import "time"

// HealthStatus represents database health check result
type HealthStatus struct {
	Status    string    `json:"status"`              // "connected", "disconnected", "error"
	Message   string    `json:"message"`             // Human-readable message
	LatencyMs int64     `json:"latency_ms"`          // Ping latency in milliseconds
	Timestamp time.Time `json:"timestamp"`           // Check timestamp
	Error     string    `json:"error,omitempty"`     // Error details if unhealthy
}

// healthCache stores the last health check result with TTL
type healthCache struct {
	status    *HealthStatus
	expiresAt time.Time
}
