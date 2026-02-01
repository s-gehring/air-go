package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/db"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"go.mongodb.org/mongo-driver/bson"
)

// T012: E2E test for teamSearch basic filtering (name startsWith filter)
func TestTeamSearch_BasicFiltering_NameStartsWith(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test teams
	seedTeamForSearch(t, dbClient, "team-001", "Sales Team Alpha", "INIT")
	seedTeamForSearch(t, dbClient, "team-002", "Sales Team Beta", "INIT")
	seedTeamForSearch(t, dbClient, "team-003", "Marketing Team", "INIT")
	seedTeamForSearch(t, dbClient, "team-004", "Engineering Team", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: name startsWith "Sales"
	startsSales := "Sales"
	filter := &generated.TeamQueryFilterInput{
		Name: &generated.StringFilterInput{
			StartsWith: &startsSales,
		},
	}

	// Execute teamSearch query
	first := int64(10)
	result, err := queryResolver.TeamSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return 2 teams starting with "Sales"
	assert.Equal(t, int64(2), result.Count)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Len(t, result.Data, 2)

	// Verify both results start with "Sales"
	for _, team := range result.Data {
		assert.True(t, strings.HasPrefix(*team.Name, "Sales"))
	}
}

// T035: E2E test for teamSearch multi-field sorting (name ASC then createDate DESC)
func TestTeamSearch_MultiFieldSorting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test teams with same name but different descriptions
	seedTeamWithDescription(t, dbClient, "team-010", "Alpha Team", "AAA Description", "INIT")
	seedTeamWithDescription(t, dbClient, "team-011", "Alpha Team", "ZZZ Description", "INIT")
	seedTeamWithDescription(t, dbClient, "team-012", "Beta Team", "BBB Description", "INIT")
	seedTeamWithDescription(t, dbClient, "team-013", "Beta Team", "YYY Description", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build sorter: name ASC, then description DESC
	sortAsc := generated.SortEnumTypeAsc
	sortDesc := generated.SortEnumTypeDesc
	sorter := []*generated.TeamQuerySorterInput{
		{
			Name: &sortAsc,
		},
		{
			Description: &sortDesc,
		},
	}

	// Execute teamSearch query
	first := int64(10)
	result, err := queryResolver.TeamSearch(ctx, nil, sorter, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(4), result.Count)
	assert.Len(t, result.Data, 4)

	// Verify sorting: Alpha teams first (Z before A in DESC), then Beta teams (Y before B in DESC)
	assert.Equal(t, "Alpha Team", *result.Data[0].Name)
	assert.Equal(t, "ZZZ Description", *result.Data[0].Description) // Z comes first in DESC
	assert.Equal(t, "Alpha Team", *result.Data[1].Name)
	assert.Equal(t, "AAA Description", *result.Data[1].Description) // A comes after in DESC
	assert.Equal(t, "Beta Team", *result.Data[2].Name)
	assert.Equal(t, "YYY Description", *result.Data[2].Description) // Y comes first in DESC
	assert.Equal(t, "Beta Team", *result.Data[3].Name)
	assert.Equal(t, "BBB Description", *result.Data[3].Description) // B comes after in DESC
}

// T062: E2E test for nested OR filters (multiple OR conditions)
func TestTeamSearch_NestedORFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	dbClient := setupTestDatabase(t)
	defer teardownTestDatabase(t, dbClient)

	// Seed test teams
	seedTeamForSearch(t, dbClient, "team-020", "Alpha Team", "INIT")
	seedTeamForSearch(t, dbClient, "team-021", "Beta Team", "INIT")
	seedTeamForSearch(t, dbClient, "team-022", "Gamma Team", "INIT")
	seedTeamForSearch(t, dbClient, "team-023", "Delta Team", "INIT")

	// Create resolver
	resolver := resolvers.NewResolver(dbClient)
	queryResolver := resolver.Query()

	// Build filter: name contains "Alpha" OR name contains "Gamma"
	containsAlpha := "Alpha"
	containsGamma := "Gamma"
	filter := &generated.TeamQueryFilterInput{
		Or: []*generated.TeamQueryFilterInput{
			{
				Name: &generated.StringFilterInput{
					Contains: &containsAlpha,
				},
			},
			{
				Name: &generated.StringFilterInput{
					Contains: &containsGamma,
				},
			},
		},
	}

	// Execute teamSearch query
	first := int64(10)
	result, err := queryResolver.TeamSearch(ctx, filter, nil, &first, nil, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return 2 teams (Alpha and Gamma)
	assert.Equal(t, int64(2), result.Count)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Len(t, result.Data, 2)

	// Verify results contain Alpha or Gamma
	foundAlpha := false
	foundGamma := false
	for _, team := range result.Data {
		if strings.Contains(*team.Name, "Alpha") {
			foundAlpha = true
		}
		if strings.Contains(*team.Name, "Gamma") {
			foundGamma = true
		}
	}
	assert.True(t, foundAlpha)
	assert.True(t, foundGamma)
}

// Helper: Seed team for search tests
func seedTeamForSearch(t *testing.T, dbClient *db.Client, identifier, name, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("teams")
	doc := bson.M{
		"identifier":      identifier,
		"name":            name,
		"createDate":      time.Now().Format(time.RFC3339),
		"status": bson.M{
			"deletion": deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}

// Helper: Seed team with specific createDate
func seedTeamWithDate(t *testing.T, dbClient *db.Client, identifier, name, createDate, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("teams")
	doc := bson.M{
		"identifier":      identifier,
		"name":            name,
		"createDate":      createDate,
		"status": bson.M{
			"deletion": deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}

// Helper: Seed team with name and description
func seedTeamWithDescription(t *testing.T, dbClient *db.Client, identifier, name, description, deletionStatus string) {
	t.Helper()
	ctx := context.Background()

	collection := dbClient.Collection("teams")
	doc := bson.M{
		"identifier":      identifier,
		"name":            name,
		"description":     description,
		"createDate":      time.Now().Format(time.RFC3339),
		"status": bson.M{
			"deletion": deletionStatus,
		},
		"actionIndicator": "NONE",
	}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)
}
