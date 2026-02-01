package resolvers

import (
	"time"

	"github.com/yourusername/air-go/internal/graphql/generated"
	"go.mongodb.org/mongo-driver/bson"
)

// T005: Shared filter converter base functions for converting GraphQL filter inputs to MongoDB filters

// convertStringFilter converts a StringFilterInput to MongoDB filter for the specified field
func convertStringFilter(field string, filter *generated.StringFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Equality operators
	if filter.Eq != nil {
		conditions = append(conditions, bson.M{field: *filter.Eq})
	}
	if filter.Neq != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$ne": *filter.Neq}})
	}

	// List operators
	if filter.In != nil && len(filter.In) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$in": filter.In}})
	}
	if filter.Nin != nil && len(filter.Nin) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$nin": filter.Nin}})
	}

	// Pattern matching operators
	if filter.Contains != nil {
		conditions = append(conditions, bson.M{field: bson.M{
			"$regex":   *filter.Contains,
			"$options": "i", // Case-insensitive
		}})
	}
	if filter.StartsWith != nil {
		conditions = append(conditions, bson.M{field: bson.M{
			"$regex":   "^" + *filter.StartsWith,
			"$options": "i",
		}})
	}
	if filter.EndsWith != nil {
		conditions = append(conditions, bson.M{field: bson.M{
			"$regex":   *filter.EndsWith + "$",
			"$options": "i",
		}})
	}

	// Logical operators (recursive)
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertStringFilter(field, f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertStringFilter(field, f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	// Return combined conditions
	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertEnumFilter converts enum filter with eq/neq/in/nin to MongoDB filter
// This is a helper for nested object filters with enum fields
// Note: There's no generic EnumFilterInput - this works with the field operators pattern
func convertEnumFilterGeneric(field string, eq, neq *string, in, nin []string) bson.M {
	conditions := []bson.M{}

	if eq != nil {
		conditions = append(conditions, bson.M{field: *eq})
	}
	if neq != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$ne": *neq}})
	}
	if in != nil && len(in) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$in": in}})
	}
	if nin != nil && len(nin) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$nin": nin}})
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertComparableFilterDateTime converts a ComparableFilterOfNullableOfDateTimeInput to MongoDB filter
func convertComparableFilterDateTime(field string, filter *generated.ComparableFilterOfNullableOfDateTimeInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Null handling
	if filter.Eq != nil {
		if *filter.Eq == "" {
			// Empty string represents null
			conditions = append(conditions, bson.M{field: nil})
		} else {
			// Parse DateTime string
			if t, err := time.Parse(time.RFC3339, *filter.Eq); err == nil {
				conditions = append(conditions, bson.M{field: t})
			}
		}
	}
	if filter.Neq != nil {
		if *filter.Neq == "" {
			conditions = append(conditions, bson.M{field: bson.M{"$ne": nil}})
		} else {
			if t, err := time.Parse(time.RFC3339, *filter.Neq); err == nil {
				conditions = append(conditions, bson.M{field: bson.M{"$ne": t}})
			}
		}
	}

	// Comparison operators
	if filter.Gt != nil {
		if t, err := time.Parse(time.RFC3339, *filter.Gt); err == nil {
			conditions = append(conditions, bson.M{field: bson.M{"$gt": t}})
		}
	}
	if filter.Gte != nil {
		if t, err := time.Parse(time.RFC3339, *filter.Gte); err == nil {
			conditions = append(conditions, bson.M{field: bson.M{"$gte": t}})
		}
	}
	if filter.Lt != nil {
		if t, err := time.Parse(time.RFC3339, *filter.Lt); err == nil {
			conditions = append(conditions, bson.M{field: bson.M{"$lt": t}})
		}
	}
	if filter.Lte != nil {
		if t, err := time.Parse(time.RFC3339, *filter.Lte); err == nil {
			conditions = append(conditions, bson.M{field: bson.M{"$lte": t}})
		}
	}

	// Logical operators (recursive)
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertComparableFilterDateTime(field, f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertComparableFilterDateTime(field, f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertBooleanFilter converts a BooleanFilterInput to MongoDB filter
func convertBooleanFilter(field string, filter *generated.BooleanFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	if filter.Eq != nil {
		conditions = append(conditions, bson.M{field: *filter.Eq})
	}
	if filter.Neq != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$ne": *filter.Neq}})
	}

	// Logical operators (recursive)
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertBooleanFilter(field, f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertBooleanFilter(field, f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertCollectionFilterCustomerGroup converts a CollectionFilterOfCustomerGroupInput to MongoDB filter
func convertCollectionFilterCustomerGroup(field string, filter *generated.CollectionFilterOfCustomerGroupInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// In/Nin operators for arrays
	if filter.In != nil && len(filter.In) > 0 {
		// MongoDB $in operator: field value must be in the list
		conditions = append(conditions, bson.M{field: bson.M{"$in": filter.In}})
	}
	if filter.Nin != nil && len(filter.Nin) > 0 {
		// MongoDB $nin operator: field value must not be in the list
		conditions = append(conditions, bson.M{field: bson.M{"$nin": filter.Nin}})
	}

	// Logical operators (recursive)
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertCollectionFilterCustomerGroup(field, f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertCollectionFilterCustomerGroup(field, f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// T017: Entity-specific filter converters
// These convert GraphQL FilterInput types to MongoDB bson.M filters

// convertCustomerFilter converts CustomerQueryFilterInput to MongoDB filter
func convertCustomerFilter(filter *generated.CustomerQueryFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Simple field filters
	if filter.FirstName != nil {
		if converted := convertStringFilter("firstName", filter.FirstName); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.LastName != nil {
		if converted := convertStringFilter("lastName", filter.LastName); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.UserEmail != nil {
		if converted := convertStringFilter("userEmail", filter.UserEmail); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.EmployeeEmail != nil {
		if converted := convertStringFilter("employeeEmail", filter.EmployeeEmail); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.IsShared != nil {
		if converted := convertBooleanFilter("isShared", filter.IsShared); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.CreateDate != nil {
		if converted := convertComparableFilterDateTime("createDate", filter.CreateDate); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// Nested object filters (status, payment) - simplified for now
	// TODO: Implement full nested object filtering

	// Collection filter
	if filter.CustomerGroups != nil {
		if converted := convertCollectionFilterCustomerGroup("customerGroups", filter.CustomerGroups); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// Recursive AND/OR
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertCustomerFilter(f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertCustomerFilter(f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// T018: convertEmployeeFilter converts EmployeeQueryFilterInput to MongoDB filter
func convertEmployeeFilter(filter *generated.EmployeeQueryFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Simple field filters
	if filter.FirstName != nil {
		if converted := convertStringFilter("firstName", filter.FirstName); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.LastName != nil {
		if converted := convertStringFilter("lastName", filter.LastName); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.UserEmail != nil {
		if converted := convertStringFilter("userEmail", filter.UserEmail); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// TODO: Add employeeGroups and status filters

	// Recursive AND/OR
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertEmployeeFilter(f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertEmployeeFilter(f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertComparableFilterGUID converts a ComparableFilterOfNullableOfGUIDInput to MongoDB filter
func convertComparableFilterGUID(field string, filter *generated.ComparableFilterOfNullableOfGUIDInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Null handling
	if filter.Eq != nil {
		conditions = append(conditions, bson.M{field: *filter.Eq})
	}
	if filter.Neq != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$ne": *filter.Neq}})
	}

	// List operators
	if filter.In != nil && len(filter.In) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$in": filter.In}})
	}
	if filter.Nin != nil && len(filter.Nin) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$nin": filter.Nin}})
	}

	// Comparison operators (for GUIDs, these are string comparisons)
	if filter.Gt != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$gt": *filter.Gt}})
	}
	if filter.Gte != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$gte": *filter.Gte}})
	}
	if filter.Lt != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$lt": *filter.Lt}})
	}
	if filter.Lte != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$lte": *filter.Lte}})
	}

	// Logical operators (recursive)
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertComparableFilterGUID(field, f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertComparableFilterGUID(field, f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertEnumFilterCreateStatus converts EnumFilterOfNullableOfCreateStatusInput to MongoDB filter
func convertEnumFilterCreateStatus(field string, filter *generated.EnumFilterOfNullableOfCreateStatusInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	if filter.Eq != nil {
		conditions = append(conditions, bson.M{field: *filter.Eq})
	}
	if filter.Neq != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$ne": *filter.Neq}})
	}
	if filter.In != nil && len(filter.In) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$in": filter.In}})
	}
	if filter.Nin != nil && len(filter.Nin) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$nin": filter.Nin}})
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertEnumFilterDeleteStatus converts EnumFilterOfNullableOfDeleteStatusInput to MongoDB filter
func convertEnumFilterDeleteStatus(field string, filter *generated.EnumFilterOfNullableOfDeleteStatusInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	if filter.Eq != nil {
		conditions = append(conditions, bson.M{field: *filter.Eq})
	}
	if filter.Neq != nil {
		conditions = append(conditions, bson.M{field: bson.M{"$ne": *filter.Neq}})
	}
	if filter.In != nil && len(filter.In) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$in": filter.In}})
	}
	if filter.Nin != nil && len(filter.Nin) > 0 {
		conditions = append(conditions, bson.M{field: bson.M{"$nin": filter.Nin}})
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// convertTeamStatusObjectFilter converts TeamStatusObjectFilterInput to MongoDB filter
func convertTeamStatusObjectFilter(filter *generated.TeamStatusObjectFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	if filter.Creation != nil {
		if converted := convertEnumFilterCreateStatus("status.creation", filter.Creation); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.Deletion != nil {
		if converted := convertEnumFilterDeleteStatus("status.deletion", filter.Deletion); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// Recursive AND/OR
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertTeamStatusObjectFilter(f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertTeamStatusObjectFilter(f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// T019: convertTeamFilter converts TeamQueryFilterInput to MongoDB filter
func convertTeamFilter(filter *generated.TeamQueryFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Simple field filters
	if filter.Name != nil {
		if converted := convertStringFilter("name", filter.Name); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.Description != nil {
		if converted := convertStringFilter("description", filter.Description); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}
	if filter.IsShared != nil {
		if converted := convertBooleanFilter("isShared", filter.IsShared); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// Nested object filter
	if filter.Status != nil {
		if converted := convertTeamStatusObjectFilter(filter.Status); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// Recursive AND/OR
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertTeamFilter(f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertTeamFilter(f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// T020: convertExecutionPlanFilter converts ExecutionPlanQueryFilterInput to MongoDB filter
func convertExecutionPlanFilter(filter *generated.ExecutionPlanQueryFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Simple field filter
	if filter.CustomerID != nil {
		if converted := convertComparableFilterGUID("customerId", filter.CustomerID); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// Recursive AND/OR
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertExecutionPlanFilter(f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertExecutionPlanFilter(f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// T021: convertReferencePortfolioFilter converts ReferencePortfolioQueryFilterInput to MongoDB filter
func convertReferencePortfolioFilter(filter *generated.ReferencePortfolioQueryFilterInput) bson.M {
	if filter == nil {
		return bson.M{}
	}

	conditions := []bson.M{}

	// Simple field filter
	if filter.CustomerID != nil {
		if converted := convertComparableFilterGUID("customerId", filter.CustomerID); len(converted) > 0 {
			conditions = append(conditions, converted)
		}
	}

	// Recursive AND/OR
	if filter.And != nil {
		andConditions := []bson.M{}
		for _, f := range filter.And {
			if converted := convertReferencePortfolioFilter(f); len(converted) > 0 {
				andConditions = append(andConditions, converted)
			}
		}
		if len(andConditions) > 0 {
			conditions = append(conditions, bson.M{"$and": andConditions})
		}
	}
	if filter.Or != nil {
		orConditions := []bson.M{}
		for _, f := range filter.Or {
			if converted := convertReferencePortfolioFilter(f); len(converted) > 0 {
				orConditions = append(orConditions, converted)
			}
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": orConditions})
		}
	}

	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}
