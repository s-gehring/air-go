package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"go.mongodb.org/mongo-driver/bson"
)

// T064: Unit test for recursive filter conversion (AND/OR recursion)
func TestConvertCustomerFilter_RecursiveAndOr(t *testing.T) {
	t.Run("Simple AND recursion", func(t *testing.T) {
		// Build filter: firstName eq "John" AND lastName eq "Doe"
		firstNameJohn := "John"
		lastNameDoe := "Doe"
		filter := &generated.CustomerQueryFilterInput{
			And: []*generated.CustomerQueryFilterInput{
				{FirstName: &generated.StringFilterInput{Eq: &firstNameJohn}},
				{LastName: &generated.StringFilterInput{Eq: &lastNameDoe}},
			},
		}

		// Convert to MongoDB filter
		result := resolvers.ConvertCustomerFilterForTest(filter)

		// Verify result contains $and with two conditions
		assert.Contains(t, result, "$and")
		andConditions := result["$and"].([]bson.M)
		assert.Len(t, andConditions, 2)

		// Verify conditions
		assert.Contains(t, andConditions[0], "firstName")
		assert.Contains(t, andConditions[1], "lastName")
	})

	t.Run("Simple OR recursion", func(t *testing.T) {
		// Build filter: firstName eq "John" OR firstName eq "Jane"
		firstNameJohn := "John"
		firstNameJane := "Jane"
		filter := &generated.CustomerQueryFilterInput{
			Or: []*generated.CustomerQueryFilterInput{
				{FirstName: &generated.StringFilterInput{Eq: &firstNameJohn}},
				{FirstName: &generated.StringFilterInput{Eq: &firstNameJane}},
			},
		}

		// Convert to MongoDB filter
		result := resolvers.ConvertCustomerFilterForTest(filter)

		// Verify result contains $or with two conditions
		assert.Contains(t, result, "$or")
		orConditions := result["$or"].([]bson.M)
		assert.Len(t, orConditions, 2)
	})

	t.Run("Nested AND/OR recursion", func(t *testing.T) {
		// Build filter: (firstName eq "John" AND lastName eq "Doe") OR (firstName eq "Jane")
		firstNameJohn := "John"
		lastNameDoe := "Doe"
		firstNameJane := "Jane"
		filter := &generated.CustomerQueryFilterInput{
			Or: []*generated.CustomerQueryFilterInput{
				{And: []*generated.CustomerQueryFilterInput{
					{FirstName: &generated.StringFilterInput{Eq: &firstNameJohn}},
					{LastName: &generated.StringFilterInput{Eq: &lastNameDoe}},
				}},
				{FirstName: &generated.StringFilterInput{Eq: &firstNameJane}},
			},
		}

		// Convert to MongoDB filter
		result := resolvers.ConvertCustomerFilterForTest(filter)

		// Verify result contains $or at top level
		assert.Contains(t, result, "$or")
		orConditions := result["$or"].([]bson.M)
		assert.Len(t, orConditions, 2)

		// Verify first OR condition contains nested $and
		assert.Contains(t, orConditions[0], "$and")
		nestedAndConditions := orConditions[0]["$and"].([]bson.M)
		assert.Len(t, nestedAndConditions, 2)

		// Verify second OR condition is simple
		assert.Contains(t, orConditions[1], "firstName")
	})

	t.Run("Deeply nested recursion", func(t *testing.T) {
		// Build filter: ((A AND B) OR (C AND D)) AND E
		nameA := "A"
		nameB := "B"
		nameC := "C"
		nameD := "D"
		nameE := "E"
		filter := &generated.CustomerQueryFilterInput{
			And: []*generated.CustomerQueryFilterInput{
				{Or: []*generated.CustomerQueryFilterInput{
					{And: []*generated.CustomerQueryFilterInput{
						{FirstName: &generated.StringFilterInput{Eq: &nameA}},
						{LastName: &generated.StringFilterInput{Eq: &nameB}},
					}},
					{And: []*generated.CustomerQueryFilterInput{
						{FirstName: &generated.StringFilterInput{Eq: &nameC}},
						{LastName: &generated.StringFilterInput{Eq: &nameD}},
					}},
				}},
				{FirstName: &generated.StringFilterInput{Eq: &nameE}},
			},
		}

		// Convert to MongoDB filter
		result := resolvers.ConvertCustomerFilterForTest(filter)

		// Verify top-level $and exists
		assert.Contains(t, result, "$and")
		topAndConditions := result["$and"].([]bson.M)
		assert.Len(t, topAndConditions, 2)

		// Verify first AND condition contains $or
		assert.Contains(t, topAndConditions[0], "$or")
		orConditions := topAndConditions[0]["$or"].([]bson.M)
		assert.Len(t, orConditions, 2)

		// Verify OR conditions contain nested $and
		assert.Contains(t, orConditions[0], "$and")
		assert.Contains(t, orConditions[1], "$and")
	})
}

// T015: Unit test for convertCustomerFilter (basic field conversion)
func TestConvertCustomerFilter_BasicFields(t *testing.T) {
	t.Run("String filter - contains", func(t *testing.T) {
		contains := "John"
		filter := &generated.CustomerQueryFilterInput{
			FirstName: &generated.StringFilterInput{
				Contains: &contains,
			},
		}

		result := resolvers.ConvertCustomerFilterForTest(filter)

		assert.Contains(t, result, "firstName")
		assert.Contains(t, result["firstName"], "$regex")
	})

	t.Run("String filter - eq", func(t *testing.T) {
		eq := "john.doe@test.com"
		filter := &generated.CustomerQueryFilterInput{
			UserEmail: &generated.StringFilterInput{
				Eq: &eq,
			},
		}

		result := resolvers.ConvertCustomerFilterForTest(filter)

		assert.Contains(t, result, "userEmail")
		assert.Equal(t, eq, result["userEmail"])
	})

	t.Run("Enum filter - status activation", func(t *testing.T) {
		status := generated.UserStatusActive
		filter := &generated.CustomerQueryFilterInput{
			Status: &generated.CustomerStatusObjectFilterInput{
				Activation: &generated.EnumFilterOfNullableOfUserStatusInput{
					Eq: &status,
				},
			},
		}

		result := resolvers.ConvertCustomerFilterForTest(filter)

		assert.Contains(t, result, "status.activation")
		assert.Equal(t, string(status), result["status.activation"])
	})
}

// T016: Unit test for convertEmployeeFilter
func TestConvertEmployeeFilter_BasicFields(t *testing.T) {
	t.Run("FirstName filter", func(t *testing.T) {
		firstName := "John"
		filter := &generated.EmployeeQueryFilterInput{
			FirstName: &generated.StringFilterInput{
				Eq: &firstName,
			},
		}

		result := resolvers.ConvertEmployeeFilterForTest(filter)

		assert.Contains(t, result, "firstName")
		assert.Equal(t, firstName, result["firstName"])
	})

	t.Run("UserEmail filter - contains", func(t *testing.T) {
		email := "john"
		filter := &generated.EmployeeQueryFilterInput{
			UserEmail: &generated.StringFilterInput{
				Contains: &email,
			},
		}

		result := resolvers.ConvertEmployeeFilterForTest(filter)

		assert.Contains(t, result, "userEmail")
		assert.Contains(t, result["userEmail"], "$regex")
	})
}
