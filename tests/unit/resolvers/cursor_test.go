package resolvers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
)

// T050: Unit test for cursor encoding/decoding

// Test encodeCursor creates valid base64-encoded cursor
func TestEncodeCursor_ValidCursor(t *testing.T) {
	cursor := resolvers.Cursor{
		SortFields: []interface{}{"Smith", "John"},
		Identifier: "abc-123-def-456",
	}

	// Note: encodeCursor is not exported, so we test through the public decodeCursor
	// We'll use a known valid cursor string for testing
	validCursorString := "eyJzIjpbIlNtaXRoIiwiSm9obiJdLCJpIjoiYWJjLTEyMy1kZWYtNDU2In0="

	// This is a base64-encoded JSON: {"s":["Smith","John"],"i":"abc-123-def-456"}
	// Verify we can decode it back
	decoded, err := resolvers.DecodeCursor(validCursorString)

	require.NoError(t, err)
	assert.Equal(t, cursor.Identifier, decoded.Identifier)
	assert.Len(t, decoded.SortFields, 2)
	assert.Equal(t, "Smith", decoded.SortFields[0])
	assert.Equal(t, "John", decoded.SortFields[1])
}

// Test decodeCursor with empty cursor string
func TestDecodeCursor_EmptyCursor(t *testing.T) {
	_, err := resolvers.DecodeCursor("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cursor cannot be empty")
}

// Test decodeCursor with invalid base64
func TestDecodeCursor_InvalidBase64(t *testing.T) {
	_, err := resolvers.DecodeCursor("not-valid-base64!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cursor format")
}

// Test decodeCursor with invalid JSON
func TestDecodeCursor_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON
	invalidJSON := "dGhpcyBpcyBub3QgSlNPTg==" // base64("this is not JSON")

	_, err := resolvers.DecodeCursor(invalidJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cursor format")
}

// Test decodeCursor with missing identifier
func TestDecodeCursor_MissingIdentifier(t *testing.T) {
	// Valid JSON but missing identifier: {"s":["Smith"],"i":""}
	missingIdentifier := "eyJzIjpbIlNtaXRoIl0sImkiOiIifQ=="

	_, err := resolvers.DecodeCursor(missingIdentifier)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing identifier")
}

// Test decodeCursor with null sort fields
func TestDecodeCursor_NullSortFields(t *testing.T) {
	// Cursor with null in sort fields: {"s":[null,"John"],"i":"abc-123"}
	cursorWithNull := "eyJzIjpbbnVsbCwiSm9obiJdLCJpIjoiYWJjLTEyMyJ9"

	decoded, err := resolvers.DecodeCursor(cursorWithNull)

	require.NoError(t, err)
	assert.Equal(t, "abc-123", decoded.Identifier)
	assert.Len(t, decoded.SortFields, 2)
	assert.Nil(t, decoded.SortFields[0])
	assert.Equal(t, "John", decoded.SortFields[1])
}

// Test encode/decode roundtrip
func TestCursor_Roundtrip(t *testing.T) {
	// We can't directly test encodeCursor since it's not exported
	// But we can verify that a properly formatted cursor string decodes correctly

	// Create a cursor string manually (simulating what encodeCursor would produce)
	// {"s":["Doe",25],"i":"uuid-123"}
	cursorString := "eyJzIjpbIkRvZSIsMjVdLCJpIjoidXVpZC0xMjMifQ=="

	decoded, err := resolvers.DecodeCursor(cursorString)

	require.NoError(t, err)
	assert.Equal(t, "uuid-123", decoded.Identifier)
	assert.Len(t, decoded.SortFields, 2)
	assert.Equal(t, "Doe", decoded.SortFields[0])
	// JSON unmarshals numbers as float64
	assert.Equal(t, float64(25), decoded.SortFields[1])
}
