# Build stage
FROM golang:1.23-alpine AS builder

# Add necessary build tools
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty || echo 'dev')" \
    -o /go/bin/metrics-api ./cmd/server

# Final stage
FROM alpine:3.18

# Add runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Import from builder
COPY --from=builder /go/bin/metrics-api /usr/local/bin/metrics-api

# Create a non-root user
RUN adduser -D -H -h /app appuser
USER appuser

# Set environment variables
ENV SERVER_PORT=8080
ENV PROMETHEUS_URL=http://prometheus:9090
ENV LOG_LEVEL=debug
ENV LOG_FORMAT=json
ENV METRICS_PATH=/metrics

# Expose the application port
EXPOSE 8080

# Set health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v1/health || exit 1

# Run the application
ENTRYPOINT ["metrics-api"]
