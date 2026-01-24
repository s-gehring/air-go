package resolvers

import (
	"context"
	"time"

	"github.com/yourusername/air-go/internal/graphql/generated"
)

// mapHealthStatus maps internal database health status to GraphQL DatabaseHealth type (T089)
// This helper ensures consistent status mapping between HTTP and GraphQL health endpoints
func mapDatabaseHealth(status, message string, latencyMs int64, errorMsg string) *generated.DatabaseHealth {
	// Convert latency to int64 as required by GraphQL schema
	return &generated.DatabaseHealth{
		Status:    status,
		Message:   message,
		LatencyMs: latencyMs,
		Error:     &errorMsg,
	}
}

// resolveHealth implements the health query resolver (T087)
// Returns system health status including optional database health
func (r *Resolver) resolveHealth(ctx context.Context) (*generated.Health, error) {
	health := &generated.Health{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Include database health if client is available
	if r.DBClient != nil {
		// Use 2-second timeout for health checks to prevent blocking GraphQL queries
		healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		dbHealth, err := r.DBClient.HealthStatus(healthCtx)
		if err == nil && dbHealth != nil {
			errorMsg := dbHealth.Error
			if errorMsg == "" {
				errorMsg = "" // Ensure non-nil string for GraphQL
			}

			health.Database = mapDatabaseHealth(
				dbHealth.Status,
				dbHealth.Message,
				dbHealth.LatencyMs,
				errorMsg,
			)

			// Set overall status to degraded if database is not connected
			if dbHealth.Status != "connected" {
				health.Status = "degraded"
			}
		} else {
			// Database health check failed
			errorMsg := ""
			if err != nil {
				errorMsg = err.Error()
			}

			health.Database = mapDatabaseHealth(
				"error",
				"Failed to check database health",
				0,
				errorMsg,
			)
			health.Status = "degraded"
		}
	}

	return health, nil
}
