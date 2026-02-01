package e2e

import (
	"context"
	"fmt"
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
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return exactly 2 customers with "John" in firstName
	assert.Equal(t, int64(2), result.Count)
	assert.Equal(t, int64(2), result.TotalCount)
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
	activeStatus := generated.UserStatusActive
	filter := &generated.CustomerQueryFilterInput{
		Status: &generated.CustomerStatusObjectFilterInput{
			Activation: &generated.EnumFilterOfNullableOfUserStatusInput{
				Eq: &activeStatus,
			},
		},
	}

	// Execute customerSearch query
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return exactly 2 customers with ACTIVE status
	assert.Equal(t, int64(2), result.Count)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Len(t, result.Data, 2)

	// Verify all results have ACTIVE status
	for _, customer := range result.Data {
		assert.Equal(t, generated.UserStatusActive, *customer.Status.Activation)
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
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return empty results
	assert.Equal(t, int64(0), result.Count)
	assert.Equal(t, int64(0), result.TotalCount)
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
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return 3 non-deleted customers (excludes customer-032)
	assert.Equal(t, int64(3), result.Count)
	assert.Equal(t, int64(3), result.TotalCount)
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
	first := int64(10)
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
	first := int64(10)
	last := int64(5)
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
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return 2 customers with null employeeEmail
	assert.Equal(t, int64(2), result.Count)
	assert.Equal(t, int64(2), result.TotalCount)
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
	assert.Equal(t, int64(200), result.Count)
	assert.Equal(t, int64(250), result.TotalCount)
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
	first := int64(10)
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

// T061: E2E test for complex AND/OR filter (firstName AND (status OR status))
func TestCustomerSearch_ComplexAndOrFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers with different combinations
	seedCustomerForSearch(t, dbClient, "cust-complex-1", "Sarah", "Active1", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-complex-2", "Sarah", "Blocked1", "BLOCKED", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-complex-3", "Sarah", "Init1", "INIT", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-complex-4", "John", "Active2", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-complex-5", "John", "Blocked2", "BLOCKED", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: firstName contains "Sarah" AND (status.activation eq ACTIVE OR BLOCKED)
	searchName := "Sarah"
	statusActive := generated.UserStatusActive
	statusBlocked := generated.UserStatusBlocked
	filter := &generated.CustomerQueryFilterInput{
		And: []*generated.CustomerQueryFilterInput{
			{FirstName: &generated.StringFilterInput{Contains: &searchName}},
			{Or: []*generated.CustomerQueryFilterInput{
				{Status: &generated.CustomerStatusObjectFilterInput{
					Activation: &generated.EnumFilterOfNullableOfUserStatusInput{Eq: &statusActive},
				}},
				{Status: &generated.CustomerStatusObjectFilterInput{
					Activation: &generated.EnumFilterOfNullableOfUserStatusInput{Eq: &statusBlocked},
				}},
			}},
		},
	}

	// Execute customerSearch
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.Count) // Should match Sarah+ACTIVE and Sarah+BLOCKED only
	assert.Len(t, result.Data, 2)

	// Verify all results match the filter criteria
	for _, customer := range result.Data {
		assert.Contains(t, *customer.FirstName, "Sarah")
		assert.True(t, *customer.Status.Activation == generated.UserStatusActive || *customer.Status.Activation == generated.UserStatusBlocked)
	}
}

// T063: E2E test for deeply nested filters (multiple nesting levels)
func TestCustomerSearch_DeeplyNestedFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers
	seedCustomerForSearch(t, dbClient, "cust-nested-1", "Alice", "ActiveAlice", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-nested-2", "Alice", "BlockedAlice", "BLOCKED", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-nested-3", "Bob", "ActiveBob", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-nested-4", "Bob", "BlockedBob", "BLOCKED", "INIT")
	seedCustomerForSearch(t, dbClient, "cust-nested-5", "Charlie", "ActiveCharlie", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build deeply nested filter: (Alice AND ACTIVE) OR (Bob AND BLOCKED)
	nameAlice := "Alice"
	nameBob := "Bob"
	statusActive := generated.UserStatusActive
	statusBlocked := generated.UserStatusBlocked
	filter := &generated.CustomerQueryFilterInput{
		Or: []*generated.CustomerQueryFilterInput{
			{And: []*generated.CustomerQueryFilterInput{
				{FirstName: &generated.StringFilterInput{Eq: &nameAlice}},
				{Status: &generated.CustomerStatusObjectFilterInput{
					Activation: &generated.EnumFilterOfNullableOfUserStatusInput{Eq: &statusActive},
				}},
			}},
			{And: []*generated.CustomerQueryFilterInput{
				{FirstName: &generated.StringFilterInput{Eq: &nameBob}},
				{Status: &generated.CustomerStatusObjectFilterInput{
					Activation: &generated.EnumFilterOfNullableOfUserStatusInput{Eq: &statusBlocked},
				}},
			}},
		},
	}

	// Execute customerSearch
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.Count) // Should match Alice+ACTIVE and Bob+BLOCKED
	assert.Len(t, result.Data, 2)

	// Verify results
	names := []string{*result.Data[0].FirstName, *result.Data[1].FirstName}
	assert.Contains(t, names, "Alice")
	assert.Contains(t, names, "Bob")
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

// T034: E2E test for customerSearch single-field sorting (createDate DESC)
func TestCustomerSearch_Sorting_CreateDateDesc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers with specific createDate values
	now := time.Now()
	seedCustomerWithCreateDate(t, dbClient, "customer-sort-1", "Alice", "First", "ACTIVE", "INIT", now.Add(-3*24*time.Hour))
	seedCustomerWithCreateDate(t, dbClient, "customer-sort-2", "Bob", "Second", "ACTIVE", "INIT", now.Add(-1*24*time.Hour))
	seedCustomerWithCreateDate(t, dbClient, "customer-sort-3", "Carol", "Third", "ACTIVE", "INIT", now.Add(-2*24*time.Hour))

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build sorter: createDate DESC
	sortDesc := generated.SortEnumTypeDesc
	sorter := []*generated.CustomerQuerySorterInput{
		{CreateDate: &sortDesc},
	}

	// Execute customerSearch query
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, nil, sorter, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(3), result.Count)
	assert.Len(t, result.Data, 3)

	// Verify results are sorted by createDate DESC (newest first)
	assert.Equal(t, "Bob", *result.Data[0].FirstName)      // Most recent
	assert.Equal(t, "Carol", *result.Data[1].FirstName)    // Middle
	assert.Equal(t, "Alice", *result.Data[2].FirstName)    // Oldest
}

// T036: E2E test for null value sorting
func TestCustomerSearch_Sorting_NullHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers with some having null employeeEmail
	seedCustomerWithEmployeeEmail(t, dbClient, "customer-null-1", "Alice", "HasEmail", "ACTIVE", "INIT", strPtr("alice@company.com"))
	seedCustomerWithEmployeeEmail(t, dbClient, "customer-null-2", "Bob", "NoEmail", "ACTIVE", "INIT", nil)
	seedCustomerWithEmployeeEmail(t, dbClient, "customer-null-3", "Carol", "HasEmail", "ACTIVE", "INIT", strPtr("carol@company.com"))
	seedCustomerWithEmployeeEmail(t, dbClient, "customer-null-4", "Dave", "NoEmail", "ACTIVE", "INIT", nil)

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build sorter: employeeEmail ASC (non-nulls first, nulls last)
	sortAsc := generated.SortEnumTypeAsc
	sorter := []*generated.CustomerQuerySorterInput{
		{EmployeeEmail: &sortAsc},
	}

	// Execute customerSearch query
	first := int64(10)
	result, err := queryResolver.CustomerSearch(ctx, nil, sorter, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(4), result.Count)
	assert.Len(t, result.Data, 4)

	// Verify ASC sorting: non-nulls first, nulls last
	// First two should have non-null employeeEmail (sorted alphabetically)
	assert.NotNil(t, result.Data[0].EmployeeEmail)
	assert.NotNil(t, result.Data[1].EmployeeEmail)

	// Last two should have null employeeEmail
	assert.Nil(t, result.Data[2].EmployeeEmail)
	assert.Nil(t, result.Data[3].EmployeeEmail)
}

// T045: E2E test for forward pagination (first page)
func TestCustomerSearch_Pagination_ForwardFirstPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 25 test customers (more than typical page size of 20)
	for i := 1; i <= 25; i++ {
		seedCustomerForSearch(t, dbClient, fmt.Sprintf("customer-page-%03d", i), fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i), "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute customerSearch with first: 20
	first := int64(20)
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(20), result.Count)
	assert.Equal(t, int64(25), result.TotalCount)
	assert.True(t, result.Paging.HasNextPage) // Should have more results
	assert.NotNil(t, result.Paging.EndCursor) // Should have cursor for next page
	assert.NotNil(t, result.Paging.StartCursor)
}

// T046: E2E test for forward pagination (next page)
func TestCustomerSearch_Pagination_ForwardNextPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 25 test customers
	for i := 1; i <= 25; i++ {
		seedCustomerForSearch(t, dbClient, fmt.Sprintf("customer-page-%03d", i), fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i), "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Get first page
	first := int64(20)
	result1, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, result1.Paging.EndCursor)

	// Get next page using endCursor from first page
	result2, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, result1.Paging.EndCursor, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, 5, result2.Count) // Remaining 5 items
	assert.Equal(t, 25, result2.TotalCount)
	assert.False(t, result2.Paging.HasNextPage) // No more results
	assert.True(t, result2.Paging.HasPreviousPage) // Has previous page
}

// T047: E2E test for pagination last page
func TestCustomerSearch_Pagination_LastPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed exactly 25 customers
	for i := 1; i <= 25; i++ {
		seedCustomerForSearch(t, dbClient, fmt.Sprintf("customer-last-%03d", i), fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i), "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Get first page (20 items)
	first := int64(20)
	result1, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, result1.Paging.EndCursor)

	// Get last page (remaining 5 items)
	result2, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, result1.Paging.EndCursor, nil, nil)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, 5, result2.Count)
	assert.False(t, result2.Paging.HasNextPage) // This is the last page
}

// T049: E2E test for bidirectional pagination
func TestCustomerSearch_Pagination_Bidirectional(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 30 customers
	for i := 1; i <= 30; i++ {
		seedCustomerForSearch(t, dbClient, fmt.Sprintf("customer-bidir-%03d", i), fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i), "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Navigate forward: page 1
	first := int64(10)
	page1, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 10, page1.Count)

	// Navigate forward: page 2
	page2, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, page1.Paging.EndCursor, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 10, page2.Count)
	assert.True(t, page2.Paging.HasPreviousPage)

	// Navigate backward: back to page 1
	last := int64(10)
	pageBack, err := queryResolver.CustomerSearch(ctx, nil, nil, nil, nil, &last, page2.Paging.StartCursor)
	require.NoError(t, err)
	assert.Equal(t, 10, pageBack.Count)

	// Verify we got back to the same identifiers
	assert.Equal(t, page1.Data[0].Identifier, pageBack.Data[0].Identifier)
}

// Helper: Seed customer with specific createDate
func seedCustomerWithCreateDate(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, activationStatus, deletionStatus string, createDate time.Time) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("customers")
	doc := bson.M{
		"identifier":      identifier,
		"firstName":       firstName,
		"lastName":        lastName,
		"createDate":      createDate.Format(time.RFC3339),
		"status": bson.M{
			"activation": activationStatus,
			"deletion":   deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}

// Helper: String pointer utility
func strPtr(s string) *string {
	return &s
}

// T073: E2E test for count and totalCount with full page (first 20 of 147 entities)
func TestCustomerSearch_CountAndTotalCount_FullPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed exactly 147 customers
	for i := 1; i <= 147; i++ {
		seedCustomerForSearch(t, dbClient, fmt.Sprintf("cust-count-%03d", i), fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i), "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute customerSearch query requesting first 20
	first := int64(20)
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(20), result.Count)      // Current page has 20
	assert.Equal(t, int64(147), result.TotalCount) // Total across all pages is 147
	assert.Len(t, result.Data, 20)
	assert.True(t, result.Paging.HasNextPage) // More pages available
}

// T075: E2E test for count and totalCount with no filters (first 50 of 1000 total)
func TestCustomerSearch_CountAndTotalCount_NoFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed exactly 1000 customers
	for i := 1; i <= 1000; i++ {
		seedCustomerForSearch(t, dbClient, fmt.Sprintf("cust-nofilter-%04d", i), fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i), "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute customerSearch query with no filter, requesting first 50
	first := int64(50)
	result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(50), result.Count)       // Current page has 50
	assert.Equal(t, int64(1000), result.TotalCount) // Total across all pages is 1000
	assert.Len(t, result.Data, 50)
	assert.True(t, result.Paging.HasNextPage) // More pages available
}

// T076: E2E test for consistent totalCount across pages
func TestCustomerSearch_TotalCount_ConsistentAcrossPages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed exactly 150 customers
	for i := 1; i <= 150; i++ {
		seedCustomerForSearch(t, dbClient, fmt.Sprintf("cust-consistent-%03d", i), fmt.Sprintf("First%d", i), fmt.Sprintf("Last%d", i), "ACTIVE", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Get page 1
	first := int64(50)
	page1, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, page1)

	// Get page 2
	page2, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, page1.Paging.EndCursor, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, page2)

	// Get page 3
	page3, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, page2.Paging.EndCursor, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, page3)

	// Assertions: totalCount should be same across all pages
	assert.Equal(t, int64(150), page1.TotalCount)
	assert.Equal(t, int64(150), page2.TotalCount)
	assert.Equal(t, int64(150), page3.TotalCount)

	// But count should reflect actual page sizes
	assert.Equal(t, int64(50), page1.Count)
	assert.Equal(t, int64(50), page2.Count)
	assert.Equal(t, int64(50), page3.Count) // Exactly 150 items, so page 3 has 50
}
