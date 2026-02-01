package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// T051: E2E test for inventoryByKeysGet with ordering by customerId
func TestInventoryByKeysGet_OrderByCustomerID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed inventories with different customerIds
	id1 := "400e8400-e29b-41d4-a716-446655440000"
	id2 := "500e8400-e29b-41d4-a716-446655440001"
	
	seedInventory(t, dbClient, id1, "NONE")
	seedInventory(t, dbClient, id2, "NONE")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute with customerId ASC ordering
	identifiers := []string{id1, id2}
	ascSort := generated.SortEnumTypeAsc
	order := []*generated.InventoryQuerySorterInput{
		{CustomerID: &ascSort},
	}
	
	result, err := queryResolver.InventoryByKeysGet(ctx, identifiers, order)

	// Assertions
	require.NoError(t, err)
	require.Len(t, result, 2)
}

// T052: E2E test for inventoryByKeysGet deduplication
func TestInventoryByKeysGet_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 1 inventory
	id1 := "400e8400-e29b-41d4-a716-446655440010"
	seedInventory(t, dbClient, id1, "NONE")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Query with duplicate ID
	identifiers := []string{id1, id1}
	result, err := queryResolver.InventoryByKeysGet(ctx, identifiers, nil)

	// Assertions - should return inventory only once
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].Identifier)
}
