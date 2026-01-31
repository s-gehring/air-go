package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

// InventoryTestData represents test inventory data
type InventoryTestData struct {
	Identifier     string
	CustomerID     *string
	Key            string
	ActionIndicator string
	IsConsistent   bool
	IsComplete     bool
	CreateDate     time.Time
	CreatedByUser  string
}

// SeedInventories seeds test inventory data into MongoDB
func SeedInventories(t *testing.T, db *mongo.Database, inventories []InventoryTestData) {
	t.Helper()
	ctx := context.Background()

	collection := db.Collection("inventories")

	for _, inv := range inventories {
		doc := map[string]interface{}{
			"identifier":      inv.Identifier,
			"key":             inv.Key,
			"actionIndicator": inv.ActionIndicator,
			"isConsistent":    inv.IsConsistent,
			"isComplete":      inv.IsComplete,
			"createDate":      inv.CreateDate,
			"createdByUser":   inv.CreatedByUser,
		}

		if inv.CustomerID != nil {
			doc["customerId"] = *inv.CustomerID
		}

		_, err := collection.InsertOne(ctx, doc)
		require.NoError(t, err, "Failed to seed inventory: %s", inv.Identifier)
	}

	t.Logf("Seeded %d test inventories", len(inventories))
}

// CreateTestInventory creates a single test inventory with sensible defaults
func CreateTestInventory(identifier, key string, deleted bool) InventoryTestData {
	actionIndicator := "NONE"
	if deleted {
		actionIndicator = "DELETE"
	}

	return InventoryTestData{
		Identifier:      identifier,
		Key:             key,
		ActionIndicator: actionIndicator,
		IsConsistent:    true,
		IsComplete:      true,
		CreateDate:      time.Now(),
		CreatedByUser:   "test-user",
	}
}

// CreateTestInventoryWithCustomer creates a test inventory with customerId
func CreateTestInventoryWithCustomer(identifier, key string, customerID *string, deleted bool) InventoryTestData {
	inv := CreateTestInventory(identifier, key, deleted)
	inv.CustomerID = customerID
	return inv
}

// GenerateUUID generates a new UUID string for testing
func GenerateUUID() string {
	return uuid.New().String()
}

// StringPtr returns a pointer to the given string (helper for optional fields)
func StringPtr(s string) *string {
	return &s
}
