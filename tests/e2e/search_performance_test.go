package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// T098: Performance testing with 10,000 entity dataset (verify <1s response time per SC-002)
func TestSearchPerformance_10KEntities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	t.Log("Seeding 10,000 test customers...")
	startSeed := time.Now()

	// Seed 10,000 customers in batches for faster insertion
	batchSize := 1000
	for batch := 0; batch < 10; batch++ {
		for i := 0; i < batchSize; i++ {
			customerNum := batch*batchSize + i + 1
			identifier := fmt.Sprintf("10k-perf-%05d-0000-0000-0000-000000000000", customerNum)
			firstName := fmt.Sprintf("First%d", customerNum%100) // 100 different first names for variety
			lastName := fmt.Sprintf("Last%d", customerNum%500)    // 500 different last names
			seedCustomerForSearch(t, dbClient, identifier, firstName, lastName, "ACTIVE", "INIT")
		}
		if (batch+1)%2 == 0 {
			t.Logf("  Seeded %d customers...", (batch+1)*batchSize)
		}
	}

	seedDuration := time.Since(startSeed)
	t.Logf("Seeding completed in %v", seedDuration)

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Test 1: Search without filters (worst case - all 10,000 entities match)
	t.Run("NoFilter_FullScan", func(t *testing.T) {
		first := int64(200) // Default max batch size

		start := time.Now()
		result, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, int64(200), result.Count)
		assert.Equal(t, int64(10000), result.TotalCount)

		t.Logf("No filter search: %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(1000), "Should complete in <1s (SC-002)")
	})

	// Test 2: Search with filter (selective - returns ~100 matches)
	t.Run("WithFilter_SelectiveMatch", func(t *testing.T) {
		contains := "First1"
		filter := &generated.CustomerQueryFilterInput{
			FirstName: &generated.StringFilterInput{
				Contains: &contains,
			},
		}
		first := int64(200)

		start := time.Now()
		result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Greater(t, result.TotalCount, int64(0))
		assert.LessOrEqual(t, result.Count, int64(200))

		t.Logf("Filtered search (%d results): %v", result.TotalCount, duration)
		assert.Less(t, duration.Milliseconds(), int64(1000), "Should complete in <1s (SC-002)")
	})

	// Test 3: Search with sorting
	t.Run("WithSorting", func(t *testing.T) {
		sortAsc := generated.SortEnumTypeAsc
		sorter := []*generated.CustomerQuerySorterInput{
			{LastName: &sortAsc},
		}
		first := int64(100)

		start := time.Now()
		result, err := queryResolver.CustomerSearch(ctx, nil, sorter, &first, nil, nil, nil)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, int64(100), result.Count)
		assert.Equal(t, int64(10000), result.TotalCount)

		t.Logf("Sorted search: %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(1000), "Should complete in <1s (SC-002)")
	})

	// Test 4: Paginated search (second page)
	t.Run("PaginationSecondPage", func(t *testing.T) {
		// Get first page
		first := int64(100)
		page1, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, page1.Paging.EndCursor)

		// Get second page
		start := time.Now()
		page2, err := queryResolver.CustomerSearch(ctx, nil, nil, &first, page1.Paging.EndCursor, nil, nil)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, int64(100), page2.Count)

		t.Logf("Paginated search (page 2): %v", duration)
		assert.Less(t, duration.Milliseconds(), int64(1000), "Should complete in <1s (SC-002)")
	})

	// Test 5: Complex filter with AND/OR
	t.Run("ComplexFilterWithAndOr", func(t *testing.T) {
		contains1 := "First1"
		contains2 := "First2"
		filter := &generated.CustomerQueryFilterInput{
			Or: []*generated.CustomerQueryFilterInput{
				{FirstName: &generated.StringFilterInput{Contains: &contains1}},
				{FirstName: &generated.StringFilterInput{Contains: &contains2}},
			},
		}
		first := int64(200)

		start := time.Now()
		result, err := queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Greater(t, result.TotalCount, int64(0))

		t.Logf("Complex filter search (%d results): %v", result.TotalCount, duration)
		assert.Less(t, duration.Milliseconds(), int64(1000), "Should complete in <1s (SC-002)")
	})
}

// T099: Performance test with 100,000 entity dataset (if optimization needed)
func TestSearchPerformance_100KEntities(t *testing.T) {
	t.Skip("Skipping 100K performance test - takes too long. Run manually if needed.")

	// This test would follow the same pattern but with 100,000 entities
	// Recommended to run manually with proper MongoDB indexing in place
	// Expected: Still <1s response time with proper indexes on commonly filtered/sorted fields
}

// Performance benchmark for measuring throughput
func BenchmarkCustomerSearch_NoFilter(b *testing.B) {
	ctx := context.Background()
	dbClient := setupTestDatabase(&testing.T{})

	// Seed a moderate dataset
	for i := 0; i < 1000; i++ {
		identifier := fmt.Sprintf("bench-%04d-0000-0000-0000-000000000000", i+1)
		seedCustomerForSearch(&testing.T{}, dbClient, identifier, "First", "Last", "ACTIVE", "INIT")
	}

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()
	first := int64(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = queryResolver.CustomerSearch(ctx, nil, nil, &first, nil, nil, nil)
	}
}

func BenchmarkCustomerSearch_WithFilter(b *testing.B) {
	ctx := context.Background()
	dbClient := setupTestDatabase(&testing.T{})

	// Seed a moderate dataset
	for i := 0; i < 1000; i++ {
		identifier := fmt.Sprintf("bench-%04d-0000-0000-0000-000000000000", i+1)
		firstName := fmt.Sprintf("First%d", i%10)
		seedCustomerForSearch(&testing.T{}, dbClient, identifier, firstName, "Last", "ACTIVE", "INIT")
	}

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()
	first := int64(100)
	contains := "First1"
	filter := &generated.CustomerQueryFilterInput{
		FirstName: &generated.StringFilterInput{Contains: &contains},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = queryResolver.CustomerSearch(ctx, filter, nil, &first, nil, nil, nil)
	}
}
