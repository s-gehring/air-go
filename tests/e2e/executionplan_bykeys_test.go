package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// T053: E2E test for executionPlanByKeysGet with default ordering
func TestExecutionPlanByKeysGet_DefaultOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed execution plans
	id1 := "800e8400-e29b-41d4-a716-446655440000"
	id2 := "900e8400-e29b-41d4-a716-446655440001"
	
	seedExecutionPlan(t, dbClient, id1, "NONE")
	seedExecutionPlan(t, dbClient, id2, "NONE")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute without order parameter
	identifiers := []string{id2, id1} // reversed order
	result, err := queryResolver.ExecutionPlanByKeysGet(ctx, identifiers, nil)

	// Assertions - should be ordered by identifier ASC (default)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, id1, result[0].Identifier)
	assert.Equal(t, id2, result[1].Identifier)
}

// T054: E2E test for executionPlanByKeysGet deduplication
func TestExecutionPlanByKeysGet_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 1 execution plan
	id1 := "800e8400-e29b-41d4-a716-446655440010"
	seedExecutionPlan(t, dbClient, id1, "NONE")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Query with duplicate ID
	identifiers := []string{id1, id1}
	result, err := queryResolver.ExecutionPlanByKeysGet(ctx, identifiers, nil)

	// Assertions - should return execution plan only once
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].Identifier)
}
