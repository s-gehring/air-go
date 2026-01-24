package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestContainerConfig holds configuration for test MongoDB container (T040)
type TestContainerConfig struct {
	Image        string        // MongoDB Docker image (default: mongo:7.0)
	Port         string        // Exposed port (default: 27017)
	Database     string        // Test database name (default: test_db)
	CleanupMode  string        // cleanup mode: "drop" (default) or "terminate"
	StartTimeout time.Duration // Container startup timeout (default: 60s)
}

// DefaultTestContainerConfig returns default configuration for test containers
func DefaultTestContainerConfig() *TestContainerConfig {
	return &TestContainerConfig{
		Image:        "mongo:7.0",
		Port:         "27017/tcp",
		Database:     "test_db",
		CleanupMode:  "drop",
		StartTimeout: 60 * time.Second,
	}
}

// StartTestContainer starts a MongoDB test container and returns connected client (T041)
// Returns: (*mongo.Client, cleanup function, error)
// The cleanup function should be called with defer to ensure proper cleanup
func StartTestContainer(ctx context.Context) (*mongo.Client, func(), error) {
	return StartTestContainerWithConfig(ctx, DefaultTestContainerConfig())
}

// StartTestContainerWithConfig starts a MongoDB test container with custom configuration
func StartTestContainerWithConfig(ctx context.Context, config *TestContainerConfig) (*mongo.Client, func(), error) {
	if config == nil {
		config = DefaultTestContainerConfig()
	}

	// Create container request
	req := testcontainers.ContainerRequest{
		Image:        config.Image,
		ExposedPorts: []string{config.Port},
		WaitingFor: wait.ForLog("Waiting for connections").
			WithStartupTimeout(config.StartTimeout),
		Env: map[string]string{
			"MONGO_INITDB_DATABASE": config.Database,
		},
	}

	// Start container
	mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start MongoDB container: %w", err)
	}

	// Get container host and port
	host, err := mongoContainer.Host(ctx)
	if err != nil {
		mongoContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := mongoContainer.MappedPort(ctx, "27017")
	if err != nil {
		mongoContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	// Build connection string
	uri := fmt.Sprintf("mongodb://%s:%s", host, mappedPort.Port())

	// Connect to MongoDB
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		mongoContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = client.Ping(pingCtx, nil)
	cancel()

	if err != nil {
		client.Disconnect(ctx)
		mongoContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Cleanup function
	cleanup := func() {
		// Cleanup database before disconnecting (if in drop mode)
		if config.CleanupMode == "drop" {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = CleanupTestDatabase(cleanupCtx, client, config.Database)
			cleanupCancel()
		}

		// Disconnect client
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = client.Disconnect(disconnectCtx)
		disconnectCancel()

		// Terminate container
		terminateCtx, terminateCancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = mongoContainer.Terminate(terminateCtx)
		terminateCancel()
	}

	return client, cleanup, nil
}

// CleanupTestDatabase drops and recreates a clean database (T042)
// This enables test isolation by ensuring each test starts with a clean state
// Performance: Completes in <2s per SC-002
func CleanupTestDatabase(ctx context.Context, client *mongo.Client, dbName string) error {
	if client == nil {
		return fmt.Errorf("client cannot be nil")
	}

	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	startTime := time.Now()

	// Drop database
	err := client.Database(dbName).Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	// Recreate database by creating a dummy collection and removing it
	// This ensures the database exists for subsequent operations
	dummyCollection := client.Database(dbName).Collection("_init")
	_, err = dummyCollection.InsertOne(ctx, map[string]interface{}{"_init": true})
	if err != nil {
		return fmt.Errorf("failed to recreate database %s: %w", dbName, err)
	}

	err = dummyCollection.Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup init collection: %w", err)
	}

	duration := time.Since(startTime)

	// Verify performance constraint (SC-002: <2s)
	if duration > 2*time.Second {
		return fmt.Errorf("cleanup took %v, exceeds 2s requirement (SC-002)", duration)
	}

	return nil
}

// StopTestContainer stops and removes a test container (T043)
// This is typically called via the cleanup function returned by StartTestContainer
func StopTestContainer(ctx context.Context, container testcontainers.Container) error {
	if container == nil {
		return nil
	}

	return container.Terminate(ctx)
}
