package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/air-go/internal/db"
)

// TestClient_IsConnected_InitialState tests initial connection state
func TestClient_IsConnected_InitialState(t *testing.T) {
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

	// Initial state should be disconnected
	if client.IsConnected() {
		t.Error("IsConnected() = true for new client, expected false")
	}
}

// TestClient_IsConnected_AfterFailedConnection tests state after failed connection
func TestClient_IsConnected_AfterFailedConnection(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://invalid-host:27017",
		Database:         "testdb",
		ConnectTimeout:   10 * time.Second,
		OperationTimeout: 5 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 1,
		RetryBaseDelay:   100 * time.Millisecond,
		RetryMaxDelay:    500 * time.Millisecond,
	}

	client, err := db.NewClient(config, logger)
	if err != nil {
		t.Fatalf("NewClient() unexpected error = %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Attempt to connect to invalid host
	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect() to invalid host returned nil, expected error")
	}

	// State should remain disconnected after failed connection
	if client.IsConnected() {
		t.Error("IsConnected() = true after failed connection, expected false")
	}
}

// TestClient_IsConnected_AfterDisconnect tests state after disconnect
func TestClient_IsConnected_AfterDisconnect(t *testing.T) {
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

	// Disconnect
	err = client.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Disconnect() unexpected error = %v", err)
	}

	// State should be disconnected
	if client.IsConnected() {
		t.Error("IsConnected() = true after Disconnect(), expected false")
	}
}

// TestClient_IsConnected_ThreadSafety tests concurrent access to IsConnected
func TestClient_IsConnected_ThreadSafety(t *testing.T) {
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

	// Spawn multiple goroutines to check connection state concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = client.IsConnected()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test should complete without data races
}

// TestClient_Database_NotConnected tests Database() accessor when not connected
func TestClient_Database_NotConnected(t *testing.T) {
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

	// Database() should return nil when not connected
	db := client.Database()
	if db != nil {
		t.Error("Database() returned non-nil before Connect(), expected nil")
	}
}

// TestClient_Collection_PanicsWhenNotConnected tests Collection() panic behavior
func TestClient_Collection_PanicsWhenNotConnected(t *testing.T) {
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

	// Collection() should panic when database is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Collection() did not panic when database is nil")
		}
	}()

	client.Collection("test_collection")
}

// TestClient_HealthStatus_NotConnected tests health status when not connected
func TestClient_HealthStatus_NotConnected(t *testing.T) {
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
	status, err := client.HealthStatus(ctx)

	if err != nil {
		t.Errorf("HealthStatus() returned error = %v, expected nil", err)
	}

	if status == nil {
		t.Fatal("HealthStatus() returned nil status")
	}

	if status.Status != "disconnected" {
		t.Errorf("HealthStatus().Status = %q, expected %q", status.Status, "disconnected")
	}

	if status.Message != "MongoDB not connected" {
		t.Errorf("HealthStatus().Message = %q, expected %q", status.Message, "MongoDB not connected")
	}

	if status.LatencyMs != 0 {
		t.Errorf("HealthStatus().LatencyMs = %d, expected 0", status.LatencyMs)
	}
}

// TestClient_HealthStatus_Caching tests health status caching behavior
func TestClient_HealthStatus_Caching(t *testing.T) {
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

	// First call
	status1, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() first call returned error = %v", err)
	}

	// Second call within cache TTL (5 seconds)
	status2, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() second call returned error = %v", err)
	}

	// Both calls should return the same cached status
	if status1.Timestamp != status2.Timestamp {
		t.Error("HealthStatus() cache not working - timestamps differ")
	}

	// Wait for cache to expire (5 seconds + buffer)
	time.Sleep(6 * time.Second)

	// Third call after cache expiry
	status3, err := client.HealthStatus(ctx)
	if err != nil {
		t.Fatalf("HealthStatus() third call returned error = %v", err)
	}

	// Third call should have a new timestamp
	if status1.Timestamp == status3.Timestamp {
		t.Error("HealthStatus() cache not expiring - timestamps should differ after 6 seconds")
	}
}
