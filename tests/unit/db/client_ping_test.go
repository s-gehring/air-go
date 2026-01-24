package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/air-go/internal/db"
)

// TestClient_Ping_NotConnected tests ping when client is not connected
func TestClient_Ping_NotConnected(t *testing.T) {
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

	// Ping without connecting first - should return error
	err = client.Ping(ctx)
	if err == nil {
		t.Error("Ping() when not connected returned nil, expected error")
	}
}

// TestClient_Ping_AfterDisconnect tests ping after disconnect
func TestClient_Ping_AfterDisconnect(t *testing.T) {
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

	// Disconnect (even though we never connected)
	err = client.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Disconnect() unexpected error = %v", err)
	}

	// Ping after disconnect should return error
	err = client.Ping(ctx)
	if err == nil {
		t.Error("Ping() after disconnect returned nil, expected error")
	}
}

// TestClient_Ping_ContextCancellation tests ping with cancelled context
func TestClient_Ping_ContextCancellation(t *testing.T) {
	// This test would require a connected client
	// For unit testing, we verify that ping respects context
	t.Skip("Requires MongoDB instance - integration test")
}

// TestClient_Ping_Success tests successful ping
func TestClient_Ping_Success(t *testing.T) {
	// This test requires actual MongoDB connection
	t.Skip("Requires MongoDB instance - integration test")
}

// TestClient_Ping_UpdatesLastPingTime tests that ping updates last ping timestamp
func TestClient_Ping_UpdatesLastPingTime(t *testing.T) {
	// This test requires actual MongoDB connection to verify state updates
	t.Skip("Requires MongoDB instance - integration test")
}
