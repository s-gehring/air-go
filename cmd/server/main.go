package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/yourusername/air-go/internal/config"
	"github.com/yourusername/air-go/internal/graphql"
	"github.com/yourusername/air-go/internal/logger"
	"github.com/yourusername/air-go/internal/server"
)

func main() {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Setup(cfg.LogFormat)

	log.Info().Msg("Starting GraphQL API server")

	// Load and validate GraphQL schema
	schema, err := graphql.LoadSchema(cfg.SchemaPath)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("schema_path", cfg.SchemaPath).
			Msg("Failed to load GraphQL schema - server cannot start")
	}

	log.Info().
		Str("schema_path", schema.SchemaPath).
		Int("types", len(schema.Schema.Types)).
		Dur("load_time", time.Since(startTime)).
		Msg("GraphQL schema loaded successfully")

	// Create and start HTTP server
	srv := server.New(cfg)

	log.Info().
		Dur("startup_time", time.Since(startTime)).
		Msg("Server initialization complete")

	// Start the server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatal().
			Err(err).
			Msg("Server failed")
	}

	log.Info().Msg("Server shutdown complete")
}
