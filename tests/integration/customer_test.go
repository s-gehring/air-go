package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestCustomerRetrieval tests customer retrieval from MongoDB (T013)
func TestCustomerRetrieval(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	require.NoError(t, err, "Failed to start test container")
	defer cleanup()

	// Get customers collection
	db := client.Database("test_db")
	collection := db.Collection("customers")

	t.Run("should retrieve customer with valid identifier", func(t *testing.T) {
		// Arrange - insert test customer
		identifier := "550e8400-e29b-41d4-a716-446655440000"
		firstName := "John"
		lastName := "Doe"
		email := "john.doe@example.com"

		testCustomer := bson.M{
			"identifier":      identifier,
			"firstName":       firstName,
			"lastName":        lastName,
			"userEmail":       email,
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "INIT",
			},
			"createDate": time.Now().Format(time.RFC3339),
		}

		_, err := collection.InsertOne(ctx, testCustomer)
		require.NoError(t, err, "Failed to insert test customer")

		// Act - query customer by identifier
		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		var result generated.Customer
		err = collection.FindOne(ctx, filter).Decode(&result)

		// Assert
		assert.NoError(t, err, "Should successfully retrieve customer")
		assert.Equal(t, identifier, result.Identifier)
		assert.NotNil(t, result.FirstName)
		assert.Equal(t, firstName, *result.FirstName)
		assert.NotNil(t, result.LastName)
		assert.Equal(t, lastName, *result.LastName)
	})
}

// TestDeletedCustomerFiltering tests that deleted customers are excluded (T014)
func TestDeletedCustomerFiltering(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	require.NoError(t, err, "Failed to start test container")
	defer cleanup()

	// Get customers collection
	db := client.Database("test_db")
	collection := db.Collection("customers")

	t.Run("should exclude customers with status.deletion = DELETED", func(t *testing.T) {
		// Arrange - insert deleted customer
		identifier := "deleted-customer-id"

		deletedCustomer := bson.M{
			"identifier":      identifier,
			"firstName":       "Deleted",
			"lastName":        "User",
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "DELETED",
			},
		}

		_, err := collection.InsertOne(ctx, deletedCustomer)
		require.NoError(t, err, "Failed to insert deleted customer")

		// Act - query with deletion status filter
		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		err = collection.FindOne(ctx, filter).Err()

		// Assert - should not find document
		assert.Equal(t, mongo.ErrNoDocuments, err, "Should not find deleted customer")
	})

	t.Run("should include customers with status.deletion = INIT", func(t *testing.T) {
		// Arrange - insert active customer
		identifier := "active-customer-id"

		activeCustomer := bson.M{
			"identifier":      identifier,
			"firstName":       "Active",
			"lastName":        "User",
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "INIT",
			},
		}

		_, err := collection.InsertOne(ctx, activeCustomer)
		require.NoError(t, err, "Failed to insert active customer")

		// Act - query with deletion status filter
		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		var result bson.M
		err = collection.FindOne(ctx, filter).Decode(&result)

		// Assert - should find document
		assert.NoError(t, err, "Should find active customer")
		assert.Equal(t, identifier, result["identifier"])
	})
}

// TestNonExistentCustomer tests handling of non-existent customer (T015)
func TestNonExistentCustomer(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	require.NoError(t, err, "Failed to start test container")
	defer cleanup()

	// Get customers collection
	db := client.Database("test_db")
	collection := db.Collection("customers")

	t.Run("should return ErrNoDocuments for non-existent identifier", func(t *testing.T) {
		// Arrange - non-existent identifier
		identifier := "non-existent-customer-id"

		// Act - query for non-existent customer
		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		err := collection.FindOne(ctx, filter).Err()

		// Assert
		assert.Equal(t, mongo.ErrNoDocuments, err, "Should return ErrNoDocuments for non-existent customer")
	})
}

// TestQueryFilterCorrectness tests the query filter structure (T016)
func TestQueryFilterCorrectness(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	require.NoError(t, err, "Failed to start test container")
	defer cleanup()

	// Get customers collection
	db := client.Database("test_db")
	collection := db.Collection("customers")

	t.Run("filter should include identifier match and deletion exclusion", func(t *testing.T) {
		// Arrange - insert test data
		identifier := "test-filter-customer"

		// Insert both active and deleted customers with same identifier base
		activeCustomer := bson.M{
			"identifier":      identifier + "-active",
			"firstName":       "Active",
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "INIT",
			},
		}

		deletedCustomer := bson.M{
			"identifier":      identifier + "-deleted",
			"firstName":       "Deleted",
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "DELETED",
			},
		}

		_, err := collection.InsertOne(ctx, activeCustomer)
		require.NoError(t, err)
		_, err = collection.InsertOne(ctx, deletedCustomer)
		require.NoError(t, err)

		// Act & Assert - verify filter works correctly
		
		// Should find active customer
		activeFilter := bson.M{
			"identifier":      identifier + "-active",
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		var activeResult bson.M
		err = collection.FindOne(ctx, activeFilter).Decode(&activeResult)
		assert.NoError(t, err, "Should find active customer")
		assert.Equal(t, "Active", activeResult["firstName"])

		// Should not find deleted customer
		deletedFilter := bson.M{
			"identifier":      identifier + "-deleted",
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		err = collection.FindOne(ctx, deletedFilter).Err()
		assert.Equal(t, mongo.ErrNoDocuments, err, "Should not find deleted customer")
	})
}

// TestMissingNestedFields tests handling of missing nested fields (T017)
func TestMissingNestedFields(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	require.NoError(t, err, "Failed to start test container")
	defer cleanup()

	// Get customers collection
	db := client.Database("test_db")
	collection := db.Collection("customers")

	t.Run("should handle Customer with null payment field", func(t *testing.T) {
		// Arrange - insert customer without payment field
		identifier := "customer-no-payment"

		testCustomer := bson.M{
			"identifier":      identifier,
			"firstName":       "NoPayment",
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "INIT",
			},
			// payment field omitted
		}

		_, err := collection.InsertOne(ctx, testCustomer)
		require.NoError(t, err, "Failed to insert test customer")

		// Act
		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		var result generated.Customer
		err = collection.FindOne(ctx, filter).Decode(&result)

		// Assert
		assert.NoError(t, err, "Should successfully decode customer without payment")
		assert.Equal(t, identifier, result.Identifier)
		assert.Nil(t, result.Payment, "Payment should be nil")
	})

	t.Run("should handle Customer with null openBanking field", func(t *testing.T) {
		// Arrange - insert customer without openBanking field
		identifier := "customer-no-openbanking"

		testCustomer := bson.M{
			"identifier":      identifier,
			"firstName":       "NoOpenBanking",
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "INIT",
			},
			// openBanking field omitted
		}

		_, err := collection.InsertOne(ctx, testCustomer)
		require.NoError(t, err, "Failed to insert test customer")

		// Act
		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		var result generated.Customer
		err = collection.FindOne(ctx, filter).Decode(&result)

		// Assert
		assert.NoError(t, err, "Should successfully decode customer without openBanking")
		assert.Equal(t, identifier, result.Identifier)
		assert.Nil(t, result.OpenBanking, "OpenBanking should be nil")
	})

	t.Run("should handle Customer with missing status field", func(t *testing.T) {
		// Arrange - insert customer without status field
		identifier := "customer-no-status"

		testCustomer := bson.M{
			"identifier":      identifier,
			"firstName":       "NoStatus",
			"actionIndicator": "NONE",
			// status field omitted - filter should still include this customer
		}

		_, err := collection.InsertOne(ctx, testCustomer)
		require.NoError(t, err, "Failed to insert test customer")

		// Act - filter with $ne "DELETED" should match documents without status field
		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		var result generated.Customer
		err = collection.FindOne(ctx, filter).Decode(&result)

		// Assert
		assert.NoError(t, err, "Should find customer without status field")
		assert.Equal(t, identifier, result.Identifier)
		assert.Nil(t, result.Status, "Status should be nil")
	})
}

// TestConcurrentQueries tests concurrent customer queries (T046)
func TestConcurrentQueries(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	require.NoError(t, err, "Failed to start test container")
	defer cleanup()

	// Get customers collection
	db := client.Database("test_db")
	collection := db.Collection("customers")

	t.Run("should handle concurrent CustomerGet queries correctly", func(t *testing.T) {
		// Arrange - insert multiple test customers
		for i := 0; i < 10; i++ {
			identifier := fmt.Sprintf("concurrent-customer-%d", i)
			testCustomer := bson.M{
				"identifier":      identifier,
				"firstName":       fmt.Sprintf("Customer%d", i),
				"actionIndicator": "NONE",
				"status": bson.M{
					"deletion": "INIT",
				},
			}

			_, err := collection.InsertOne(ctx, testCustomer)
			require.NoError(t, err)
		}

		// Act - query customers concurrently
		done := make(chan bool)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				identifier := fmt.Sprintf("concurrent-customer-%d", id)
				filter := bson.M{
					"identifier":      identifier,
					"status.deletion": bson.M{"$ne": "DELETED"},
				}

				var result generated.Customer
				err := collection.FindOne(ctx, filter).Decode(&result)
				if err != nil {
					errors <- err
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		close(errors)

		// Assert - no errors
		errorCount := 0
		for err := range errors {
			t.Errorf("Concurrent query error: %v", err)
			errorCount++
		}

		assert.Equal(t, 0, errorCount, "No concurrent query errors expected")
	})
}

// TestQueryTimeoutHandling tests timeout error handling (T047)
func TestQueryTimeoutHandling(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	require.NoError(t, err, "Failed to start test container")
	defer cleanup()

	// Get customers collection
	db := client.Database("test_db")
	collection := db.Collection("customers")

	t.Run("should handle timeout gracefully", func(t *testing.T) {
		// Arrange - create a context with very short timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		// Ensure context expires
		time.Sleep(1 * time.Millisecond)

		// Act - attempt query with expired context
		filter := bson.M{
			"identifier":      "any-id",
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		err := collection.FindOne(timeoutCtx, filter).Err()

		// Assert - should get context deadline exceeded or similar error
		assert.Error(t, err, "Should return error for timeout")
		// The error will be context.DeadlineExceeded or mongo error wrapping it
		t.Logf("Timeout error: %v", err)
	})

	t.Run("should complete within reasonable timeout", func(t *testing.T) {
		// Arrange - insert test customer
		identifier := "timeout-test-customer"
		testCustomer := bson.M{
			"identifier":      identifier,
			"firstName":       "Timeout",
			"actionIndicator": "NONE",
			"status": bson.M{
				"deletion": "INIT",
			},
		}

		_, err := collection.InsertOne(ctx, testCustomer)
		require.NoError(t, err)

		// Act - query with 5 second timeout (should be plenty)
		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		filter := bson.M{
			"identifier":      identifier,
			"status.deletion": bson.M{"$ne": "DELETED"},
		}

		var result generated.Customer
		err = collection.FindOne(timeoutCtx, filter).Decode(&result)

		// Assert
		assert.NoError(t, err, "Query should complete within 5s timeout")
		assert.Equal(t, identifier, result.Identifier)
	})
}
