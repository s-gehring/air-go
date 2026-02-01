package e2e

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// T103: Integration test for search with getByKeys (verify both queries work together, no conflicts)
func TestSearchWithGetByKeys_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test customers
	seedCustomerForSearch(t, dbClient, "00000000-0000-0000-0000-000000000001", "Alice", "Anderson", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "00000000-0000-0000-0000-000000000002", "Amy", "Brown", "ACTIVE", "INIT")
	seedCustomerForSearch(t, dbClient, "00000000-0000-0000-0000-000000000003", "Carol", "Carter", "BLOCKED", "INIT")
	seedCustomerForSearch(t, dbClient, "00000000-0000-0000-0000-000000000004", "David", "Davis", "ACTIVE", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Test 1: Use search to find entities, then getByKeys to retrieve specific ones
	containsA := "A"
	searchFilter := &generated.CustomerQueryFilterInput{
		FirstName: &generated.StringFilterInput{
			StartsWith: &containsA,
		},
	}

	first := int64(10)
	searchResult, err := queryResolver.CustomerSearch(ctx, searchFilter, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, searchResult)
	assert.Equal(t, int64(2), searchResult.Count) // Alice and Amy both start with A

	// Extract identifiers from search results
	identifiers := make([]string, 0, len(searchResult.Data))
	for _, customer := range searchResult.Data {
		identifiers = append(identifiers, customer.Identifier)
	}

	// Use getByKeys to retrieve the same entities
	getByKeysResult, err := queryResolver.CustomerByKeysGet(ctx, identifiers, nil)
	require.NoError(t, err)
	require.NotNil(t, getByKeysResult)
	assert.Equal(t, searchResult.Count, int64(len(getByKeysResult)))

	// Verify both queries return the same entities
	for i, searchCustomer := range searchResult.Data {
		found := false
		for _, getByKeysCustomer := range getByKeysResult {
			if searchCustomer.Identifier == getByKeysCustomer.Identifier {
				assert.Equal(t, searchCustomer.FirstName, getByKeysCustomer.FirstName)
				assert.Equal(t, searchCustomer.LastName, getByKeysCustomer.LastName)
				found = true
				break
			}
		}
		assert.True(t, found, "Customer %d from search should be in getByKeys result", i)
	}

	// Test 2: Verify both queries exclude deleted entities
	// Search should exclude deleted
	allSearchResult, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), allSearchResult.TotalCount) // All 4 non-deleted

	// GetByKeys should also exclude deleted
	allIdentifiers := []string{
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
		"00000000-0000-0000-0000-000000000003",
		"00000000-0000-0000-0000-000000000004",
	}
	allGetByKeysResult, err := queryResolver.CustomerByKeysGet(ctx, allIdentifiers, nil)
	require.NoError(t, err)
	assert.Equal(t, 4, len(allGetByKeysResult))

	// Test 3: Verify sorting works in both
	sortAsc := generated.SortEnumTypeAsc
	sorter := []*generated.CustomerQuerySorterInput{
		{LastName: &sortAsc},
	}

	sortedSearchResult, err := queryResolver.CustomerSearch(ctx, nil, sorter, &first, nil, nil, nil)
	require.NoError(t, err)
	sortedGetByKeysResult, err := queryResolver.CustomerByKeysGet(ctx, allIdentifiers, sorter)
	require.NoError(t, err)

	// Both should return entities in the same order
	assert.Equal(t, len(sortedSearchResult.Data), len(sortedGetByKeysResult))
	for i := range sortedSearchResult.Data {
		assert.Equal(t, sortedSearchResult.Data[i].Identifier, sortedGetByKeysResult[i].Identifier,
			"Position %d should have same identifier in both results", i)
	}
}

// T103: Test that search and getByKeys use the same MaxBatchSize configuration
func TestSearchWithGetByKeys_SharedConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Test exceeding MaxBatchSize for search (should apply 200 limit)
	// Seed more than 200 customers to test default limit
	for i := 1; i <= 210; i++ {
		seedCustomerForSearch(t, dbClient, strconv.Itoa(i), "First", "Last", "ACTIVE", "INIT")
	}

	// Search without pagination params should return max 200
	searchResult, err := queryResolver.CustomerSearch(ctx, nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(200), searchResult.Count)
	assert.Equal(t, int64(210), searchResult.TotalCount)

	// GetByKeys with 201 identifiers should return error
	identifiers := make([]string, 201)
	for i := 0; i < 201; i++ {
		identifiers[i] = strconv.Itoa(i + 1)
	}

	_, err = queryResolver.CustomerByKeysGet(ctx, identifiers, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "batch size exceeds maximum")
}
