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

// T011: E2E test for employeeSearch basic filtering (userEmail filter)
func TestEmployeeSearch_BasicFiltering_UserEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test employees
	seedEmployeeForSearch(t, dbClient, "employee-001", "John", "Doe", "john.doe@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-002", "Jane", "Smith", "jane.smith@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-003", "Bob", "Brown", "bob.brown@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-004", "Alice", "Green", "john.alice@company.com", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: userEmail contains "john"
	containsJohn := "john"
	filter := &generated.EmployeeQueryFilterInput{
		UserEmail: &generated.StringFilterInput{
			Contains: &containsJohn,
		},
	}

	// Execute employeeSearch query
	first := int64(10)
	result, err := queryResolver.EmployeeSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return 2 employees with "john" in userEmail
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Data, 2)

	// Verify both results contain "john" in email
	for _, employee := range result.Data {
		assert.Contains(t, *employee.UserEmail, "john")
	}
}

// T033: E2E test for employeeSearch single-field sorting (lastName ASC)
func TestEmployeeSearch_SingleFieldSorting_LastNameASC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test employees with different lastNames
	seedEmployeeForSearch(t, dbClient, "employee-010", "Alice", "Wilson", "alice.wilson@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-011", "Bob", "Smith", "bob.smith@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-012", "Carol", "Anderson", "carol.anderson@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-013", "Dave", "Taylor", "dave.taylor@company.com", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build sorter: lastName ASC
	sortAsc := generated.SortEnumTypeAsc
	sorter := []*generated.EmployeeQuerySorterInput{
		{
			LastName: &sortAsc,
		},
	}

	// Execute employeeSearch query
	first := int64(10)
	result, err := queryResolver.EmployeeSearch(ctx, nil, sorter, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 4, result.Count)
	assert.Len(t, result.Data, 4)

	// Verify results are sorted by lastName ascending
	assert.Equal(t, "Anderson", *result.Data[0].LastName)
	assert.Equal(t, "Smith", *result.Data[1].LastName)
	assert.Equal(t, "Taylor", *result.Data[2].LastName)
	assert.Equal(t, "Wilson", *result.Data[3].LastName)
}

// T048: E2E test for backward pagination (last 10, verify hasPreviousPage)
func TestEmployeeSearch_BackwardPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 25 test employees
	for i := 0; i < 25; i++ {
		identifier := "employee-020-" + string(rune(65+i))
		firstName := "Employee" + string(rune(65+i))
		seedEmployeeForSearch(t, dbClient, identifier, firstName, "LastName", "user@company.com", "INIT")
	}

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute employeeSearch query with last: 10 (backward pagination)
	last := int64(10)
	result, err := queryResolver.EmployeeSearch(ctx, nil, nil, nil, nil, &last, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return last 10 employees
	assert.Equal(t, 10, result.Count)
	assert.Equal(t, 25, result.TotalCount)
	assert.Len(t, result.Data, 10)

	// Should have previous page available
	assert.True(t, result.Paging.HasPreviousPage)
	assert.False(t, result.Paging.HasNextPage) // At the end
	assert.NotNil(t, result.Paging.StartCursor)
	assert.NotNil(t, result.Paging.EndCursor)
}

// T074: E2E test for count and totalCount with partial page (first 20 with only 5 results)
func TestEmployeeSearch_CountWithPartialPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed only 5 test employees
	seedEmployeeForSearch(t, dbClient, "employee-030", "Alice", "One", "alice@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-031", "Bob", "Two", "bob@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-032", "Carol", "Three", "carol@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-033", "Dave", "Four", "dave@company.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "employee-034", "Eve", "Five", "eve@company.com", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute employeeSearch query requesting first 20 (but only 5 exist)
	first := int64(20)
	result, err := queryResolver.EmployeeSearch(ctx, nil, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return only 5 employees
	assert.Equal(t, 5, result.Count)       // Current page count
	assert.Equal(t, 5, result.TotalCount)  // Total matching entities
	assert.Len(t, result.Data, 5)
	assert.False(t, result.Paging.HasNextPage)
	assert.False(t, result.Paging.HasPreviousPage)
}

// T060: E2E test for AND filter combination
func TestEmployeeSearch_ComplexFilter_AndCombination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test employees
	seedEmployeeForSearch(t, dbClient, "emp-and-1", "John", "Smith", "john.smith@test.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "emp-and-2", "John", "Doe", "john.doe@test.com", "INIT")
	seedEmployeeForSearch(t, dbClient, "emp-and-3", "Jane", "Smith", "jane.smith@test.com", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: firstName="John" AND lastName="Smith"
	firstNameJohn := "John"
	lastNameSmith := "Smith"
	filter := &generated.EmployeeQueryFilterInput{
		And: []*generated.EmployeeQueryFilterInput{
			{FirstName: &generated.StringFilterInput{Eq: &firstNameJohn}},
			{LastName: &generated.StringFilterInput{Eq: &lastNameSmith}},
		},
	}

	// Execute employeeSearch
	first := int64(10)
	result, err := queryResolver.EmployeeSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Count) // Only John Smith
	assert.Equal(t, "John", *result.Data[0].FirstName)
	assert.Equal(t, "Smith", *result.Data[0].LastName)
}

// Helper: Seed employee for search tests
func seedEmployeeForSearch(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, userEmail, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("employees")
	doc := bson.M{
		"identifier":      identifier,
		"firstName":       firstName,
		"lastName":        lastName,
		"userEmail":       userEmail,
		"createDate":      time.Now().Format(time.RFC3339),
		"status": bson.M{
			"deletion": deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}
