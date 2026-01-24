package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/air-go/internal/db"
)

// TestHealthStatus_HealthyConnection tests health status with healthy connection (T077)
func TestHealthStatus_HealthyConnection(t *testing.T) {
	// This test requires actual MongoDB connection
	// For unit testing, we test the disconnected case
	// Healthy connection is tested in integration tests
	t.Skip("Requires MongoDB instance - integration test")
}

// TestHealthStatus_WithCache tests 5-second TTL caching (T078)
func TestHealthStatus_WithCache(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://localhost:27017",
		Database:         "testdb",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	client, err := db.NewClient(config, logger)
	if err != nil {
		t.Fatalf("NewClient() unexpected error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First call - creates cache
	status1, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() first call returned error = %v", err)
	}

	if status1 == nil {
		t.Fatal("HealthStatus() returned nil status")
	}

	// Capture timestamp
	timestamp1 := status1.Timestamp

	// Immediate second call - should return cached result
	status2, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() second call returned error = %v", err)
	}

	if status2 == nil {
		t.Fatal("HealthStatus() returned nil status on second call")
	}

	// Timestamps should be identical (cached)
	if !timestamp1.Equal(status2.Timestamp) {
		t.Errorf("Cache not working: timestamps differ (first=%v, second=%v)",
			timestamp1, status2.Timestamp)
	}

	// Wait for cache to expire (5 seconds + buffer)
	time.Sleep(6 * time.Second)

	// Third call after expiry - should have new timestamp
	status3, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() third call returned error = %v", err)
	}

	if status3 == nil {
		t.Fatal("HealthStatus() returned nil status on third call")
	}

	// Timestamp should be different (cache expired)
	if timestamp1.Equal(status3.Timestamp) {
		t.Error("Cache not expiring: timestamps should differ after 6 seconds")
	}

	// Verify new timestamp is later
	if !status3.Timestamp.After(timestamp1) {
		t.Errorf("New timestamp (%v) should be after old timestamp (%v)",
			status3.Timestamp, timestamp1)
	}
}

// TestHealthStatus_WhenDisconnected tests disconnected status (T079)
func TestHealthStatus_WhenDisconnected(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://localhost:27017",
		Database:         "testdb",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	client, err := db.NewClient(config, logger)
	if err != nil {
		t.Fatalf("NewClient() unexpected error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get health status without connecting
	status, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() returned error = %v, expected nil", err)
	}

	if status == nil {
		t.Fatal("HealthStatus() returned nil status")
	}

	// Verify disconnected status
	if status.Status != "disconnected" {
		t.Errorf("Status = %q, expected %q", status.Status, "disconnected")
	}

	if status.Message != "MongoDB not connected" {
		t.Errorf("Message = %q, expected %q", status.Message, "MongoDB not connected")
	}

	if status.LatencyMs != 0 {
		t.Errorf("LatencyMs = %d, expected 0 when disconnected", status.LatencyMs)
	}

	// Verify timestamp is set
	if status.Timestamp.IsZero() {
		t.Error("Timestamp should be set even when disconnected")
	}

	// Verify error field is empty
	if status.Error != "" {
		t.Errorf("Error field should be empty when disconnected, got %q", status.Error)
	}
}

// TestHealthStatus_StatusValues tests different status values
func TestHealthStatus_StatusValues(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://localhost:27017",
		Database:         "testdb",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	client, err := db.NewClient(config, logger)
	if err != nil {
		t.Fatalf("NewClient() unexpected error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test: Disconnected state
	status, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() returned error = %v", err)
	}

	// Valid status values: "connected", "disconnected", "error"
	validStatuses := map[string]bool{
		"connected":    true,
		"disconnected": true,
		"error":        true,
	}

	if !validStatuses[status.Status] {
		t.Errorf("Invalid status value: %q, expected one of: connected, disconnected, error",
			status.Status)
	}
}

// TestHealthStatus_ConcurrentAccess tests thread-safe concurrent access
func TestHealthStatus_ConcurrentAccess(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://localhost:27017",
		Database:         "testdb",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	client, err := db.NewClient(config, logger)
	if err != nil {
		t.Fatalf("NewClient() unexpected error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Spawn multiple goroutines calling HealthStatus concurrently
	done := make(chan bool)
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_, err := client.HealthStatus(ctx)
				if err != nil {
					errors <- err
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent HealthStatus() error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Failed %d concurrent HealthStatus calls", errorCount)
	}

	// Test should complete without data races
}

// TestHealthStatus_LatencyMeasurement tests latency measurement
func TestHealthStatus_LatencyMeasurement(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://localhost:27017",
		Database:         "testdb",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	client, err := db.NewClient(config, logger)
	if err != nil {
		t.Fatalf("NewClient() unexpected error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// When disconnected, latency should be 0
	status, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() returned error = %v", err)
	}

	if status.LatencyMs < 0 {
		t.Errorf("LatencyMs should never be negative, got %d", status.LatencyMs)
	}

	// When disconnected, latency must be 0
	if !client.IsConnected() && status.LatencyMs != 0 {
		t.Errorf("LatencyMs should be 0 when disconnected, got %d", status.LatencyMs)
	}
}
