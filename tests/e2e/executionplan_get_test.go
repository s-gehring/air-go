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

// T026: E2E test for executionPlanGet query success case
func TestExecutionPlanGet_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test execution plan
	executionPlanID := "cc0e8400-e29b-41d4-a716-446655440001"
	seedExecutionPlan(t, dbClient, executionPlanID, "NONE")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query
	result, err := queryResolver.ExecutionPlanGet(ctx, executionPlanID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, executionPlanID, result.Identifier)
}

// T027: E2E test for executionPlanGet query not found case
func TestExecutionPlanGet_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query with non-existent UUID
	nonExistentID := "dd0e8400-e29b-41d4-a716-446655440002"
	result, err := queryResolver.ExecutionPlanGet(ctx, nonExistentID)

	// Assertions: should return nil, not error
	require.NoError(t, err)
	assert.Nil(t, result)
}

// Helper: Seed execution plan data
func seedExecutionPlan(t *testing.T, dbClient *db.Client, identifier, actionIndicator string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("executionPlans")

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
