package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourusername/air-go/internal/db"
)

// DatabaseHealth represents database connectivity status (T091)
type DatabaseHealth struct {
	Status    string `json:"status"`          // connected, disconnected, error
	Message   string `json:"message"`         // Human-readable status message
	LatencyMs int64  `json:"latency_ms"`      // Ping latency in milliseconds
	Error     string `json:"error,omitempty"` // Error details if status is error
}

// Response represents the health check response structure (T091)
type Response struct {
	Status    string          `json:"status"`             // Overall status: ok, degraded
	Timestamp string          `json:"timestamp"`          // RFC3339 timestamp
	Database  *DatabaseHealth `json:"database,omitempty"` // Database health (optional)
}

// DBHealthChecker interface for checking database health
// This interface is implemented by *db.Client
type DBHealthChecker interface {
	HealthStatus(ctx context.Context) (*db.HealthStatus, error)
	IsConnected() bool
}

// Handler returns an HTTP handler for the health check endpoint
// If dbClient is nil, only basic health status is returned
func Handler(dbClient DBHealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		// Include database health if client is provided (T090)
		if dbClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			dbHealth, err := dbClient.HealthStatus(ctx)
			if err == nil && dbHealth != nil {
				response.Database = &DatabaseHealth{
					Status:    dbHealth.Status,
					Message:   dbHealth.Message,
					LatencyMs: dbHealth.LatencyMs,
					Error:     dbHealth.Error,
				}

				// If database is not connected, set overall status to degraded
				if dbHealth.Status != "connected" {
					response.Status = "degraded"
				}
			} else {
				// Database health check failed
				response.Database = &DatabaseHealth{
					Status:  "error",
					Message: "Failed to check database health",
				}
				if err != nil {
					response.Database.Error = err.Error()
				}
				response.Status = "degraded"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			// If encoding fails, log but don't change response
			// (headers already sent)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
