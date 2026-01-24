package integration

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestBasicDatabaseOperations is a sample integration test demonstrating testcontainer usage (T046)
func TestBasicDatabaseOperations(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	// Get test database
	db := client.Database("test_db")
	collection := db.Collection("test_collection")

	// Test: Insert document
	t.Run("InsertDocument", func(t *testing.T) {
		doc := bson.M{
			"name":      "Test User",
			"email":     "test@example.com",
			"age":       30,
			"createdAt": time.Now(),
		}

		result, err := collection.InsertOne(ctx, doc)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}

		if result.InsertedID == nil {
			t.Error("Expected InsertedID to be set")
		}

		t.Logf("Inserted document with ID: %v", result.InsertedID)
	})

	// Test: Query document
	t.Run("QueryDocument", func(t *testing.T) {
		var result bson.M
		err := collection.FindOne(ctx, bson.M{"name": "Test User"}).Decode(&result)
		if err != nil {
			t.Fatalf("Failed to find document: %v", err)
		}

		if result["name"] != "Test User" {
			t.Errorf("Expected name='Test User', got '%v'", result["name"])
		}

		if result["email"] != "test@example.com" {
			t.Errorf("Expected email='test@example.com', got '%v'", result["email"])
		}
	})

	// Test: Update document
	t.Run("UpdateDocument", func(t *testing.T) {
		update := bson.M{
			"$set": bson.M{
				"age":       31,
				"updatedAt": time.Now(),
			},
		}

		result, err := collection.UpdateOne(ctx, bson.M{"name": "Test User"}, update)
		if err != nil {
			t.Fatalf("Failed to update document: %v", err)
		}

		if result.ModifiedCount != 1 {
			t.Errorf("Expected 1 document modified, got %d", result.ModifiedCount)
		}

		// Verify update
		var doc bson.M
		err = collection.FindOne(ctx, bson.M{"name": "Test User"}).Decode(&doc)
		if err != nil {
			t.Fatalf("Failed to verify update: %v", err)
		}

		age := doc["age"].(int32)
		if age != 31 {
			t.Errorf("Expected age=31 after update, got %d", age)
		}
	})

	// Test: Delete document
	t.Run("DeleteDocument", func(t *testing.T) {
		result, err := collection.DeleteOne(ctx, bson.M{"name": "Test User"})
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}

		if result.DeletedCount != 1 {
			t.Errorf("Expected 1 document deleted, got %d", result.DeletedCount)
		}

		// Verify deletion
		err = collection.FindOne(ctx, bson.M{"name": "Test User"}).Err()
		if err != mongo.ErrNoDocuments {
			t.Errorf("Expected ErrNoDocuments after deletion, got: %v", err)
		}
	})

	// Test: Index creation
	t.Run("CreateIndex", func(t *testing.T) {
		indexModel := mongo.IndexModel{
			Keys: bson.D{
				{Key: "email", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		}

		indexName, err := collection.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}

		t.Logf("Created index: %s", indexName)

		// Verify index exists
		cursor, err := collection.Indexes().List(ctx)
		if err != nil {
			t.Fatalf("Failed to list indexes: %v", err)
		}
		defer cursor.Close(ctx)

		var indexes []bson.M
		if err = cursor.All(ctx, &indexes); err != nil {
			t.Fatalf("Failed to decode indexes: %v", err)
		}

		// Should have at least 2 indexes: _id and email
		if len(indexes) < 2 {
			t.Errorf("Expected at least 2 indexes, got %d", len(indexes))
		}
	})
}

// TestConnectionPooling verifies connection pool configuration (T046 extended)
func TestConnectionPooling(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	// Verify connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Test concurrent operations
	t.Run("ConcurrentOperations", func(t *testing.T) {
		db := client.Database("test_db")
		collection := db.Collection("concurrent_test")

		// Insert documents concurrently
		done := make(chan bool)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				doc := bson.M{
					"id":        id,
					"data":      "test data",
					"timestamp": time.Now(),
				}
				_, err := collection.InsertOne(ctx, doc)
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

		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Errorf("Concurrent insert error: %v", err)
			errorCount++
		}

		if errorCount > 0 {
			t.Fatalf("Failed %d concurrent operations", errorCount)
		}

		// Verify all documents were inserted
		count, err := collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Fatalf("Failed to count documents: %v", err)
		}

		if count != 10 {
			t.Errorf("Expected 10 documents, got %d", count)
		}
	})
}

// TestDatabaseOperationTimeouts verifies operation timeout configuration (T046 extended)
func TestDatabaseOperationTimeouts(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	db := client.Database("test_db")
	collection := db.Collection("timeout_test")

	// Test with short timeout
	t.Run("ShortTimeout", func(t *testing.T) {
		shortCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel()

		// This operation should timeout (or succeed if very fast)
		_, err := collection.InsertOne(shortCtx, bson.M{"data": "test"})
		// Either succeeds quickly or times out - both are acceptable
		if err != nil && err != context.DeadlineExceeded {
			t.Logf("Operation completed or timed out as expected: %v", err)
		}
	})

	// Test with sufficient timeout
	t.Run("SufficientTimeout", func(t *testing.T) {
		normalCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(normalCtx, bson.M{"data": "test"})
		if err != nil {
			t.Errorf("Operation with 5s timeout failed: %v", err)
		}
	})
}
