package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/db"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"go.mongodb.org/mongo-driver/bson"
)

// T015: E2E test for customerGet query success case
func TestCustomerGet_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customer
	customerID := "550e8400-e29b-41d4-a716-446655440000"
	seedCustomer(t, dbClient, customerID, "John", "Doe", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query
	result, err := queryResolver.CustomerGet(ctx, customerID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, customerID, result.Identifier)
	assert.Equal(t, "John", *result.FirstName)
	assert.Equal(t, "Doe", *result.LastName)
}

// T016: E2E test for customerGet query not found case
func TestCustomerGet_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query with non-existent UUID
	nonExistentID := "660e8400-e29b-41d4-a716-446655440000"
	result, err := queryResolver.CustomerGet(ctx, nonExistentID)

	// Assertions: should return nil, not error
	require.NoError(t, err)
	assert.Nil(t, result)
}

// T017: E2E test for customerGet query deleted entity exclusion
func TestCustomerGet_DeletedExclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed deleted customer (status.deletion = "DELETED")
	customerID := "770e8400-e29b-41d4-a716-446655440000"
	seedCustomer(t, dbClient, customerID, "Jane", "Smith", "DELETED")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query
	result, err := queryResolver.CustomerGet(ctx, customerID)

	// Assertions: deleted customer should return nil
	require.NoError(t, err)
	assert.Nil(t, result)
}

// T018: E2E test for customerGet query invalid UUID error
func TestCustomerGet_InvalidUUID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	testCases := []struct {
		name       string
		identifier string
	}{
		{"malformed UUID", "not-a-uuid"},
		{"incomplete UUID", "550e8400-e29b-41d4"},
		{"empty string", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := queryResolver.CustomerGet(ctx, tc.identifier)

			// Should return error with INVALID_INPUT code
			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "invalid UUID format")
		})
	}
}

// T019: E2E test for customerGet query null identifier error
func TestCustomerGet_NullIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute query with empty string (null equivalent in Go)
	result, err := queryResolver.CustomerGet(ctx, "")

	// Should return error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid UUID format")
}

// Helper: Setup test database - returns db.Client which implements resolvers.DBClient
func setupTestDatabase(t *testing.T) *db.Client {
	t.Helper()

	config := &db.DBConfig{
		URI:              "mongodb://localhost:27017",
		Database:         "test_air_go",
		ConnectTimeout:   30 * time.Second,
		OperationTimeout: 10 * time.Second,
		MinPoolSize:      5,
		MaxPoolSize:      10,
		MaxConnIdleTime:  5 * time.Minute,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   1 * time.Second,
		RetryMaxDelay:    10 * time.Second,
	}

	logger := zerolog.Nop()
	dbClient, err := db.NewClient(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = dbClient.Connect(ctx)
	require.NoError(t, err)

	// Verify connection
	require.True(t, dbClient.IsConnected(), "Database should be connected")

	// Clean all entity collections before each test
	collections := []string{"customers", "employees", "teams", "inventories", "executionPlans", "referencePortfolios"}
	for _, collName := range collections {
		collection := dbClient.Collection(collName)
		_, err = collection.DeleteMany(ctx, bson.M{})
		require.NoError(t, err)
	}

	return dbClient
}

// Helper: Teardown test database
func teardownTestDatabase(t *testing.T, dbClient *db.Client) {
	t.Helper()
	ctx := context.Background()
	if dbClient != nil {
		_ = dbClient.Disconnect(ctx)
	}
}

// Helper: Seed customer data
func seedCustomer(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("customers")
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
