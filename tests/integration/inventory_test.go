package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/tests/testutil"
)

// T012: Test batch MongoDB query with $in operator
func TestInventoryBatchQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	db := testutil.SetupTestDB(t, mongoURI)
	defer testutil.TeardownTestDB(t, db)

	testutil.CreateIndexes(t, db)
	testutil.CleanCollection(t, db, "inventories")

	ctx := context.Background()

	// Seed test data
	id1 := testutil.GenerateUUID()
	id2 := testutil.GenerateUUID()
	id3 := testutil.GenerateUUID()
	id4 := testutil.GenerateUUID() // Won't query for this one

	inventories := []testutil.InventoryTestData{
		testutil.CreateTestInventory(id1, "INV-001", false),
		testutil.CreateTestInventory(id2, "INV-002", false),
		testutil.CreateTestInventory(id3, "INV-003", false),
		testutil.CreateTestInventory(id4, "INV-004", false),
	}
	testutil.SeedInventories(t, db, inventories)

	// Query for 3 out of 4 inventories
	collection := db.Collection("inventories")
	filter := map[string]interface{}{
		"identifier": map[string]interface{}{
			"$in": []string{id1, id2, id3},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	require.NoError(t, err)

	var results []map[string]interface{}
	err = cursor.All(ctx, &results)
	require.NoError(t, err)

	assert.Len(t, results, 3, "Should return exactly 3 inventories")
}

// T013: Test deletion status filtering
func TestInventoryDeletionFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	db := testutil.SetupTestDB(t, mongoURI)
	defer testutil.TeardownTestDB(t, db)

	testutil.CreateIndexes(t, db)
	testutil.CleanCollection(t, db, "inventories")

	ctx := context.Background()

	// Seed test data: 2 active, 1 deleted
	id1 := testutil.GenerateUUID()
	id2 := testutil.GenerateUUID()
	id3 := testutil.GenerateUUID()

	inventories := []testutil.InventoryTestData{
		testutil.CreateTestInventory(id1, "INV-001", false), // Active
		testutil.CreateTestInventory(id2, "INV-002", true),  // Deleted
		testutil.CreateTestInventory(id3, "INV-003", false), // Active
	}
	testutil.SeedInventories(t, db, inventories)

	// Query with deletion filter
	collection := db.Collection("inventories")
	filter := map[string]interface{}{
		"identifier": map[string]interface{}{
			"$in": []string{id1, id2, id3},
		},
		"actionIndicator": map[string]interface{}{
			"$ne": "DELETE",
		},
	}

	cursor, err := collection.Find(ctx, filter)
	require.NoError(t, err)

	var results []map[string]interface{}
	err = cursor.All(ctx, &results)
	require.NoError(t, err)

	assert.Len(t, results, 2, "Should return only 2 non-deleted inventories")

	// Verify deleted inventory is not in results
	for _, result := range results {
		assert.NotEqual(t, id2, result["identifier"], "Deleted inventory should be excluded")
	}
}

// T033: Test ordering by customerId ASC with MongoDB
func TestInventoryOrderingByCustomerIdASC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	db := testutil.SetupTestDB(t, mongoURI)
	defer testutil.TeardownTestDB(t, db)

	testutil.CreateIndexes(t, db)
	testutil.CleanCollection(t, db, "inventories")

	ctx := context.Background()

	// Seed test data with different customerIds (sorted order: cust1, cust2, cust3)
	id1 := testutil.GenerateUUID()
	id2 := testutil.GenerateUUID()
	id3 := testutil.GenerateUUID()

	cust3 := "750e8400-e29b-41d4-a716-446655440003" // Highest UUID
	cust1 := "550e8400-e29b-41d4-a716-446655440001" // Lowest UUID
	cust2 := "650e8400-e29b-41d4-a716-446655440002" // Middle UUID

	inventories := []testutil.InventoryTestData{
		testutil.CreateTestInventoryWithCustomer(id1, "INV-001", &cust3, false), // Should be 3rd
		testutil.CreateTestInventoryWithCustomer(id2, "INV-002", &cust1, false), // Should be 1st
		testutil.CreateTestInventoryWithCustomer(id3, "INV-003", &cust2, false), // Should be 2nd
	}
	testutil.SeedInventories(t, db, inventories)

	// Build aggregation pipeline with ASC ordering
	collection := db.Collection("inventories")
	pipeline := []map[string]interface{}{
		{
			"$match": map[string]interface{}{
				"identifier": map[string]interface{}{
					"$in": []string{id1, id2, id3},
				},
			},
		},
		{
			"$addFields": map[string]interface{}{
				"_sortKey": map[string]interface{}{
					"$cond": map[string]interface{}{
						"if":   map[string]interface{}{"$eq": []interface{}{"$customerId", nil}},
						"then": "zzzzzzz-null-placeholder",
						"else": "$customerId",
					},
				},
			},
		},
		{
			"$sort": map[string]interface{}{
				"_sortKey": 1, // ASC
			},
		},
		{
			"$project": map[string]interface{}{
				"_sortKey": 0,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	var results []map[string]interface{}
	err = cursor.All(ctx, &results)
	require.NoError(t, err)

	assert.Len(t, results, 3, "Should return 3 inventories")

	// Verify ASC ordering: cust1, cust2, cust3
	assert.Equal(t, id2, results[0]["identifier"], "First should be inventory with cust1")
	assert.Equal(t, id3, results[1]["identifier"], "Second should be inventory with cust2")
	assert.Equal(t, id1, results[2]["identifier"], "Third should be inventory with cust3")
}

// T034: Test ordering by customerId DESC with MongoDB
func TestInventoryOrderingByCustomerIdDESC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	db := testutil.SetupTestDB(t, mongoURI)
	defer testutil.TeardownTestDB(t, db)

	testutil.CreateIndexes(t, db)
	testutil.CleanCollection(t, db, "inventories")

	ctx := context.Background()

	// Seed test data with different customerIds
	id1 := testutil.GenerateUUID()
	id2 := testutil.GenerateUUID()
	id3 := testutil.GenerateUUID()

	cust3 := "750e8400-e29b-41d4-a716-446655440003" // Highest UUID
	cust1 := "550e8400-e29b-41d4-a716-446655440001" // Lowest UUID
	cust2 := "650e8400-e29b-41d4-a716-446655440002" // Middle UUID

	inventories := []testutil.InventoryTestData{
		testutil.CreateTestInventoryWithCustomer(id1, "INV-001", &cust1, false), // Should be 3rd
		testutil.CreateTestInventoryWithCustomer(id2, "INV-002", &cust3, false), // Should be 1st
		testutil.CreateTestInventoryWithCustomer(id3, "INV-003", &cust2, false), // Should be 2nd
	}
	testutil.SeedInventories(t, db, inventories)

	// Build aggregation pipeline with DESC ordering
	collection := db.Collection("inventories")
	pipeline := []map[string]interface{}{
		{
			"$match": map[string]interface{}{
				"identifier": map[string]interface{}{
					"$in": []string{id1, id2, id3},
				},
			},
		},
		{
			"$addFields": map[string]interface{}{
				"_sortKey": map[string]interface{}{
					"$cond": map[string]interface{}{
						"if":   map[string]interface{}{"$eq": []interface{}{"$customerId", nil}},
						"then": "0000000-null-placeholder",
						"else": "$customerId",
					},
				},
			},
		},
		{
			"$sort": map[string]interface{}{
				"_sortKey": -1, // DESC
			},
		},
		{
			"$project": map[string]interface{}{
				"_sortKey": 0,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	var results []map[string]interface{}
	err = cursor.All(ctx, &results)
	require.NoError(t, err)

	assert.Len(t, results, 3, "Should return 3 inventories")

	// Verify DESC ordering: cust3, cust2, cust1
	assert.Equal(t, id2, results[0]["identifier"], "First should be inventory with cust3")
	assert.Equal(t, id3, results[1]["identifier"], "Second should be inventory with cust2")
	assert.Equal(t, id1, results[2]["identifier"], "Third should be inventory with cust1")
}

// T035: Test null customerId handling with ASC ordering (nulls last)
func TestInventoryNullCustomerIdASC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	db := testutil.SetupTestDB(t, mongoURI)
	defer testutil.TeardownTestDB(t, db)

	testutil.CreateIndexes(t, db)
	testutil.CleanCollection(t, db, "inventories")

	ctx := context.Background()

	// Seed test data: 2 with customerIds, 1 with null customerId
	id1 := testutil.GenerateUUID()
	id2 := testutil.GenerateUUID()
	id3 := testutil.GenerateUUID()

	cust1 := "550e8400-e29b-41d4-a716-446655440001"
	cust2 := "650e8400-e29b-41d4-a716-446655440002"

	inventories := []testutil.InventoryTestData{
		testutil.CreateTestInventoryWithCustomer(id1, "INV-001", nil, false),    // Null - should be last
		testutil.CreateTestInventoryWithCustomer(id2, "INV-002", &cust2, false), // Should be 2nd
		testutil.CreateTestInventoryWithCustomer(id3, "INV-003", &cust1, false), // Should be 1st
	}
	testutil.SeedInventories(t, db, inventories)

	// Build aggregation pipeline with ASC ordering (nulls last)
	collection := db.Collection("inventories")
	pipeline := []map[string]interface{}{
		{
			"$match": map[string]interface{}{
				"identifier": map[string]interface{}{
					"$in": []string{id1, id2, id3},
				},
			},
		},
		{
			"$addFields": map[string]interface{}{
				"_sortKey": map[string]interface{}{
					"$ifNull": []interface{}{
						"$customerId",
						"zzzzzzz-null-placeholder", // Sorts after all UUIDs
					},
				},
			},
		},
		{
			"$sort": map[string]interface{}{
				"_sortKey": 1, // ASC
			},
		},
		{
			"$project": map[string]interface{}{
				"_sortKey": 0,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	var results []map[string]interface{}
	err = cursor.All(ctx, &results)
	require.NoError(t, err)

	assert.Len(t, results, 3, "Should return 3 inventories")

	// Verify ASC ordering with null last: cust1, cust2, null
	assert.Equal(t, id3, results[0]["identifier"], "First should be inventory with cust1")
	assert.Equal(t, id2, results[1]["identifier"], "Second should be inventory with cust2")
	assert.Equal(t, id1, results[2]["identifier"], "Third should be inventory with null customerId")
}

// T036: Test null customerId handling with DESC ordering (nulls first)
func TestInventoryNullCustomerIdDESC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	db := testutil.SetupTestDB(t, mongoURI)
	defer testutil.TeardownTestDB(t, db)

	testutil.CreateIndexes(t, db)
	testutil.CleanCollection(t, db, "inventories")

	ctx := context.Background()

	// Seed test data: 2 with customerIds, 1 with null customerId
	id1 := testutil.GenerateUUID()
	id2 := testutil.GenerateUUID()
	id3 := testutil.GenerateUUID()

	cust1 := "550e8400-e29b-41d4-a716-446655440001"
	cust2 := "650e8400-e29b-41d4-a716-446655440002"

	inventories := []testutil.InventoryTestData{
		testutil.CreateTestInventoryWithCustomer(id1, "INV-001", &cust1, false), // Should be 3rd
		testutil.CreateTestInventoryWithCustomer(id2, "INV-002", nil, false),    // Null - should be 1st
		testutil.CreateTestInventoryWithCustomer(id3, "INV-003", &cust2, false), // Should be 2nd
	}
	testutil.SeedInventories(t, db, inventories)

	// Build aggregation pipeline with DESC ordering (nulls first)
	collection := db.Collection("inventories")
	pipeline := []map[string]interface{}{
		{
			"$match": map[string]interface{}{
				"identifier": map[string]interface{}{
					"$in": []string{id1, id2, id3},
				},
			},
		},
		{
			"$addFields": map[string]interface{}{
				"_sortKey": map[string]interface{}{
					"$ifNull": []interface{}{
						"$customerId",
						"zzzzzzz-null-placeholder", // Sorts first when descending
					},
				},
			},
		},
		{
			"$sort": map[string]interface{}{
				"_sortKey": -1, // DESC
			},
		},
		{
			"$project": map[string]interface{}{
				"_sortKey": 0,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	var results []map[string]interface{}
	err = cursor.All(ctx, &results)
	require.NoError(t, err)

	assert.Len(t, results, 3, "Should return 3 inventories")

	// Verify DESC ordering with null first: null, cust2, cust1
	assert.Equal(t, id2, results[0]["identifier"], "First should be inventory with null customerId")
	assert.Equal(t, id3, results[1]["identifier"], "Second should be inventory with cust2")
	assert.Equal(t, id1, results[2]["identifier"], "Third should be inventory with cust1")
}
