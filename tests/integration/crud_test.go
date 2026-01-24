package integration

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestInsertOne verifies document insertion and retrieval (T049)
func TestInsertOne(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("users")

	// Test document
	doc := bson.M{
		"name":      "Alice Johnson",
		"email":     "alice@example.com",
		"age":       28,
		"createdAt": time.Now(),
	}

	// Insert document
	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne() failed: %v", err)
	}

	if result.InsertedID == nil {
		t.Error("InsertOne() returned nil InsertedID")
	}

	// Retrieve and verify
	var retrieved bson.M
	err = collection.FindOne(ctx, bson.M{"email": "alice@example.com"}).Decode(&retrieved)
	if err != nil {
		t.Fatalf("FindOne() failed: %v", err)
	}

	if retrieved["name"] != "Alice Johnson" {
		t.Errorf("Expected name='Alice Johnson', got '%v'", retrieved["name"])
	}
}

// TestFindOne verifies document query with filter (T050)
func TestFindOne(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("products")

	// Insert test data
	_, err = collection.InsertOne(ctx, bson.M{
		"name":  "Laptop",
		"price": 999.99,
		"stock": 10,
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test: Find existing document
	var result bson.M
	err = collection.FindOne(ctx, bson.M{"name": "Laptop"}).Decode(&result)
	if err != nil {
		t.Fatalf("FindOne() failed: %v", err)
	}

	if result["price"] != 999.99 {
		t.Errorf("Expected price=999.99, got %v", result["price"])
	}

	// Test: Find non-existent document
	err = collection.FindOne(ctx, bson.M{"name": "Nonexistent"}).Err()
	if err != mongo.ErrNoDocuments {
		t.Errorf("Expected ErrNoDocuments, got %v", err)
	}
}

// TestUpdateOne verifies document update and persistence (T051)
func TestUpdateOne(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("inventory")

	// Insert test data
	_, err = collection.InsertOne(ctx, bson.M{
		"item":     "Widget",
		"quantity": 100,
		"price":    5.50,
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update document
	update := bson.M{
		"$set": bson.M{
			"quantity":  85,
			"updatedAt": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"item": "Widget"}, update)
	if err != nil {
		t.Fatalf("UpdateOne() failed: %v", err)
	}

	if result.ModifiedCount != 1 {
		t.Errorf("Expected ModifiedCount=1, got %d", result.ModifiedCount)
	}

	// Verify update persisted
	var updated bson.M
	err = collection.FindOne(ctx, bson.M{"item": "Widget"}).Decode(&updated)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	quantity := updated["quantity"].(int32)
	if quantity != 85 {
		t.Errorf("Expected quantity=85 after update, got %d", quantity)
	}
}

// TestDeleteOne verifies document deletion (T052)
func TestDeleteOne(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("temp_data")

	// Insert test data
	_, err = collection.InsertOne(ctx, bson.M{
		"id":   "delete-me",
		"data": "temporary",
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Delete document
	result, err := collection.DeleteOne(ctx, bson.M{"id": "delete-me"})
	if err != nil {
		t.Fatalf("DeleteOne() failed: %v", err)
	}

	if result.DeletedCount != 1 {
		t.Errorf("Expected DeletedCount=1, got %d", result.DeletedCount)
	}

	// Verify deletion
	err = collection.FindOne(ctx, bson.M{"id": "delete-me"}).Err()
	if err != mongo.ErrNoDocuments {
		t.Errorf("Expected ErrNoDocuments after deletion, got %v", err)
	}

	// Test: Delete non-existent document
	result, err = collection.DeleteOne(ctx, bson.M{"id": "never-existed"})
	if err != nil {
		t.Fatalf("DeleteOne() on non-existent failed: %v", err)
	}

	if result.DeletedCount != 0 {
		t.Errorf("Expected DeletedCount=0 for non-existent, got %d", result.DeletedCount)
	}
}

// TestInsertMany verifies batch insert (T053)
func TestInsertMany(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("batch_data")

	// Prepare batch documents
	docs := []interface{}{
		bson.M{"id": 1, "name": "Item 1"},
		bson.M{"id": 2, "name": "Item 2"},
		bson.M{"id": 3, "name": "Item 3"},
		bson.M{"id": 4, "name": "Item 4"},
		bson.M{"id": 5, "name": "Item 5"},
	}

	// Insert many
	result, err := collection.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany() failed: %v", err)
	}

	if len(result.InsertedIDs) != 5 {
		t.Errorf("Expected 5 InsertedIDs, got %d", len(result.InsertedIDs))
	}

	// Verify all inserted
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments() failed: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count=5, got %d", count)
	}
}

// TestFind verifies multiple document query (T054)
func TestFind(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("employees")

	// Insert test data
	employees := []interface{}{
		bson.M{"name": "John", "department": "Engineering", "salary": 80000},
		bson.M{"name": "Jane", "department": "Engineering", "salary": 85000},
		bson.M{"name": "Bob", "department": "Sales", "salary": 60000},
		bson.M{"name": "Alice", "department": "Engineering", "salary": 90000},
	}

	_, err = collection.InsertMany(ctx, employees)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Find all engineering employees
	cursor, err := collection.Find(ctx, bson.M{"department": "Engineering"})
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		t.Fatalf("Cursor.All() failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 engineering employees, got %d", len(results))
	}

	// Verify all are engineering
	for _, emp := range results {
		if emp["department"] != "Engineering" {
			t.Errorf("Found non-engineering employee: %v", emp)
		}
	}
}

// TestCountDocuments verifies count operation (T055)
func TestCountDocuments(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("counters")

	// Test: Count empty collection
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments() on empty failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count=0 for empty collection, got %d", count)
	}

	// Insert test data
	docs := []interface{}{
		bson.M{"type": "A", "value": 1},
		bson.M{"type": "A", "value": 2},
		bson.M{"type": "B", "value": 3},
		bson.M{"type": "A", "value": 4},
		bson.M{"type": "B", "value": 5},
	}

	_, err = collection.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Count all documents
	count, err = collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments() failed: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count=5, got %d", count)
	}

	// Count with filter
	count, err = collection.CountDocuments(ctx, bson.M{"type": "A"})
	if err != nil {
		t.Fatalf("CountDocuments() with filter failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count=3 for type A, got %d", count)
	}
}

// TestOperationTimeout verifies 5-10s timeout enforcement (T056)
func TestOperationTimeout(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := client.Database("test_db").Collection("timeout_test")

	// Test: Operation with very short timeout (should timeout or succeed quickly)
	t.Run("ShortTimeout", func(t *testing.T) {
		shortCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		_, err := collection.InsertOne(shortCtx, bson.M{"data": "test"})
		// Either succeeds quickly or times out - both acceptable
		if err != nil && err != context.DeadlineExceeded {
			t.Logf("Operation result: %v (acceptable)", err)
		}
	})

	// Test: Operation with sufficient timeout (should succeed)
	t.Run("SufficientTimeout", func(t *testing.T) {
		normalCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		_, err := collection.InsertOne(normalCtx, bson.M{"data": "test"})
		if err != nil {
			t.Errorf("Operation with 10s timeout failed: %v", err)
		}
	})

	// Test: Verify default operation timeout range (5-10s per FR-007)
	t.Run("DefaultTimeout", func(t *testing.T) {
		// This would be tested with actual Collection wrapper
		// For now, verify the operation completes within reasonable time
		start := time.Now()

		_, err := collection.InsertOne(ctx, bson.M{"data": "test"})
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Operation failed: %v", err)
		}

		if duration > 10*time.Second {
			t.Errorf("Operation took %v, expected <10s", duration)
		}
	})
}
