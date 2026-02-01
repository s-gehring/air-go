package resolvers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Cursor represents the internal structure of a pagination cursor
// T004: Cursor encoding/decoding utilities for pagination
type Cursor struct {
	SortFields []interface{} `json:"s"` // Values of sort fields at cursor position
	Identifier string        `json:"i"` // Entity identifier (UUID) as tiebreaker
}

// encodeCursor serializes a Cursor to a base64-encoded JSON string
// Used to create opaque cursor strings for pagination (startCursor, endCursor)
func encodeCursor(cursor Cursor) (string, error) {
	// Serialize to JSON
	jsonBytes, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return encoded, nil
}

// decodeCursor deserializes a base64-encoded cursor string back to a Cursor struct
// Returns error if cursor format is invalid (invalid base64 or malformed JSON)
func decodeCursor(cursorStr string) (*Cursor, error) {
	return DecodeCursor(cursorStr)
}

// DecodeCursor is the exported version for testing
func DecodeCursor(cursorStr string) (*Cursor, error) {
	if cursorStr == "" {
		return nil, newInvalidInputError("cursor cannot be empty")
	}

	// Decode from base64
	jsonBytes, err := base64.StdEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, newInvalidInputError("invalid cursor format: not valid base64")
	}

	// Deserialize from JSON
	var cursor Cursor
	if err := json.Unmarshal(jsonBytes, &cursor); err != nil {
		return nil, newInvalidInputError("invalid cursor format: malformed cursor data")
	}

	// Validate cursor has identifier
	if cursor.Identifier == "" {
		return nil, newInvalidInputError("invalid cursor: missing identifier")
	}

	return &cursor, nil
}
