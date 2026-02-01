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

// T020: E2E test for employeeGet query success case
func TestEmployeeGet_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test employee
	employeeID := "880e8400-e29b-41d4-a716-446655440000"
	seedEmployee(t, dbClient, employeeID, "Alice", "Johnson", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query
	result, err := queryResolver.EmployeeGet(ctx, employeeID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, employeeID, result.Identifier)
	assert.Equal(t, "Alice", *result.FirstName)
	assert.Equal(t, "Johnson", *result.LastName)
}

// T021: E2E test for employeeGet query not found case
func TestEmployeeGet_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query with non-existent UUID
	nonExistentID := "990e8400-e29b-41d4-a716-446655440000"
	result, err := queryResolver.EmployeeGet(ctx, nonExistentID)

	// Assertions: should return nil, not error
	require.NoError(t, err)
	assert.Nil(t, result)
}

// Helper: Seed employee data
func seedEmployee(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("employees")

	doc := bson.M{
		"identifier":  identifier,
		"firstName":   firstName,
		"lastName":    lastName,
		"createDate":  time.Now().Format(time.RFC3339),
		"status": bson.M{
			"deletion": deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}
