package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/air-go/internal/db"
)

// TestClient_Connect_InvalidConfiguration tests client creation with invalid config
func TestClient_Connect_InvalidConfiguration(t *testing.T) {
	logger := zerolog.Nop()

	tests := []struct {
		name   string
		config *db.DBConfig
	}{
		{
			name:   "nil configuration",
			config: nil,
		},
		{
			name: "empty URI",
			config: &db.DBConfig{
				URI:              "",
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      5,
				MaxPoolSize:      10,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			},
		},
		{
			name: "invalid pool size",
			config: &db.DBConfig{
				URI:              "mongodb://localhost:27017",
				Database:         "testdb",
				ConnectTimeout:   30 * time.Second,
				OperationTimeout: 10 * time.Second,
				MinPoolSize:      15,
				MaxPoolSize:      10,
				MaxConnIdleTime:  5 * time.Minute,
				MaxRetryAttempts: 3,
				RetryBaseDelay:   1 * time.Second,
				RetryMaxDelay:    10 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.NewClient(tt.config, logger)
			if err == nil {
				t.Error("NewClient() expected error for invalid configuration, got nil")
			}
		})
	}
}

// TestClient_Connect_ContextCancellation tests connection with context cancellation
func TestClient_Connect_ContextCancellation(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://nonexistent-host:27017",
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

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect() expected error for cancelled context, got nil")
	}

	// Verify client is not connected after failed connection
	if client.IsConnected() {
		t.Error("IsConnected() = true, expected false after failed connection")
	}
}

// TestClient_Connect_AlreadyConnected tests connecting an already connected client
func TestClient_Connect_AlreadyConnected(t *testing.T) {
	// This test requires a mock or a real MongoDB instance
	// For now, we'll test the error path with a non-existent host
	// and verify that subsequent Connect() calls fail appropriately
	t.Skip("Requires MongoDB instance or mock - integration test")
}

// TestClient_Connect_InvalidHost tests connection to invalid host
func TestClient_Connect_InvalidHost(t *testing.T) {
	logger := zerolog.Nop()

	config := &db.DBConfig{
		URI:              "mongodb://invalid-host-that-does-not-exist:27017",
		Database:         "testdb",
		ConnectTimeout:   10 * time.Second,
		OperationTimeout: 5 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 1, // Reduce retries for faster test
		RetryBaseDelay:   100 * time.Millisecond,
		RetryMaxDelay:    500 * time.Millisecond,
	}

	client, err := db.NewClient(config, logger)
	if err != nil {
		t.Fatalf("NewClient() unexpected error = %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect() expected error for invalid host, got nil")
	}

	// Verify client is not connected
	if client.IsConnected() {
		t.Error("IsConnected() = true, expected false after failed connection")
	}
}

// TestClient_NewClient_ValidConfiguration tests successful client creation
func TestClient_NewClient_ValidConfiguration(t *testing.T) {
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
		t.Errorf("NewClient() unexpected error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}

	// Verify initial state
	if client.IsConnected() {
		t.Error("IsConnected() = true, expected false before Connect()")
	}

	client.Close()
}

// TestClient_Connect_StateTransitions tests connection state transitions
func TestClient_Connect_StateTransitions(t *testing.T) {
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

	// Initial state: not connected
	if client.IsConnected() {
		t.Error("Initial state: IsConnected() = true, expected false")
	}

	// Note: Actual connection test would require MongoDB instance
	// State verification is tested in integration tests
}
