# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

```bash
# Run the server (default port 9898)
go run cmd/api/main.go

# Build the application
go build -o bin/api cmd/api/main.go

# Install/update dependencies  
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run only unit tests (skip slow integration tests)
go test -short ./...

# Run specific test suites
go test ./internal/services -run TestLRUCache     # Cache tests
go test ./internal/services -run TestYouTubeService  # YouTube service tests
go test ./internal/handlers                       # Handler tests
go test ./test                                   # Integration tests

# Generate OpenAPI documentation
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go

# Docker commands
docker build -t youtube-music-video-api .
docker run -p 9898:9898 youtube-music-video-api
```

## Architecture

This is a Go REST API built with Gin framework following clean architecture principles:

- **cmd/api/** - Application entry point with server setup and routing
- **internal/handlers/** - HTTP request handlers (health, search, swagger redirect)
- **internal/services/** - Business logic layer with YouTube search and LRU caching
- **internal/models/** - Data structures and types (ready for implementation)
- **docs/** - Auto-generated OpenAPI/Swagger documentation
- **test/** - Integration tests

The API uses Gin for HTTP routing and Swagger/OpenAPI for documentation. The server supports configurable ports via the PORT environment variable (defaults to 9898).

## Testing Strategy

The codebase includes comprehensive test coverage:

- **Unit Tests** (`*_test.go` files in internal/ packages)
  - Cache functionality with LRU eviction
  - YouTube service with mock HTTP clients
  - Handler validation and response formatting
  
- **Integration Tests** (`test/` package)
  - Complete API flows with real YouTube requests
  - Caching behavior verification
  - Concurrent request handling
  - Error scenarios and edge cases

- **Mock Tests** (`*_mock_test.go` files)
  - Isolated testing with HTTP mocking
  - Network error simulation
  - Various response scenarios

Use `go test -short ./...` during development to skip slow integration tests.

## Key Endpoints

- `GET /` - Redirects to Swagger UI
- `GET /health` - Health check
- `GET /search?title=TITLE&artists=ARTIST1,ARTIST2` - YouTube music video search with caching
- `GET /swagger/index.html` - Interactive API documentation

## Framework Details

- **Gin** web framework with JSON responses
- **Swaggo** for OpenAPI spec generation with annotations
- **gin-swagger** for serving interactive documentation
- **LRU Cache** for improved search performance (5000 item capacity)
- **YouTube Search** with HTML parsing and video ID extraction
- **Docker** support for containerized deployment
- Standard Go project layout with cmd/ and internal/ structure

The API implements full YouTube music video search functionality with intelligent caching and comprehensive error handling.