package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/db"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"go.mongodb.org/mongo-driver/bson"
)

// T010: E2E test for customerSearch basic filtering (firstName contains filter)
func TestCustomerSearch_BasicFiltering_FirstName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers
	seedCustomerForSearch(t, dbClient, "customer-001", "John", "Doe", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-002", "John", "Smith", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-003", "Jane", "Doe", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-004", "Robert", "Brown", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: firstName contains "John"
	containsJohn := "John"
	filter := &generated.CustomerQueryFilterInput{
		FirstName: &generated.StringFilterInput{
			Contains: &containsJohn,
		},
	}

	// Execute customerSearch query
	first := 10
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return exactly 2 customers with "John" in firstName
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Data, 2)

	// Verify both results contain "John"
	for _, customer := range result.Data {
		assert.Contains(t, *customer.FirstName, "John")
	}
}

// T013: E2E test for customerSearch status filtering (activation status filter)
func TestCustomerSearch_StatusFiltering_Activation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers with different activation statuses
	seedCustomerForSearch(t, dbClient, "customer-010", "Alice", "Active", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-011", "Bob", "Blocked", "BLOCKED", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-012", "Carol", "Active", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-013", "Dave", "Suspended", "SUSPENDED", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: status.activation eq ACTIVE
	activeStatus := generated.CustomerActivationStatusActive
	filter := &generated.CustomerQueryFilterInput{
		Status: &generated.CustomerStatusObjectFilterInput{
			Activation: &generated.CustomerActivationStatusFilterInput{
				Eq: &activeStatus,
			},
		},
	}

	// Execute customerSearch query
	first := 10
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return exactly 2 customers with ACTIVE status
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Data, 2)

	// Verify all results have ACTIVE status
	for _, customer := range result.Data {
		assert.Equal(t, generated.CustomerActivationStatusActive, customer.Status.Activation)
	}
}

// T014: E2E test for empty result set (no matches)
func TestCustomerSearch_EmptyResultSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers
	seedCustomerForSearch(t, dbClient, "customer-020", "John", "Doe", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-021", "Jane", "Smith", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: firstName contains "Nonexistent"
	nonexistent := "Nonexistent"
	filter := &generated.CustomerQueryFilterInput{
		FirstName: &generated.StringFilterInput{
			Contains: &nonexistent,
		},
	}

	// Execute customerSearch query
	first := 10
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return empty results
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, 0, result.TotalCount)
	assert.Empty(t, result.Data)
	assert.False(t, result.Paging.HasNextPage)
	assert.False(t, result.Paging.HasPreviousPage)
	assert.Nil(t, result.Paging.StartCursor)
	assert.Nil(t, result.Paging.EndCursor)
}

// T083: E2E test for empty filter criteria (no where param, returns all non-deleted entities)
func TestCustomerSearch_EmptyFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers (mix of deleted and non-deleted)
	seedCustomerForSearch(t, dbClient, "customer-030", "John", "Doe", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-031", "Jane", "Smith", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-032", "Bob", "Brown", "ACTIVE", "DELETED") // Should be excluded
	seedCustomerForSearch(t, dbClient, "customer-033", "Alice", "Green", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute customerSearch query with no filter (nil)
	first := 10
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return 3 non-deleted customers (excludes customer-032)
	assert.Equal(t, 3, result.Count)
	assert.Equal(t, 3, result.TotalCount)
	assert.Len(t, result.Data, 3)

	// Verify no deleted customers in results
	for _, customer := range result.Data {
		assert.NotEqual(t, "DELETED", customer.Status.Deletion)
		assert.NotEqual(t, "customer-032", customer.Identifier)
	}
}

// T084: E2E test for invalid cursor (malformed cursor returns INVALID_INPUT error)
func TestCustomerSearch_InvalidCursor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers
	seedCustomerForSearch(t, dbClient, "customer-040", "John", "Doe", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute customerSearch query with invalid cursor
	first := 10
	invalidCursor := "not-a-valid-base64-cursor"
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, &invalidCursor, nil, nil)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid")
}

// T085: E2E test for conflicting pagination params (both first and last returns error)
func TestCustomerSearch_ConflictingPaginationParams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers
	seedCustomerForSearch(t, dbClient, "customer-050", "John", "Doe", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute customerSearch query with both first and last
	first := 10
	last := 5
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, &last, nil)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "first")
	assert.Contains(t, err.Error(), "last")
}

// T086: E2E test for null value filters (employeeEmail eq null finds entities with null)
func TestCustomerSearch_NullValueFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers with and without employeeEmail
	seedCustomerWithEmployeeEmail(t, dbClient, "customer-060", "John", "Doe", "ACTIVE", "INIT", nil)
	seedCustomerWithEmployeeEmail(t, dbClient, "customer-061", "Jane", "Smith", "ACTIVE", "INIT", strPtr("employee@company.com"))
	seedCustomerWithEmployeeEmail(t, dbClient, "customer-062", "Bob", "Brown", "ACTIVE", "INIT", nil)

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: employeeEmail eq null
	filter := &generated.CustomerQueryFilterInput{
		EmployeeEmail: &generated.StringFilterInput{
			Eq: nil, // null value
		},
	}

	// Execute customerSearch query
	first := 10
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return 2 customers with null employeeEmail
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Data, 2)

	// Verify all results have null employeeEmail
	for _, customer := range result.Data {
		assert.Nil(t, customer.EmployeeEmail)
	}
}

// T087: E2E test for very large result set without pagination (applies 200 default limit)
func TestCustomerSearch_DefaultLimitApplied(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 250 test customers to exceed default limit
	for i := 0; i < 250; i++ {
		identifier := "customer-" + string(rune(70+i))
		seedCustomerForSearch(t, dbClient, identifier, "John", "Doe", "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute customerSearch query without pagination params
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, nil, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return maximum of 200 customers (default limit)
	assert.Equal(t, 200, result.Count)
	assert.Equal(t, 250, result.TotalCount)
	assert.Len(t, result.Data, 200)
	assert.True(t, result.Paging.HasNextPage) // More results available
}

// T088: E2E test for cursor beyond dataset (returns empty results with appropriate hasNext/hasPrevious)
func TestCustomerSearch_CursorBeyondDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers
	seedCustomerForSearch(t, dbClient, "customer-080", "John", "Doe", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "customer-081", "Jane", "Smith", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Get first page to obtain cursor
	first := 10
	result1, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	require.False(t, result1.Paging.HasNextPage) // No more pages

	// Try to fetch next page with cursor (should return empty)
	if result1.Paging.EndCursor != nil {
		result2, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, result1.Paging.EndCursor, nil, nil)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result2)
		assert.Equal(t, 0, result2.Count)
		assert.False(t, result2.Paging.HasNextPage)
	}
}

// Helper: Seed customer for search tests
func seedCustomerForSearch(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, activationStatus, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("customers")
	doc := bson.M{
		"identifier":      identifier,
		"firstName":       firstName,
		"lastName":        lastName,
		"createDate":      time.Now().Format(time.RFC3339),
		"status": bson.M{
			"activation": activationStatus,
			"deletion":   deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}

// Helper: Seed customer with employeeEmail for null filter tests
func seedCustomerWithEmployeeEmail(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, activationStatus, deletionStatus string, employeeEmail *string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("customers")
	doc := bson.M{
		"identifier":      identifier,
		"firstName":       firstName,
		"lastName":        lastName,
		"createDate":      time.Now().Format(time.RFC3339),
		"status": bson.M{
			"activation": activationStatus,
			"deletion":   deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	if employeeEmail != nil {
		doc["employeeEmail"] = *employeeEmail
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}

// Helper: String pointer utility
func strPtr(s string) *string {
	return &s
}
