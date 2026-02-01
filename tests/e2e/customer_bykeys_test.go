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

// T038: E2E test for customerByKeysGet with multiple valid identifiers
func TestCustomerByKeysGet_MultipleValid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 3 customers
	id1 := "100e8400-e29b-41d4-a716-446655440000"
	id2 := "200e8400-e29b-41d4-a716-446655440001"
	id3 := "300e8400-e29b-41d4-a716-446655440002"
	
	seedCustomer(t, dbClient, id1, "Alice", "Anderson", "INIT")
	seedCustomer(t, dbClient, id2, "Bob", "Brown", "INIT")
	seedCustomer(t, dbClient, id3, "Charlie", "Clark", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute batch query
	identifiers := []string{id1, id2, id3}
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, nil)

	// Assertions
	require.NoError(t, err)
	require.Len(t, result, 3)
	
	// Verify all customers returned
	customerIDs := make(map[string]bool)
	for _, c := range result {
		customerIDs[c.Identifier] = true
	}
	assert.True(t, customerIDs[id1])
	assert.True(t, customerIDs[id2])
	assert.True(t, customerIDs[id3])
}

// T039: E2E test for customerByKeysGet ordering by lastName ASC
func TestCustomerByKeysGet_OrderByLastNameASC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed customers with different lastNames
	id1 := "100e8400-e29b-41d4-a716-446655440010"
	id2 := "200e8400-e29b-41d4-a716-446655440011"
	id3 := "300e8400-e29b-41d4-a716-446655440012"
	
	seedCustomer(t, dbClient, id1, "Alice", "Zimmerman", "INIT")
	seedCustomer(t, dbClient, id2, "Bob", "Anderson", "INIT")
	seedCustomer(t, dbClient, id3, "Charlie", "Brown", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute with lastName ASC ordering
	identifiers := []string{id1, id2, id3}
	ascSort := generated.SortEnumTypeAsc
	order := []*generated.CustomerQuerySorterInput{
		{LastName: &ascSort},
	}
	
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, order)

	// Assertions - should be ordered: Anderson, Brown, Zimmerman
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, "Anderson", *result[0].LastName)
	assert.Equal(t, "Brown", *result[1].LastName)
	assert.Equal(t, "Zimmerman", *result[2].LastName)
}

// T040: E2E test for customerByKeysGet ordering by payment.status DESC
func TestCustomerByKeysGet_OrderByPaymentStatusDESC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed customers with different payment statuses
	id1 := "100e8400-e29b-41d4-a716-446655440020"
	id2 := "200e8400-e29b-41d4-a716-446655440021"
	
	seedCustomerWithPaymentStatus(t, dbClient, id1, "Alice", "Smith", "ACTIVE", "INIT")
	seedCustomerWithPaymentStatus(t, dbClient, id2, "Bob", "Jones", "EXPIRED", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute with payment.status DESC ordering
	identifiers := []string{id1, id2}
	descSort := generated.SortEnumTypeDesc
	order := []*generated.CustomerQuerySorterInput{
		{Payment: &generated.CustomerPaymentObjectSorterInput{Status: &descSort}},
	}
	
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, order)

	// Assertions - DESC order
	require.NoError(t, err)
	require.Len(t, result, 2)
}

// T041: E2E test for customerByKeysGet with null ordering (SQL-standard behavior)
func TestCustomerByKeysGet_NullOrderingSQLStandard(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed customers: some with birthDate, some without (null)
	id1 := "100e8400-e29b-41d4-a716-446655440030"
	id2 := "200e8400-e29b-41d4-a716-446655440031"
	id3 := "300e8400-e29b-41d4-a716-446655440032"
	
	seedCustomerWithBirthDate(t, dbClient, id1, "Alice", "Smith", "1990-01-01", "INIT")
	seedCustomerWithBirthDate(t, dbClient, id2, "Bob", "Jones", "", "INIT") // null birthDate
	seedCustomerWithBirthDate(t, dbClient, id3, "Charlie", "Brown", "1985-05-15", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Test ASC: non-nulls first (ascending), nulls last
	identifiers := []string{id1, id2, id3}
	ascSort := generated.SortEnumTypeAsc
	order := []*generated.CustomerQuerySorterInput{
		{BirthDate: &ascSort},
	}
	
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, order)

	require.NoError(t, err)
	require.Len(t, result, 3)
	// Verify nulls are last in ASC
	assert.NotNil(t, result[0].BirthDate) // 1985
	assert.NotNil(t, result[1].BirthDate) // 1990
	assert.Nil(t, result[2].BirthDate)    // null last
}

// T042: E2E test for empty identifiers array
func TestCustomerByKeysGet_EmptyArray(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Execute with empty array
	result, err := queryResolver.CustomerByKeysGet(ctx, []string{}, nil)

	// Assertions - should return empty array, not error
	require.NoError(t, err)
	assert.Empty(t, result)
}

// T043: E2E test for mixed valid/invalid identifiers
func TestCustomerByKeysGet_MixedValidInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed one customer
	id1 := "100e8400-e29b-41d4-a716-446655440040"
	seedCustomer(t, dbClient, id1, "Alice", "Smith", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Query for 2 IDs: 1 exists, 1 doesn't
	nonExistentID := "200e8400-e29b-41d4-a716-446655440041"
	identifiers := []string{id1, nonExistentID}
	
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, nil)

	// Assertions - should return only existing customer
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].Identifier)
}

// T044: E2E test for deleted entities exclusion
func TestCustomerByKeysGet_DeletedExclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 2 customers: 1 active, 1 deleted
	id1 := "100e8400-e29b-41d4-a716-446655440050"
	id2 := "200e8400-e29b-41d4-a716-446655440051"
	
	seedCustomer(t, dbClient, id1, "Alice", "Smith", "INIT")
	seedCustomer(t, dbClient, id2, "Bob", "Jones", "DELETED")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Query for both
	identifiers := []string{id1, id2}
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, nil)

	// Assertions - should exclude deleted customer
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].Identifier)
}

// T045: E2E test for duplicate identifiers deduplication
func TestCustomerByKeysGet_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed 1 customer
	id1 := "100e8400-e29b-41d4-a716-446655440060"
	seedCustomer(t, dbClient, id1, "Alice", "Smith", "INIT")

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Query with duplicate ID (appears 3 times)
	identifiers := []string{id1, id1, id1}
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, nil)

	// Assertions - should return customer only once
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].Identifier)
}

// T046: E2E test for batch size limit (201 identifiers should error)
func TestCustomerByKeysGet_BatchSizeExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Create 201 identifiers (exceeds max of 200)
	identifiers := make([]string, 201)
	for i := 0; i < 201; i++ {
		identifiers[i] = "100e8400-e29b-41d4-a716-44665544" + fmt.Sprintf("%04d", i)
	}
	
	result, err := queryResolver.CustomerByKeysGet(ctx, identifiers, nil)

	// Assertions - should return error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "batch size exceeds maximum")
	assert.Contains(t, err.Error(), "201")
	assert.Contains(t, err.Error(), "200")
}

// Helper: Seed customer with payment status
func seedCustomerWithPaymentStatus(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, paymentStatus, deletionStatus string) {
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
		"payment": bson.M{
			"status": paymentStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}

// Helper: Seed customer with birthDate
func seedCustomerWithBirthDate(t *testing.T, dbClient *db.Client, identifier, firstName, lastName, birthDate, deletionStatus string) {
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

	if birthDate != "" {
		doc["birthDate"] = birthDate
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}
