package testutil

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MockCollection is a mock implementation of *mongo.Collection
type MockCollection struct {
	mock.Mock
}

// Find mocks the Find method
func (m *MockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

// FindOne mocks the FindOne method
func (m *MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(*mongo.SingleResult)
}

// InsertOne mocks the InsertOne method
func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

// UpdateOne mocks the UpdateOne method
func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// DeleteOne mocks the DeleteOne method
func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

// CountDocuments mocks the CountDocuments method
func (m *MockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

// MockDatabase is a mock implementation of *mongo.Database
type MockDatabase struct {
	mock.Mock
	collections map[string]*MockCollection
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		collections: make(map[string]*MockCollection),
	}
}

// Collection returns a mock collection
func (m *MockDatabase) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	args := m.Called(name, opts)
	if _, exists := m.collections[name]; exists {
		return nil // Return nil but track the mock
	}
	return args.Get(0).(*mongo.Collection)
}

// AddMockCollection adds a mock collection for testing
func (m *MockDatabase) AddMockCollection(name string, collection *MockCollection) {
	m.collections[name] = collection
}

// MockClient is a mock implementation of *mongo.Client
type MockClient struct {
	mock.Mock
}

// Database mocks the Database method
func (m *MockClient) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	args := m.Called(name, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*mongo.Database)
}

// Ping mocks the Ping method
func (m *MockClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	args := m.Called(ctx, rp)
	return args.Error(0)
}

// Disconnect mocks the Disconnect method
func (m *MockClient) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Helper function to create BSON filter from map
func BSONFilter(filter map[string]interface{}) bson.M {
	return bson.M(filter)
}
