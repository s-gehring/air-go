package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/air-go/internal/config"
	"github.com/yourusername/air-go/internal/server"
)

// TestHealthCheckEndpoint verifies that the health check endpoint returns 200 OK with proper JSON response
func TestHealthCheckEndpoint(t *testing.T) {
	// Create test server
	cfg := &config.Config{
		Port:        8080,
		LogFormat:   "json",
		SchemaPath:  "../../schema.graphqls",
		JWTSecret:   "test-secret-key-at-least-32-characters-long",
		CORSOrigins: []string{"*"},
	}

	srv := server.New(cfg)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Send GET request to /health
	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 OK status
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Assert Content-Type is application/json
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Parse JSON response
	var healthResponse struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
	}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	// Assert response contains status "ok"
	assert.Equal(t, "ok", healthResponse.Status)

	// Assert timestamp is present and valid
	assert.NotEmpty(t, healthResponse.Timestamp)
	_, err = time.Parse(time.RFC3339, healthResponse.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")
}

// TestHealthCheckPerformance verifies that the health check responds in <100ms for 99% of requests
func TestHealthCheckPerformance(t *testing.T) {
	// Create test server
	cfg := &config.Config{
		Port:        8080,
		LogFormat:   "json",
		SchemaPath:  "../../schema.graphqls",
		JWTSecret:   "test-secret-key-at-least-32-characters-long",
		CORSOrigins: []string{"*"},
	}

	srv := server.New(cfg)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Send 100 requests and measure response times
	const numRequests = 100
	const maxDuration = 100 * time.Millisecond
	const percentile99 = 99

	durations := make([]time.Duration, numRequests)

	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := http.Get(ts.URL + "/health")
		duration := time.Since(start)
		durations[i] = duration

		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Calculate 99th percentile
	// Sort durations
	sortedDurations := make([]time.Duration, numRequests)
	copy(sortedDurations, durations)
	for i := 0; i < len(sortedDurations); i++ {
		for j := i + 1; j < len(sortedDurations); j++ {
			if sortedDurations[i] > sortedDurations[j] {
				sortedDurations[i], sortedDurations[j] = sortedDurations[j], sortedDurations[i]
			}
		}
	}

	// Get 99th percentile (99 out of 100)
	p99Index := (numRequests * percentile99) / 100
	if p99Index >= numRequests {
		p99Index = numRequests - 1
	}
	p99Duration := sortedDurations[p99Index]

	// Assert 99th percentile is under 100ms
	assert.Less(t, p99Duration, maxDuration,
		"99th percentile response time should be less than 100ms, got %v", p99Duration)

	// Log performance stats
	t.Logf("Performance stats: min=%v, median=%v, p99=%v, max=%v",
		sortedDurations[0],
		sortedDurations[numRequests/2],
		p99Duration,
		sortedDurations[numRequests-1])
}

// TestHealthCheckWithoutAuthentication verifies that the health check endpoint does NOT require authentication
func TestHealthCheckWithoutAuthentication(t *testing.T) {
	// Create test server
	cfg := &config.Config{
		Port:        8080,
		LogFormat:   "json",
		SchemaPath:  "../../schema.graphqls",
		JWTSecret:   "test-secret-key-at-least-32-characters-long",
		CORSOrigins: []string{"*"},
	}

	srv := server.New(cfg)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Send GET request to /health WITHOUT Authorization header
	req, err := http.NewRequest("GET", ts.URL+"/health", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 OK status (not 401 Unauthorized)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse JSON response to verify it's a valid health check response
	var healthResponse struct {
		Status string `json:"status"`
	}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	// Assert response contains status "ok"
	assert.Equal(t, "ok", healthResponse.Status)
}
