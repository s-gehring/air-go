# air-go GraphQL API

Enterprise GraphQL API backend built with Go, gqlgen, and MongoDB.

## Features

- **GraphQL API**: Type-safe GraphQL server with gqlgen
- **MongoDB Persistence**: Robust MongoDB integration with connection pooling and automatic retry logic
- **Health Monitoring**: HTTP and GraphQL health check endpoints with database status
- **Structured Logging**: JSON structured logging with zerolog
- **Authentication**: JWT-based authentication middleware
- **CORS Support**: Configurable CORS for frontend integration
- **Docker Support**: Docker Compose for local development and testing
- **Comprehensive Testing**: Unit, integration, and E2E tests with testcontainers

## Quick Start

### Prerequisites

- Go 1.25.6 or later
- Docker and Docker Compose (for local development)
- MongoDB 8.2+ (or use Docker Compose)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/air-go.git
cd air-go

# Install dependencies
go mod download
```

### Running with Docker Compose

The easiest way to get started is using Docker Compose, which sets up MongoDB automatically:

```bash
# Start MongoDB and the API server
docker compose up

# The API will be available at http://localhost:8080
```

### Running Locally

1. **Start MongoDB** (if not using Docker Compose):

```bash
# Using Docker
docker run -d \
  --name air-mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_DATABASE=air \
  mongo:8.2.3

# Or use an existing MongoDB instance
```

2. **Configure environment variables** (optional):

```bash
cp .env.example .env
# Edit .env with your MongoDB connection details
```

3. **Run the server**:

```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default.

## MongoDB Setup

### Using Docker Compose (Recommended)

The included `docker-compose.yml` file sets up MongoDB automatically:

```yaml
services:
  mongodb:
    image: mongo:8.2.3
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_DATABASE: air
    volumes:
      - mongodb_data:/data/db

volumes:
  mongodb_data:
```

### Manual MongoDB Setup

If you prefer to run MongoDB manually:

```bash
# Install MongoDB 8.2+
# https://www.mongodb.com/docs/manual/installation/

# Start MongoDB
mongod --dbpath /path/to/data --port 27017

# Create database
mongosh
> use air
```

### Connection Configuration

Configure MongoDB connection via environment variables:

- `MONGODB_URI`: Connection string (default: `mongodb://localhost:27017`)
- `MONGODB_DATABASE`: Database name (default: `air`)
- `MONGODB_MAX_POOL_SIZE`: Max connections (default: 10, range: 10-20)
- `MONGODB_CONNECT_TIMEOUT`: Connection timeout (default: 30s, range: 10-60s)
- `MONGODB_OPERATION_TIMEOUT`: Operation timeout (default: 10s, range: 1-30s)

See `.env.example` for all configuration options.

## API Endpoints

### Health Check

```bash
# HTTP endpoint
curl http://localhost:8080/health

# Response
{
  "status": "ok",
  "timestamp": "2026-01-24T22:00:00Z",
  "database": {
    "status": "connected",
    "message": "MongoDB connected",
    "latency_ms": 2
  }
}
```

### GraphQL Endpoint

```bash
# GraphQL endpoint (requires authentication)
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "{ health { status timestamp } }"}'
```

### GraphQL Playground

Visit `http://localhost:8080/graphql` in your browser for the interactive GraphQL playground (development only).

## Testing

### Run All Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run only unit tests
go test ./tests/unit/...

# Run integration tests (requires Docker)
go test ./tests/integration/...

# Run E2E tests
go test ./tests/e2e/...
```

### Run Benchmarks

```bash
# Run performance benchmarks
go test -bench=. ./tests/integration/

# Run specific benchmark
go test -bench=BenchmarkSimpleDatabaseOperations ./tests/integration/
```

### Integration Testing

Integration tests use testcontainers to automatically manage MongoDB Docker containers:

```bash
# Integration tests will automatically:
# 1. Start a MongoDB container
# 2. Run tests
# 3. Clean up the container

go test ./tests/integration/... -v
```

## Development

### Project Structure

```
air-go/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── db/              # MongoDB client and operations
│   ├── graphql/         # GraphQL schema and resolvers
│   ├── health/          # Health check handlers
│   ├── logger/          # Logging setup
│   ├── server/          # HTTP server and routing
│   └── middleware/      # HTTP middleware
├── tests/
│   ├── unit/            # Unit tests
│   ├── integration/     # Integration tests
│   └── e2e/             # End-to-end tests
├── schema.graphqls      # GraphQL schema definition
├── docker-compose.yml   # Docker Compose configuration
└── README.md            # This file
```

### Code Generation

GraphQL code is generated using gqlgen:

```bash
# Generate GraphQL resolvers and types
go run github.com/99designs/gqlgen generate

# Configuration is in gqlgen.yml
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check for code issues
go vet ./...
```

## Configuration

### Environment Variables

See `.env.example` for all available configuration options. Key variables:

- `PORT`: HTTP server port (default: 8080)
- `LOG_FORMAT`: Logging format - json or text (default: json)
- `MONGODB_URI`: MongoDB connection string
- `JWT_SECRET`: Secret key for JWT authentication (min 32 characters)
- `CORS_ORIGINS`: Allowed CORS origins (comma-separated)

### Configuration File

Alternatively, use a configuration file (config.yaml):

```yaml
port: 8080
log_format: json
mongodb:
  uri: mongodb://localhost:27017
  database: air
  max_pool_size: 10
  connect_timeout: 30s
  operation_timeout: 10s
```

## Deployment

### Production Considerations

1. **Connection Pool**: Set `MONGODB_MAX_POOL_SIZE` to 20 for production
2. **Timeouts**: Use default timeouts (30s connect, 10s operation)
3. **Logging**: Use JSON format for structured logging
4. **Health Checks**: Configure orchestrator to poll `/health` endpoint
5. **Graceful Shutdown**: Server handles SIGTERM/SIGINT for graceful shutdown

### Docker Deployment

```bash
# Build Docker image
docker build -t air-go:latest .

# Run with Docker Compose
docker compose up -d
```

## Documentation

For detailed documentation, see:

- [MongoDB Persistence Layer Quickstart](specs/002-mongodb-persistence/quickstart.md)
- [GraphQL Schema](schema.graphqls)
- [API Contracts](specs/002-mongodb-persistence/contracts/)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[License TBD]

## Support

For questions or issues, please open a GitHub issue.
