package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupTestDB creates a test MongoDB connection
// For integration tests, this connects to a real MongoDB instance (testcontainers or local)
func SetupTestDB(t *testing.T, uri string) *mongo.Database {
	t.Helper()
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err, "Failed to connect to test database")

	// Verify connection
	err = client.Ping(ctx, nil)
	require.NoError(t, err, "Failed to ping test database")

	dbName := "test_air_go"
	db := client.Database(dbName)

	t.Logf("Connected to test database: %s", dbName)

	return db
}

// TeardownTestDB cleans up test database
func TeardownTestDB(t *testing.T, db *mongo.Database) {
	t.Helper()
	ctx := context.Background()

	// Drop test database to clean up
	err := db.Drop(ctx)
	require.NoError(t, err, "Failed to drop test database")

	// Disconnect client
	err = db.Client().Disconnect(ctx)
	require.NoError(t, err, "Failed to disconnect from test database")

	t.Log("Test database cleaned up")
}

// CleanCollection removes all documents from a collection
func CleanCollection(t *testing.T, db *mongo.Database, collectionName string) {
	t.Helper()
	ctx := context.Background()

	collection := db.Collection(collectionName)
	_, err := collection.DeleteMany(ctx, map[string]interface{}{})
	require.NoError(t, err, "Failed to clean collection: %s", collectionName)

	t.Logf("Cleaned collection: %s", collectionName)
}

// CreateIndexes creates required indexes for testing
func CreateIndexes(t *testing.T, db *mongo.Database) {
	t.Helper()
	ctx := context.Background()

	collection := db.Collection("inventories")

	// Create identifier index (unique)
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"identifier": 1},
		Options: options.Index().SetUnique(true),
	})
	require.NoError(t, err, "Failed to create identifier index")

	// Create customerId index (sparse for ordering)
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"customerId": 1},
		Options: options.Index().SetSparse(true),
	})
	require.NoError(t, err, "Failed to create customerId index")

	t.Log("Created required indexes")
}
