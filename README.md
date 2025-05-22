# YouTube Music Video API

A Go API for searching YouTube music videos, built with Gin and OpenAPI documentation.

## Features

- RESTful API with Gin framework
- OpenAPI/Swagger documentation
- Health check endpoint
- Search endpoint for music videos with caching
- LRU cache for improved performance
- Docker support for deployment

## Getting Started

### Prerequisites

- Go 1.24.2 or later

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

### Running the Server

```bash
go run cmd/api/main.go
```

The server will start on port 9898 by default, or use the `PORT` environment variable.

### Docker

Build and run with Docker:

```bash
docker build -t youtube-music-video-api .
docker run -p 9898:9898 youtube-music-video-api
```

## Testing

### Running Tests

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Run only unit tests (skip integration tests):
```bash
go test -short ./...
```

Run specific test packages:
```bash
# Cache tests
go test ./internal/services -run TestLRUCache

# YouTube service tests
go test ./internal/services -run TestYouTubeService

# Handler tests
go test ./internal/handlers

# Integration tests
go test ./test
```

### Test Coverage

The test suite includes:
- **Unit Tests**: Cache, YouTube service, and HTTP handlers
- **Integration Tests**: Complete API flows with real YouTube requests
- **Mock Tests**: Isolated testing with mocked HTTP responses
- **Concurrent Tests**: Thread safety and caching behavior
- **Edge Cases**: Error handling and parameter validation

Note: Integration tests may be slow as they make real HTTP requests to YouTube. Use `-short` flag to skip them during development.

## API Endpoints

- `GET /health` - Health check
- `GET /search?title=TITLE&artists=ARTIST1,ARTIST2` - Search for music videos
- `GET /swagger/index.html` - OpenAPI documentation

### Search Parameters

- `title` (required): The song title to search for
- `artists` (optional): Comma-separated list of artist names

## Project Structure

```
├── cmd/api/          # Application entry point
├── internal/
│   ├── handlers/     # HTTP handlers
│   ├── models/       # Data models
│   └── services/     # Business logic
└── docs/             # Generated OpenAPI docs
```

## License

MIT License - see [LICENSE](LICENSE) file for details.