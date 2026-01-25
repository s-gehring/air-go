 # air-go Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-24

## Active Technologies
- Go 1.25.6 + Official MongoDB Go Driver (mongo-go-driver), testcontainers-go (test automation), Docker Compose (dev environment) (002-mongodb-persistence)
- MongoDB 8.2.3 (containerized) (002-mongodb-persistence)
- Go 1.25.6 + gqlgen v0.17.86 (GraphQL server), mongo-go-driver v1.13.1 (MongoDB client), zerolog v1.34.0 (structured logging) (001-graphql-queries)
- MongoDB 8.2.3 (existing collections with schemas already defined) (001-graphql-queries)

- Go 1.21+ + gqlgen (GraphQL server), chi v5 (HTTP router), golang-jwt/jwt v5 (authentication), zerolog (logging), viper (config), rs/cors (CORS) (001-graphql-api-setup)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.21+

## Code Style

Go 1.21+: Follow standard conventions

## Recent Changes
- 001-graphql-queries: Added Go 1.25.6 + gqlgen v0.17.86 (GraphQL server), mongo-go-driver v1.13.1 (MongoDB client), zerolog v1.34.0 (structured logging)
- 002-mongodb-persistence: Added Go 1.25.6 + Official MongoDB Go Driver (mongo-go-driver), testcontainers-go (test automation), Docker Compose (dev environment)

- 001-graphql-api-setup: Added Go 1.21+ + gqlgen (GraphQL server), chi v5 (HTTP router), golang-jwt/jwt v5 (authentication), zerolog (logging), viper (config), rs/cors (CORS)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
