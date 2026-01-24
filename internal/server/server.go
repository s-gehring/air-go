package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	"github.com/yourusername/air-go/internal/config"
	"github.com/yourusername/air-go/internal/graphql/generated"
	"github.com/yourusername/air-go/internal/graphql/resolvers"
	"github.com/yourusername/air-go/internal/health"
	"github.com/yourusername/air-go/internal/server/middleware"
)

// Server represents the HTTP server
type Server struct {
	config   *config.Config
	router   *chi.Mux
	srv      *http.Server
	dbClient health.DBHealthChecker // Database client for health checks
}

// Option is a function that configures the server
type Option func(*Server)

// WithDatabaseClient sets the database client for health checks
func WithDatabaseClient(dbClient health.DBHealthChecker) Option {
	return func(s *Server) {
		s.dbClient = dbClient
	}
}

// New creates a new HTTP server with configured routes and middleware
func New(cfg *config.Config, opts ...Option) *Server {
	s := &Server{
		config: cfg,
		router: chi.NewRouter(),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// setupMiddleware configures the middleware chain
func (s *Server) setupMiddleware() {
	// Basic middleware (applied to all routes)
	s.router.Use(chimiddleware.RequestID)
	s.router.Use(chimiddleware.RealIP)
	s.router.Use(chimiddleware.Recoverer)
	s.router.Use(middleware.LoggingMiddleware)

	// CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   s.config.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	s.router.Use(corsMiddleware.Handler)
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint (no authentication required)
	// Passes database client if available for health monitoring
	s.router.Get("/health", health.Handler(s.dbClient))

	// GraphQL endpoint (authentication required)
	// This will be implemented in later phases (T025)
	s.router.Route("/graphql", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		r.Post("/", s.graphQLHandler)
	})
}

// graphQLHandler handles GraphQL requests
func (s *Server) graphQLHandler(w http.ResponseWriter, r *http.Request) {
	// Create resolver with database client for health monitoring (T088)
	resolver := &resolvers.Resolver{
		DBClient: s.dbClient,
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	srv.ServeHTTP(w, r)
}

// ServeHTTP implements http.Handler interface to allow using Server with httptest
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Start begins listening for HTTP requests and handles graceful shutdown
func (s *Server) Start() error {
	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Info().
			Int("port", s.config.Port).
			Str("schema_path", s.config.SchemaPath).
			Msg("Starting HTTP server")

		serverErrors <- s.srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal, starting graceful shutdown")

		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error during server shutdown")
			// Force close the server
			if closeErr := s.srv.Close(); closeErr != nil {
				return fmt.Errorf("could not stop server gracefully: %w", closeErr)
			}
			return fmt.Errorf("server shutdown error: %w", err)
		}

		log.Info().Msg("Server stopped gracefully")
	}

	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
