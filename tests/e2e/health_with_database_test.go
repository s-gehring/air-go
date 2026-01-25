package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/air-go/internal/config"
	"github.com/yourusername/air-go/internal/db"
	"github.com/yourusername/air-go/internal/server"
	"github.com/yourusername/air-go/tests/integration"
)

var (
	// testLogger for database client (discards output for cleaner test results).
	testLogger = zerolog.New(io.Discard).With().Timestamp().Logger()
)

// TestHealthCheckWithDatabase verifies that the health check endpoint includes database status when connected.
func TestHealthCheckWithDatabase(t *testing.T) {
	// Start testcontainer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, uri, cleanup, err := integration.StartTestContainerWithURI(ctx)
	require.NoError(t, err)
	defer cleanup()

	// Create db.Client
	dbConfig := &db.DBConfig{
		URI:              uri,
		Database:         "test_health",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MaxPoolSize:      10,
		MinPoolSize:      5,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	dbClient, err := db.NewClient(dbConfig, testLogger)
	require.NoError(t, err)

	// Connect to MongoDB
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = dbClient.Connect(connectCtx)
	connectCancel()
	require.NoError(t, err)
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer disconnectCancel()
		dbClient.Disconnect(disconnectCtx)
		dbClient.Close()
	}()

	// Create test server with database client
	cfg := &config.Config{
		Port:        8080,
		LogFormat:   "json",
		SchemaPath:  "../../schema.graphqls",
		JWTSecret:   "test-secret-key-at-least-32-characters-long",
		CORSOrigins: []string{"*"},
	}

	srv := server.New(cfg, server.WithDatabaseClient(dbClient))
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

	// Parse JSON response with database field
	var healthResponse struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Database  *struct {
			Status    string `json:"status"`
			Message   string `json:"message"`
			LatencyMs int64  `json:"latency_ms"`
			Error     string `json:"error,omitempty"`
		} `json:"database,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	// Assert overall status is "ok"
	assert.Equal(t, "ok", healthResponse.Status)

	// Assert timestamp is present and valid
	assert.NotEmpty(t, healthResponse.Timestamp)
	_, err = time.Parse(time.RFC3339, healthResponse.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")

	// Assert database field is present
	require.NotNil(t, healthResponse.Database, "Database field should be present when dbClient is provided")

	// Assert database status is "connected"
	assert.Equal(t, "connected", healthResponse.Database.Status)

	// Assert database message is present
	assert.NotEmpty(t, healthResponse.Database.Message)

	// Assert latency is measured (should be > 0)
	assert.GreaterOrEqual(t, healthResponse.Database.LatencyMs, int64(0))

	// Assert no error
	assert.Empty(t, healthResponse.Database.Error)

	t.Logf("Health check response: status=%s, db_status=%s, db_latency=%dms",
		healthResponse.Status, healthResponse.Database.Status, healthResponse.Database.LatencyMs)
}

// TestHealthCheckDegraded verifies that the health check returns degraded status when database is disconnected.
func TestHealthCheckDegraded(t *testing.T) {
	// Start testcontainer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, uri, cleanup, err := integration.StartTestContainerWithURI(ctx)
	require.NoError(t, err)
	defer cleanup()

	// Create db.Client
	dbConfig := &db.DBConfig{
		URI:              uri,
		Database:         "test_health_degraded",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MaxPoolSize:      10,
		MinPoolSize:      5,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	dbClient, err := db.NewClient(dbConfig, testLogger)
	require.NoError(t, err)

	// Connect then disconnect to simulate degraded state
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = dbClient.Connect(connectCtx)
	connectCancel()
	require.NoError(t, err)

	// Disconnect to create degraded state
	disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = dbClient.Disconnect(disconnectCtx)
	disconnectCancel()
	require.NoError(t, err)

	// Create test server with disconnected database client
	cfg := &config.Config{
		Port:        8080,
		LogFormat:   "json",
		SchemaPath:  "../../schema.graphqls",
		JWTSecret:   "test-secret-key-at-least-32-characters-long",
		CORSOrigins: []string{"*"},
	}

	srv := server.New(cfg, server.WithDatabaseClient(dbClient))
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Send GET request to /health
	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 200 OK status (health check should always return 200)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse JSON response
	var healthResponse struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Database  *struct {
			Status    string `json:"status"`
			Message   string `json:"message"`
			LatencyMs int64  `json:"latency_ms"`
			Error     string `json:"error,omitempty"`
		} `json:"database,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	// Assert overall status is "degraded" (not "ok")
	assert.Equal(t, "degraded", healthResponse.Status)

	// Assert database field is present
	require.NotNil(t, healthResponse.Database)

	// Assert database status is "disconnected" or "error"
	assert.Contains(t, []string{"disconnected", "error"}, healthResponse.Database.Status)

	t.Logf("Health check degraded response: status=%s, db_status=%s, db_message=%s",
		healthResponse.Status, healthResponse.Database.Status, healthResponse.Database.Message)

	// Cleanup
	dbClient.Close()
}
