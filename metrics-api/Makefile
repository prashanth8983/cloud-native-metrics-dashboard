# Makefile for metrics-api
.PHONY: build clean test lint run docker-build docker-run help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GORUN=$(GOCMD) run
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=metrics-api
MAIN_PATH=cmd/server/main.go

# Docker parameters
DOCKER_IMAGE=metrics-api
DOCKER_TAG=latest
DOCKER_BUILD=docker build
DOCKER_RUN=docker run

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

# Clean build files
clean:
	rm -f $(BINARY_NAME)
	rm -rf ./dist

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt

# Lint the code
lint:
	$(GOVET) ./...
	golangci-lint run

# Format the code
fmt:
	$(GOFMT) ./...

# Run the application
run:
	$(GORUN) $(MAIN_PATH)

# Build docker image
docker-build:
	$(DOCKER_BUILD) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run docker container
docker-run:
	$(DOCKER_RUN) -p 8000:8000 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Download dependencies
deps:
	$(GOMOD) download

# Run dependency updates
deps-update:
	$(GOMOD) tidy

# Create build for multiple platforms
build-all: clean
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# Show help
help:
	@echo "Available commands:"
	@echo "  make build            - Build the binary"
	@echo "  make clean            - Remove built binary and dist directory"
	@echo "  make test             - Run tests"
	@echo "  make test-coverage    - Run tests with coverage"
	@echo "  make lint             - Lint the code"
	@echo "  make fmt              - Format the code"
	@echo "  make run              - Run the application"
	@echo "  make docker-build     - Build docker image"
	@echo "  make docker-run       - Run docker container"
	@echo "  make deps             - Download dependencies"
	@echo "  make deps-update      - Update dependencies"
	@echo "  make build-all        - Build for multiple platforms"