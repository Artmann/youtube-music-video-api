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

# Generate OpenAPI documentation
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go
```

## Architecture

This is a Go REST API built with Gin framework following clean architecture principles:

- **cmd/api/** - Application entry point with server setup and routing
- **internal/handlers/** - HTTP request handlers (health, search, swagger redirect)
- **internal/services/** - Business logic layer (ready for implementation)
- **internal/models/** - Data structures and types (ready for implementation)
- **docs/** - Auto-generated OpenAPI/Swagger documentation

The API uses Gin for HTTP routing and Swagger/OpenAPI for documentation. The server supports configurable ports via the PORT environment variable (defaults to 9898).

## Key Endpoints

- `GET /` - Redirects to Swagger UI
- `GET /health` - Health check
- `GET /search` - Music video search (stub implementation)
- `GET /swagger/index.html` - Interactive API documentation

## Framework Details

- **Gin** web framework with JSON responses
- **Swaggo** for OpenAPI spec generation with annotations
- **gin-swagger** for serving interactive documentation
- Standard Go project layout with cmd/ and internal/ structure

The codebase is prepared for YouTube music video search functionality with placeholder endpoints and proper API documentation integration.