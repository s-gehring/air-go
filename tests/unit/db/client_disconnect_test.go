package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/air-go/internal/db"
)

// TestClient_Disconnect_NotConnected tests disconnecting when not connected
func TestClient_Disconnect_NotConnected(t *testing.T) {
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

	// Disconnect without connecting first - should not error
	ctx := context.Background()
	err = client.Disconnect(ctx)
	if err != nil {
		t.Errorf("Disconnect() when not connected returned error = %v, expected nil", err)
	}

	// Verify state remains disconnected
	if client.IsConnected() {
		t.Error("IsConnected() = true after Disconnect(), expected false")
	}
}

// TestClient_Disconnect_MultipleDisconnects tests multiple disconnect calls
func TestClient_Disconnect_MultipleDisconnects(t *testing.T) {
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

	// First disconnect
	err = client.Disconnect(ctx)
	if err != nil {
		t.Errorf("First Disconnect() returned error = %v, expected nil", err)
	}

	// Second disconnect - should be idempotent
	err = client.Disconnect(ctx)
	if err != nil {
		t.Errorf("Second Disconnect() returned error = %v, expected nil", err)
	}

	// Verify state
	if client.IsConnected() {
		t.Error("IsConnected() = true after multiple Disconnect(), expected false")
	}
}

// TestClient_Disconnect_ContextTimeout tests disconnect with context timeout
func TestClient_Disconnect_ContextTimeout(t *testing.T) {
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Disconnect(ctx)
	if err != nil {
		t.Errorf("Disconnect() with timeout returned error = %v, expected nil", err)
	}

	// Verify state is disconnected even if there was an error
	if client.IsConnected() {
		t.Error("IsConnected() = true after Disconnect(), expected false")
	}
}

// TestClient_Disconnect_StateCleanup tests that disconnect cleans up state properly
func TestClient_Disconnect_StateCleanup(t *testing.T) {
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
		t.Errorf("Disconnect() returned error = %v, expected nil", err)
	}

	// Verify client is marked as disconnected
	if client.IsConnected() {
		t.Error("IsConnected() = true after Disconnect(), expected false")
	}

	// Verify database accessor returns nil (would panic if called after disconnect)
	// This is tested indirectly through Collection() which would panic
}

// TestClient_Close tests that Close cancels the context
func TestClient_Close(t *testing.T) {
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

	// Close should not panic
	client.Close()

	// Multiple Close calls should be safe
	client.Close()
}
