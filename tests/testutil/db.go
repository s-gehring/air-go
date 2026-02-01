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

	// T001-T004: Create indexes for all entity collections

	// Inventories collection (existing)
	inventoriesCol := db.Collection("inventories")
	_, err := inventoriesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"identifier": 1},
		Options: options.Index().SetUnique(true),
	})
	require.NoError(t, err, "Failed to create inventories identifier index")

	_, err = inventoriesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"customerId": 1},
		Options: options.Index().SetSparse(true),
	})
	require.NoError(t, err, "Failed to create inventories customerId index")

	// Customers collection (existing unique index, add compound for sorting)
	customersCol := db.Collection("customers")
	_, err = customersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"identifier": 1},
		Options: options.Index().SetUnique(true),
	})
	require.NoError(t, err, "Failed to create customers identifier index")

	_, err = customersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"lastName": 1, "identifier": 1},
	})
	require.NoError(t, err, "Failed to create customers lastName+identifier index")

	// T001: Employees collection indexes
	employeesCol := db.Collection("employees")
	_, err = employeesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"identifier": 1},
		Options: options.Index().SetUnique(true),
	})
	require.NoError(t, err, "Failed to create employees identifier index")

	_, err = employeesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"lastName": 1, "identifier": 1},
	})
	require.NoError(t, err, "Failed to create employees lastName+identifier index")

	// T002: Teams collection indexes
	teamsCol := db.Collection("teams")
	_, err = teamsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"identifier": 1},
		Options: options.Index().SetUnique(true),
	})
	require.NoError(t, err, "Failed to create teams identifier index")

	_, err = teamsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"name": 1, "identifier": 1},
	})
	require.NoError(t, err, "Failed to create teams name+identifier index")

	// T003: ExecutionPlans collection indexes
	executionPlansCol := db.Collection("executionPlans")
	_, err = executionPlansCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"identifier": 1},
		Options: options.Index().SetUnique(true),
	})
	require.NoError(t, err, "Failed to create executionPlans identifier index")

	_, err = executionPlansCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"customerId": 1, "identifier": 1},
	})
	require.NoError(t, err, "Failed to create executionPlans customerId+identifier index")

	// T004: ReferencePortfolios collection indexes
	referencePortfoliosCol := db.Collection("referencePortfolios")
	_, err = referencePortfoliosCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"identifier": 1},
		Options: options.Index().SetUnique(true),
	})
	require.NoError(t, err, "Failed to create referencePortfolios identifier index")

	_, err = referencePortfoliosCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"customerId": 1, "identifier": 1},
	})
	require.NoError(t, err, "Failed to create referencePortfolios customerId+identifier index")

	t.Log("Created required indexes for all entity collections")
}
