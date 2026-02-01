package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// T049: E2E test for teamByKeysGet with default ordering (no sorter input)
func TestTeamByKeysGet_DefaultOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed teams (will be ordered by identifier by default)
	id1 := "300e8400-e29b-41d4-a716-446655440000"
	id2 := "100e8400-e29b-41d4-a716-446655440001"
	id3 := "200e8400-e29b-41d4-a716-446655440002"
	
	seedTeam(t, dbClient, id1, "Team C", "INIT")
	seedTeam(t, dbClient, id2, "Team A", "INIT")
	seedTeam(t, dbClient, id3, "Team B", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute without order parameter
	identifiers := []string{id1, id2, id3}
	result, err := queryResolver.TeamByKeysGet(ctx, identifiers, nil)

	// Assertions - should be ordered by identifier ASC (default)
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, id2, result[0].Identifier) // 100e...
	assert.Equal(t, id3, result[1].Identifier) // 200e...
	assert.Equal(t, id1, result[2].Identifier) // 300e...
}

// T050: E2E test for teamByKeysGet deduplication
func TestTeamByKeysGet_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 1 team
	id1 := "100e8400-e29b-41d4-a716-446655440010"
	seedTeam(t, dbClient, id1, "Engineering Team", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Query with duplicate ID
	identifiers := []string{id1, id1}
	result, err := queryResolver.TeamByKeysGet(ctx, identifiers, nil)

	// Assertions - should return team only once
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].Identifier)
}
