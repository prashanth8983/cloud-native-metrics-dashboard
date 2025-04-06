# Cloud-Native Metrics API

A sophisticated API service for querying and analyzing Kubernetes and application metrics from Prometheus, built with Go.

## Features

- **Prometheus Integration**: Query Prometheus metrics with a simple, REST-based API
- **Caching**: Configurable caching to improve performance and reduce load on Prometheus
- **Authentication**: Optional API key, Bearer token, or Basic auth support
- **CORS Support**: Configurable Cross-Origin Resource Sharing
- **Metrics Exposure**: Prometheus-compatible metrics for self-monitoring
- **Health Checks**: Comprehensive health check endpoint
- **High Test Coverage**: Thoroughly tested with unit and integration tests
- **YAML Configuration**: Flexible configuration via YAML, environment variables, or flags
- **Docker Support**: Ready to deploy with Docker and Kubernetes

## Architecture

The API follows a clean architecture pattern with clear separation of concerns:

- **API Layer**: HTTP handlers and middleware
- **Service Layer**: Business logic and metric processing
- **Data Layer**: Prometheus client and caching
- **Configuration**: Environment-aware configuration

## API Endpoints

### Query Endpoints

- `GET /api/query`: Execute an instant Prometheus query
- `GET /api/query_range`: Execute a range Prometheus query
- `GET /api/alerts`: Get current alerts from Prometheus
- `GET /api/metrics/summary`: Get a summary of key metrics

### Utility Endpoints

- `GET /health`: Health check endpoint
- `GET /metrics`: Prometheus metrics endpoint

## Quick Start

### Using Docker

```bash
# Pull the image
docker pull yourusername/metrics-api:latest

# Run with default configuration
docker run -p 8000:8000 yourusername/metrics-api:latest

# Run with custom configuration
docker run -p 8000:8000 \
  -e PROMETHEUS_URL=http://your-prometheus:9090 \
  -e LOG_LEVEL=debug \
  yourusername/metrics-api:latest
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/yourusername/metrics-api.git
cd metrics-api

# Edit environment variables in docker-compose.yml if needed
# Start the services
docker-compose up -d
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/metrics-api.git
cd metrics-api

# Build the binary
make build

# Run the binary
./metrics-api
```

## Configuration

Configuration can be provided via YAML file, environment variables, or command-line flags:

### Environment Variables

```
# Server settings
API_SERVER_PORT=8000                  # Server port
API_SERVER_HOST=0.0.0.0               # Server host
API_READ_TIMEOUT=10s                  # HTTP read timeout
API_WRITE_TIMEOUT=20s                 # HTTP write timeout
API_IDLE_TIMEOUT=120s                 # HTTP idle timeout

# Prometheus settings
PROMETHEUS_URL=http://prometheus:9090 # Prometheus URL
PROMETHEUS_TIMEOUT=10s                # Query timeout
PROMETHEUS_KEEP_ALIVE=30s             # Connection keep-alive
PROMETHEUS_MAX_CONNECTIONS=100        # Max connections

# CORS settings
CORS_ENABLED=true                     # Enable CORS
CORS_ALLOWED_ORIGINS=*                # Allowed origins
CORS_ALLOWED_METHODS=GET,POST,OPTIONS # Allowed methods
CORS_ALLOWED_HEADERS=Content-Type     # Allowed headers
CORS_MAX_AGE=86400                    # Max age in seconds

# Cache settings
CACHE_ENABLED=true                    # Enable cache
CACHE_TTL=60s                         # Cache TTL
CACHE_CLEANUP_PERIOD=5m               # Cache cleanup period
CACHE_MAX_ITEMS=1000                  # Max cache items

# Logging settings
LOG_LEVEL=info                        # Log level
LOG_FORMAT=json                       # Log format (json or text)
LOG_FILE_PATH=                        # Log file path (empty for stdout)
LOG_MAX_SIZE=100                      # Max log file size in MB
LOG_MAX_BACKUPS=5                     # Max number of log files
LOG_MAX_AGE=30                        # Max age of log files in days
LOG_COMPRESSION=true                  # Compress log files

# Metrics settings
METRICS_ENABLED=true                  # Enable metrics endpoint
METRICS_PATH=/metrics                 # Metrics endpoint path
```

### Command-line Flags

```
--port=8000                          # Server port
--host=0.0.0.0                       # Server host
--prometheus-url=http://prometheus:9090 # Prometheus URL
--log-level=info                     # Log level
--cors-enabled=true                  # Enable CORS
--cache-enabled=true                 # Enable cache
--cache-ttl=60s                      # Cache TTL
--metrics-enabled=true               # Enable metrics endpoint
```

## API Usage Examples

### Query Endpoint

```bash
# Simple query
curl "http://localhost:8000/api/query?query=up"

# Query with time
curl "http://localhost:8000/api/query?query=up&time=2023-04-01T00:00:00Z"

# POST query
curl -X POST "http://localhost:8000/api/query" \
  -H "Content-Type: application/json" \
  -d '{"query": "up", "time": "2023-04-01T00:00:00Z"}'
```

### Range Query Endpoint

```bash
# Range query
curl "http://localhost:8000/api/query_range?query=up&start=2023-04-01T00:00:00Z&end=2023-04-01T01:00:00Z&step=1m"

# POST range query
curl -X POST "http://localhost:8000/api/query_range" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "up",
    "start": "2023-04-01T00:00:00Z",
    "end": "2023-04-01T01:00:00Z",
    "step": "1m"
  }'
```

### Alerts Endpoint

```bash
# Get all alerts
curl "http://localhost:8000/api/alerts"

# Filter alerts by severity
curl "http://localhost:8000/api/alerts?severity=critical"

# Filter alerts by state
curl "http://localhost:8000/api/alerts?active=true"
```

### Metrics Summary Endpoint

```bash
# Get metrics summary
curl "http://localhost:8000/api/metrics/summary"
```

## Development

### Prerequisites

- Go 1.19 or higher
- Docker (for containerized development)
- Prometheus instance (for testing)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/yourusername/metrics-api.git
cd metrics-api

# Install dependencies
make deps

# Run tests
make test

# Run linting
make lint

# Build
make build

# Run locally
make run
```

### Folder Structure

```
metrics-api/
├── cmd/                    # Application entrypoints
│   └── server/             # API server
│       └── main.go
├── internal/               # Internal packages
│   ├── api/                # API server implementation
│   │   ├── handlers/       # Request handlers
│   │   ├── middleware/     # HTTP middleware
│   │   ├── routes/         # API routes
│   │   └── server.go       # HTTP server
│   ├── cache/              # Caching implementation
│   ├── config/             # Configuration
│   ├── models/             # Data models
│   └── prometheus/         # Prometheus client
├── pkg/                    # Public packages
│   └── logger/             # Logging package
├── Makefile                # Build commands
├── Dockerfile              # Container definition
├── docker-compose.yml      # Local development
└── README.md               # Documentation
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.