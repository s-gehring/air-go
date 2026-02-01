package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/db"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"go.mongodb.org/mongo-driver/bson"
)

// T024: E2E test for inventoryGet query success case
func TestInventoryGet_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test inventory
	inventoryID := "bb0e8400-e29b-41d4-a716-446655440000"
	seedInventory(t, dbClient, inventoryID, "NONE")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query
	result, err := queryResolver.InventoryGet(ctx, inventoryID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, inventoryID, result.Identifier)
}

// T025: E2E test for inventoryGet query not found case
func TestInventoryGet_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query with non-existent UUID
	nonExistentID := "cc0e8400-e29b-41d4-a716-446655440000"
	result, err := queryResolver.InventoryGet(ctx, nonExistentID)

	// Assertions: should return nil, not error
	require.NoError(t, err)
	assert.Nil(t, result)
}

// Helper: Seed inventory data
func seedInventory(t *testing.T, dbClient *db.Client, identifier, actionIndicator string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("inventories")

	doc := bson.M{
		"identifier":      identifier,
		"createDate":      time.Now().Format(time.RFC3339),
		"actionIndicator": actionIndicator,
		"isConsistent":    true,
		"isComplete":      true,
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}
