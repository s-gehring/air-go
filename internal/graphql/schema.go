package graphql

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// Schema represents a loaded and validated GraphQL schema
type Schema struct {
	Schema     *ast.Schema
	RawContent string
	LoadedAt   time.Time
	SchemaPath string
}

// LoadSchema loads and validates the GraphQL schema from the specified file
func LoadSchema(schemaPath string) (*Schema, error) {
	log.Info().Str("path", schemaPath).Msg("Loading GraphQL schema")

	// Check if file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("schema file not found: %s", schemaPath)
	}

	// Read schema file
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("schema file is empty: %s", schemaPath)
	}

	// Parse and validate schema
	source := &ast.Source{
		Name:  schemaPath,
		Input: string(content),
	}

	schema, gqlErr := gqlparser.LoadSchema(source)
	if gqlErr != nil {
		return nil, fmt.Errorf("schema validation failed: %w", gqlErr)
	}

	// Verify schema has at least a Query type
	if schema.Query == nil {
		return nil, fmt.Errorf("schema must define a Query type")
	}

	loadedSchema := &Schema{
		Schema:     schema,
		RawContent: string(content),
		LoadedAt:   time.Now(),
		SchemaPath: schemaPath,
	}

	log.Info().
		Str("path", schemaPath).
		Int("types", len(schema.Types)).
		Msg("Schema loaded and validated successfully")

	return loadedSchema, nil
}
