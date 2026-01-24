package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/yourusername/air-go/internal/config"
	"github.com/yourusername/air-go/internal/db"
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

	// Initialize MongoDB client
	dbClient, err := db.NewClient(cfg.Database, log.Logger)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to create MongoDB client")
	}

	// Connect to MongoDB with timeout
	connectCtx, connectCancel := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	err = dbClient.Connect(connectCtx)
	connectCancel()

	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to connect to MongoDB")
	}

	log.Info().
		Str("database", cfg.Database.Database).
		Uint64("pool_size", cfg.Database.MaxPoolSize).
		Msg("MongoDB connection established")

	// Setup graceful shutdown for MongoDB
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer disconnectCancel()

		if err := dbClient.Disconnect(disconnectCtx); err != nil {
			log.Error().Err(err).Msg("Error disconnecting from MongoDB")
		}
		dbClient.Close()
	}()

	// Create and start HTTP server
	srv := server.New(cfg)

	log.Info().
		Dur("startup_time", time.Since(startTime)).
		Msg("Server initialization complete")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-errChan:
		log.Fatal().
			Err(err).
			Msg("Server failed")
	case sig := <-sigChan:
		log.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received")
	}

	log.Info().Msg("Server shutdown complete")
}
