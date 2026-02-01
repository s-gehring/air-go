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

// T028: E2E test for referencePortfolioGet query success case
func TestReferencePortfolioGet_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test reference portfolio
	portfolioID := "990e8400-e29b-41d4-a716-446655440001"
	seedReferencePortfolio(t, dbClient, portfolioID, "NONE")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query
	result, err := queryResolver.ReferencePortfolioGet(ctx, portfolioID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, portfolioID, result.Identifier)
}

// T029: E2E test for referencePortfolioGet query not found case
func TestReferencePortfolioGet_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query with non-existent UUID
	nonExistentID := "aa0e8400-e29b-41d4-a716-446655440002"
	result, err := queryResolver.ReferencePortfolioGet(ctx, nonExistentID)

	// Assertions: should return nil, not error
	require.NoError(t, err)
	assert.Nil(t, result)
}

// Helper: Seed reference portfolio data
func seedReferencePortfolio(t *testing.T, dbClient *db.Client, identifier, actionIndicator string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("referencePortfolios")

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
