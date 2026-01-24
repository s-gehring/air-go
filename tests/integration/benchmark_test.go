package integration

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/air-go/internal/db"
)

// BenchmarkSimpleDatabaseOperations benchmarks basic operations (T096)
// Success Criteria: SC-004 requires <100ms for simple operations in development
func BenchmarkSimpleDatabaseOperations(b *testing.B) {
	ctx := context.Background()

	// Start test container
	mongoClient, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := mongoClient.Database("bench_db").Collection("bench_collection")

	// Prepare test document
	testDoc := bson.M{
		"name":  "Benchmark Test",
		"value": 42,
	}

	b.Run("InsertOne", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := collection.InsertOne(ctx, testDoc)
			if err != nil {
				b.Fatalf("InsertOne failed: %v", err)
			}
		}
		b.StopTimer()

		// Verify performance: operations should complete in <100ms on average
		avgDuration := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
		if avgDuration > 100*time.Millisecond {
			b.Logf("WARNING: Average insert duration %v exceeds 100ms target (SC-004)", avgDuration)
		}
	})

	b.Run("FindOne", func(b *testing.B) {
		// Insert document for querying
		result, _ := collection.InsertOne(ctx, testDoc)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var doc bson.M
			err := collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&doc)
			if err != nil {
				b.Fatalf("FindOne failed: %v", err)
			}
		}
		b.StopTimer()

		avgDuration := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
		if avgDuration > 100*time.Millisecond {
			b.Logf("WARNING: Average find duration %v exceeds 100ms target (SC-004)", avgDuration)
		}
	})

	b.Run("UpdateOne", func(b *testing.B) {
		// Insert document for updating
		result, _ := collection.InsertOne(ctx, testDoc)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := collection.UpdateOne(ctx,
				bson.M{"_id": result.InsertedID},
				bson.M{"$set": bson.M{"value": i}})
			if err != nil {
				b.Fatalf("UpdateOne failed: %v", err)
			}
		}
		b.StopTimer()

		avgDuration := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
		if avgDuration > 100*time.Millisecond {
			b.Logf("WARNING: Average update duration %v exceeds 100ms target (SC-004)", avgDuration)
		}
	})

	b.Run("DeleteOne", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Insert document to delete
			result, _ := collection.InsertOne(ctx, testDoc)
			b.StartTimer()

			_, err := collection.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
			if err != nil {
				b.Fatalf("DeleteOne failed: %v", err)
			}
		}
		b.StopTimer()

		avgDuration := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
		if avgDuration > 100*time.Millisecond {
			b.Logf("WARNING: Average delete duration %v exceeds 100ms target (SC-004)", avgDuration)
		}
	})
}

// BenchmarkConnectionEstablishment benchmarks connection time (T097)
// Success Criteria: SC-003 requires connection establishment in <30s
func BenchmarkConnectionEstablishment(b *testing.B) {
	logger := zerolog.Nop()

	// Start test container once
	ctx := context.Background()
	_, uri, cleanup, err := StartTestContainerWithURI(ctx)
	if err != nil {
		b.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new client
		config := &db.DBConfig{
			URI:              uri,
			Database:         "bench_db",
			ConnectTimeout:   30 * time.Second,
			OperationTimeout: 10 * time.Second,
			MinPoolSize:      5,
			MaxPoolSize:      10,
			MaxConnIdleTime:  5 * time.Minute,
			MaxRetryAttempts: 3,
			RetryBaseDelay:   1 * time.Second,
			RetryMaxDelay:    10 * time.Second,
		}

		client, err := db.NewClient(config, logger)
		if err != nil {
			b.Fatalf("NewClient failed: %v", err)
		}

		// Measure connection time
		connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		start := time.Now()
		err = client.Connect(connectCtx)
		duration := time.Since(start)
		cancel()

		if err != nil {
			b.Fatalf("Connect failed: %v", err)
		}

		// Disconnect
		disconnectCtx, disconnectCancel := context.WithTimeout(ctx, 10*time.Second)
		_ = client.Disconnect(disconnectCtx)
		disconnectCancel()
		client.Close()

		// Verify performance target (SC-003)
		if duration > 30*time.Second {
			b.Errorf("Connection took %v, exceeds 30s limit (SC-003)", duration)
		}
	}
	b.StopTimer()

	// Report average connection time
	avgDuration := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
	b.Logf("Average connection establishment time: %v (target: <30s)", avgDuration)
}

// BenchmarkTestDatabaseCleanup benchmarks cleanup performance (T098)
// Success Criteria: SC-002 requires test database cleanup in <2s
func BenchmarkTestDatabaseCleanup(b *testing.B) {
	ctx := context.Background()

	// Start test container
	mongoClient, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	dbName := "bench_cleanup_db"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// Populate database with test data
		db := mongoClient.Database(dbName)
		collection := db.Collection("test_collection")

		// Insert 100 documents
		docs := make([]interface{}, 100)
		for j := 0; j < 100; j++ {
			docs[j] = bson.M{"index": j, "data": "test data"}
		}
		_, _ = collection.InsertMany(ctx, docs)

		b.StartTimer()

		// Measure cleanup time
		start := time.Now()
		err := CleanupTestDatabase(ctx, mongoClient, dbName)
		duration := time.Since(start)

		b.StopTimer()

		if err != nil {
			b.Fatalf("CleanupTestDatabase failed: %v", err)
		}

		// Verify performance target (SC-002)
		if duration > 2*time.Second {
			b.Errorf("Cleanup took %v, exceeds 2s limit (SC-002)", duration)
		}
	}
	b.StopTimer()

	// Report average cleanup time
	avgDuration := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
	b.Logf("Average cleanup time: %v (target: <2s)", avgDuration)
}

// BenchmarkOperationLatency measures end-to-end operation latency (T096)
func BenchmarkOperationLatency(b *testing.B) {
	ctx := context.Background()

	mongoClient, cleanup, err := StartTestContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to start test container: %v", err)
	}
	defer cleanup()

	collection := mongoClient.Database("latency_db").Collection("latency_test")

	// Prepare test data
	testDoc := bson.M{
		"name":      "Latency Test",
		"timestamp": time.Now(),
		"data":      make([]byte, 1024), // 1KB document
	}

	b.ResetTimer()
	b.Run("InsertAndFind", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Insert
			result, err := collection.InsertOne(ctx, testDoc)
			if err != nil {
				b.Fatalf("Insert failed: %v", err)
			}

			// Find
			var doc bson.M
			err = collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&doc)
			if err != nil {
				b.Fatalf("Find failed: %v", err)
			}

			// Clean up
			_, _ = collection.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
		}
	})
	b.StopTimer()

	avgDuration := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
	b.Logf("Average insert+find latency: %v", avgDuration)
}
