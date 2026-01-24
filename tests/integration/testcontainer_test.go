package integration

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// TestContainerStartup tests that MongoDB container starts automatically (T036)
func TestContainerStartup(t *testing.T) {
	ctx := context.Background()

	// Start test container
	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	// Verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Errorf("Failed to ping MongoDB: %v", err)
	}

	// Verify database exists
	db := client.Database("test_db")
	if db == nil {
		t.Error("Database() returned nil")
	}

	// Verify we can list databases (proves server is responding)
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		t.Errorf("Failed to list databases: %v", err)
	}

	if len(databases) == 0 {
		t.Error("Expected at least one database, got none")
	}
}

// TestDatabaseCleanup tests that database is dropped and recreated between tests (T037)
func TestDatabaseCleanup(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	db := client.Database("test_db")

	// Insert test data
	collection := db.Collection("test_collection")
	_, err = collection.InsertOne(ctx, bson.M{"test": "data", "timestamp": time.Now()})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Verify data exists
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 document, got %d", count)
	}

	// Clean up database
	err = CleanupTestDatabase(ctx, client, "test_db")
	if err != nil {
		t.Fatalf("Failed to cleanup database: %v", err)
	}

	// Verify data is gone
	count, err = collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Failed to count documents after cleanup: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 documents after cleanup, got %d", count)
	}
}

// TestTestIsolation tests that no data contamination occurs across test runs (T038)
func TestTestIsolation(t *testing.T) {
	ctx := context.Background()

	// First test run
	t.Run("first_run", func(t *testing.T) {
		client, cleanup, err := StartTestContainer(ctx)
		if err != nil {
			t.Fatalf("Failed to start test container: %v", err)
		}
		defer cleanup()

		db := client.Database("test_db")
		collection := db.Collection("isolation_test")

		// Insert data with unique identifier
		_, err = collection.InsertOne(ctx, bson.M{
			"run":       "first",
			"timestamp": time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		// Verify only our data exists
		count, err := collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Fatalf("Failed to count documents: %v", err)
		}
		if count != 1 {
			t.Errorf("First run: expected 1 document, got %d", count)
		}
	})

	// Second test run - should not see data from first run
	t.Run("second_run", func(t *testing.T) {
		client, cleanup, err := StartTestContainer(ctx)
		if err != nil {
			t.Fatalf("Failed to start test container: %v", err)
		}
		defer cleanup()

		db := client.Database("test_db")
		collection := db.Collection("isolation_test")

		// Should start with empty collection
		count, err := collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Fatalf("Failed to count documents: %v", err)
		}
		if count != 0 {
			t.Errorf("Second run: expected 0 documents (clean state), got %d - data contamination detected!", count)
		}

		// Insert new data
		_, err = collection.InsertOne(ctx, bson.M{
			"run":       "second",
			"timestamp": time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}

		// Verify only our data exists
		count, err = collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Fatalf("Failed to count documents: %v", err)
		}
		if count != 1 {
			t.Errorf("Second run: expected 1 document, got %d", count)
		}
	})
}

// TestContainerLifecycle tests container cleanup on test completion (T039)
func TestContainerLifecycle(t *testing.T) {
	ctx := context.Background()

	var containerID string

	// Start container
	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}

	// Verify container is running
	err = client.Ping(ctx, nil)
	if err != nil {
		cleanup()
		t.Fatalf("Container not responding: %v", err)
	}

	// Store container info for verification
	// Note: In actual implementation, we'd need a way to get container ID
	// For now, we trust that cleanup will be called

	// Cleanup container
	cleanup()

	// Verify connection is closed (ping should fail)
	ctx2, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = client.Ping(ctx2, nil)
	if err == nil {
		t.Error("Expected ping to fail after cleanup, but it succeeded")
	}

	// Note: Full container lifecycle verification would check that the
	// Docker container is actually stopped and removed, but that requires
	// Docker API access. The key test is that cleanup() is called and
	// the client connection is terminated.

	_ = containerID // Prevent unused variable error
}

// TestCleanupPerformance tests that cleanup completes in <2s (T045 validation)
func TestCleanupPerformance(t *testing.T) {
	ctx := context.Background()

	client, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	db := client.Database("test_db")

	// Insert some test data (moderate amount)
	collection := db.Collection("perf_test")
	docs := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		docs[i] = bson.M{
			"index":     i,
			"data":      "test data for performance testing",
			"timestamp": time.Now(),
		}
	}
	_, err = collection.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Measure cleanup time
	start := time.Now()
	err = CleanupTestDatabase(ctx, client, "test_db")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify cleanup time is under 2 seconds (SC-002)
	if duration > 2*time.Second {
		t.Errorf("Cleanup took %v, expected <2s (SC-002 violation)", duration)
	}

	t.Logf("Cleanup completed in %v", duration)
}
