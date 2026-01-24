package health

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response represents the health check response structure
type Response struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Handler returns an HTTP handler for the health check endpoint
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
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
