package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// T047: E2E test for employeeByKeysGet with ordering by lastName ASC
func TestEmployeeByKeysGet_OrderByLastNameASC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed employees with different lastNames
	id1 := "500e8400-e29b-41d4-a716-446655440000"
	id2 := "600e8400-e29b-41d4-a716-446655440001"
	id3 := "700e8400-e29b-41d4-a716-446655440002"
	
	seedEmployee(t, dbClient, id1, "Alice", "Zimmerman", "INIT")
	seedEmployee(t, dbClient, id2, "Bob", "Anderson", "INIT")
	seedEmployee(t, dbClient, id3, "Charlie", "Brown", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute with lastName ASC ordering
	identifiers := []string{id1, id2, id3}
	ascSort := generated.SortEnumTypeAsc
	order := []*generated.EmployeeQuerySorterInput{
		{LastName: &ascSort},
	}
	
	result, err := queryResolver.EmployeeByKeysGet(ctx, identifiers, order)

	// Assertions - should be ordered: Anderson, Brown, Zimmerman
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, "Anderson", *result[0].LastName)
	assert.Equal(t, "Brown", *result[1].LastName)
	assert.Equal(t, "Zimmerman", *result[2].LastName)
}

// T048: E2E test for employeeByKeysGet deduplication
func TestEmployeeByKeysGet_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 1 employee
	id1 := "500e8400-e29b-41d4-a716-446655440010"
	seedEmployee(t, dbClient, id1, "Alice", "Smith", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Query with duplicate ID
	identifiers := []string{id1, id1, id1}
	result, err := queryResolver.EmployeeByKeysGet(ctx, identifiers, nil)

	// Assertions - should return employee only once
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].Identifier)
}
