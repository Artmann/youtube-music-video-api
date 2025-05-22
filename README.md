# YouTube Music Video API

A Go API for searching YouTube music videos, built with Gin and OpenAPI documentation.

## Features

- RESTful API with Gin framework
- OpenAPI/Swagger documentation
- Health check endpoint
- Search endpoint for music videos

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

## API Endpoints

- `GET /health` - Health check
- `GET /search` - Search for music videos
- `GET /swagger/index.html` - OpenAPI documentation

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