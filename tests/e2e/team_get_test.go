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

// T022: E2E test for teamGet query success case
func TestTeamGet_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test team
	teamID := "dd0e8400-e29b-41d4-a716-446655440000"
	seedTeam(t, dbClient, teamID, "Engineering Team", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query
	result, err := queryResolver.TeamGet(ctx, teamID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, teamID, result.Identifier)
	assert.Equal(t, "Engineering Team", *result.Name)
}

// T023: E2E test for teamGet query not found case
func TestTeamGet_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query with non-existent UUID
	nonExistentID := "ee0e8400-e29b-41d4-a716-446655440000"
	result, err := queryResolver.TeamGet(ctx, nonExistentID)

	// Assertions: should return nil, not error
	require.NoError(t, err)
	assert.Nil(t, result)
}

// Helper: Seed team data
func seedTeam(t *testing.T, dbClient *db.Client, identifier, name, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("teams")

	doc := bson.M{
		"identifier":  identifier,
		"name":        name,
		"description": "Test team description",
		"createDate":  time.Now().Format(time.RFC3339),
		"status": bson.M{
			"deletion": deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}
