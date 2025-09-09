GO_VERSION ?= 1.22
SERVICE_NAME := releaseradar

.PHONY: all build api worker test lint run clean docker-build docker-compose-up docker-compose-down swag

all: build test

# Build targets
build-api:
	@echo "Building API binary..."
	go build -o bin/api ./cmd/api

build-worker:
	@echo "Building Worker binary..."
	go build -o bin/worker ./cmd/worker

build: build-api build-worker

# Test targets
test:
	@echo "Running tests..."
	go test -v ./...

# Lint target (requires golangci-lint installed: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Run targets
run-api:
	@echo "Running API..."
	go run ./cmd/api

run-worker:
	@echo "Running Worker..."
	go run ./cmd/worker

run:
	@echo "Running API and Worker..."
	go run ./cmd/api &
	go run ./cmd/worker &
	wait

# Clean target
clean:
	@echo "Cleaning up..."
	rm -rf bin
	go clean

# Docker targets
docker-build:
	@echo "Building Docker images..."
	docker build -t $(SERVICE_NAME)-api -f build/Dockerfile.api .
	docker build -t $(SERVICE_NAME)-worker -f build/Dockerfile.worker .

docker-compose-up:
	@echo "Starting Docker Compose services..."
	docker-compose -f docker-compose.yml up --build -d

docker-compose-down:
	@echo "Stopping Docker Compose services..."
	docker-compose -f docker-compose.yml down

# Swag generation (requires swag installed: go install github.com/swaggo/swag/cmd/swag@latest)
swag:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go
