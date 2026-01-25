package testutil

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test UUIDs for consistent testing
const (
	TestCustomerID1        = "550e8400-e29b-41d4-a716-446655440000"
	TestCustomerID2        = "660e8400-e29b-41d4-a716-446655440001"
	TestEmployeeID1        = "770e8400-e29b-41d4-a716-446655440002"
	TestEmployeeID2        = "880e8400-e29b-41d4-a716-446655440003"
	TestPortfolioID1       = "990e8400-e29b-41d4-a716-446655440004"
	TestPortfolioID2       = "aa0e8400-e29b-41d4-a716-446655440005"
	TestInventoryID1       = "bb0e8400-e29b-41d4-a716-446655440006"
	TestExecutionPlanID1   = "cc0e8400-e29b-41d4-a716-446655440007"
	TestTeamID1            = "dd0e8400-e29b-41d4-a716-446655440008"
	TestAttachmentID1      = "ee0e8400-e29b-41d4-a716-446655440009"
)

// CustomerFixture represents a test customer
type CustomerFixture struct {
	Identifier  string
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	DateOfBirth string
	Address     AddressFixture
}

// AddressFixture represents a test address
type AddressFixture struct {
	Street     string
	City       string
	PostalCode string
	Country    string
}

// NewCustomerFixture creates a test customer
func NewCustomerFixture(identifier, firstName, lastName string) *CustomerFixture {
	return &CustomerFixture{
		Identifier:  identifier,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       firstName + "." + lastName + "@test.com",
		Phone:       "+49 123 456789",
		DateOfBirth: "1985-06-15",
		Address: AddressFixture{
			Street:     "Hauptstraße 123",
			City:       "Berlin",
			PostalCode: "10115",
			Country:    "Germany",
		},
	}
}

// EmployeeFixture represents a test employee
type EmployeeFixture struct {
	Identifier string
	FirstName  string
	LastName   string
	Email      string
	Role       string
	TeamID     *string
}

// NewEmployeeFixture creates a test employee
func NewEmployeeFixture(identifier, firstName, lastName, role string) *EmployeeFixture {
	return &EmployeeFixture{
		Identifier: identifier,
		FirstName:  firstName,
		LastName:   lastName,
		Email:      firstName + "." + lastName + "@company.com",
		Role:       role,
	}
}

// ReferencePortfolioFixture represents a test reference portfolio
type ReferencePortfolioFixture struct {
	Identifier string
	CustomerID string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewReferencePortfolioFixture creates a test reference portfolio
func NewReferencePortfolioFixture(identifier, customerID string) *ReferencePortfolioFixture {
	now := time.Now()
	return &ReferencePortfolioFixture{
		Identifier: identifier,
		CustomerID: customerID,
		Status:     "ACTIVE",
		CreatedAt:  now.Add(-24 * time.Hour),
		UpdatedAt:  now,
	}
}

// InventoryFixture represents a test inventory
type InventoryFixture struct {
	Identifier string
	CustomerID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewInventoryFixture creates a test inventory
func NewInventoryFixture(identifier, customerID string) *InventoryFixture {
	now := time.Now()
	return &InventoryFixture{
		Identifier: identifier,
		CustomerID: customerID,
		CreatedAt:  now.Add(-24 * time.Hour),
		UpdatedAt:  now,
	}
}

// ExecutionPlanFixture represents a test execution plan
type ExecutionPlanFixture struct {
	Identifier string
	CustomerID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewExecutionPlanFixture creates a test execution plan
func NewExecutionPlanFixture(identifier, customerID string) *ExecutionPlanFixture {
	now := time.Now()
	return &ExecutionPlanFixture{
		Identifier: identifier,
		CustomerID: customerID,
		CreatedAt:  now.Add(-24 * time.Hour),
		UpdatedAt:  now,
	}
}

// TeamFixture represents a test team
type TeamFixture struct {
	Identifier string
	Name       string
	LeaderID   string
}

// NewTeamFixture creates a test team
func NewTeamFixture(identifier, name, leaderID string) *TeamFixture {
	return &TeamFixture{
		Identifier: identifier,
		Name:       name,
		LeaderID:   leaderID,
	}
}

// AttachmentFixture represents a test attachment
type AttachmentFixture struct {
	Identifier  string
	Filename    string
	ContentType string
	Size        int64
	UploadedAt  time.Time
}

// NewAttachmentFixture creates a test attachment
func NewAttachmentFixture(identifier, filename string) *AttachmentFixture {
	return &AttachmentFixture{
		Identifier:  identifier,
		Filename:    filename,
		ContentType: "application/pdf",
		Size:        245632,
		UploadedAt:  time.Now(),
	}
}

// NewObjectID creates a new MongoDB ObjectID for testing
func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// ObjectIDFromHex creates an ObjectID from hex string (panics on error for test simplicity)
func ObjectIDFromHex(hex string) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(hex)
	return id
}
