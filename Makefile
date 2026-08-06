# Makefile for Stalkeer

.PHONY: all build test clean run lint fmt help front-install front-dev front-build front-lint dev

# Variables
BINARY_NAME=stalkeer
CONFIG_FILE=config.yml
BIN_DIR=bin
CMD_DIR=cmd
MAIN_FILE=$(CMD_DIR)/main.go
VERSION ?= dev
REGISTRY ?= docker.io/germainlefebvre4
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

all: test build

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./$(CMD_DIR)/...
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo "Tests complete"

## coverage: Generate test coverage report
coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## run: Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BIN_DIR)/$(BINARY_NAME) server

## lint: Run linters
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run
	@echo "Linting complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not found, skipping import organization"
	@echo "Formatting complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies downloaded"

## verify: Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Dependencies verified"

## docker-up: Start Docker services
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "Docker services started"

## docker-down: Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "Docker services stopped"

## docker-logs: View Docker logs
docker-logs:
	docker-compose logs -f

## seed-test-data: Load test data into database for development
seed-test-data:
	@echo "Loading test data into database..."
	@if [ ! -f config.yml ]; then \
		echo "Error: config.yml not found. Copy config.yml.example and configure it first."; \
		exit 1; \
	fi
	@DB_HOST=$$(grep -A10 '^database:' config.yml | grep 'host:' | awk '{print $$2}'); \
	DB_PORT=$$(grep -A10 '^database:' config.yml | grep 'port:' | awk '{print $$2}'); \
	DB_USER=$$(grep -A10 '^database:' config.yml | grep 'user:' | awk '{print $$2}'); \
	DB_NAME=$$(grep -A10 '^database:' config.yml | grep 'dbname:' | awk '{print $$2}'); \
	echo "Connecting to PostgreSQL at $$DB_HOST:$$DB_PORT as $$DB_USER..."; \
	PGPASSWORD=$$(grep -A10 '^database:' config.yml | grep 'password:' | awk '{print $$2}') \
	psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER -d $$DB_NAME -f scripts/seed-test-data.sql
	@echo "Test data loaded successfully!"

## front-install: Install Frontend Node/npm dependencies
front-install:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Frontend dependencies installed"

## front-dev: Start Frontend Vite development server
front-dev:
	@echo "Starting frontend dev server..."
	cd frontend && npm run dev

## front-build: Build Frontend production-ready static assets
front-build:
	@echo "Building frontend static assets..."
	cd frontend && npm run build

## front-lint: Lint Frontend source code using ESLint
front-lint:
	@echo "Linting frontend code..."
	cd frontend && npm run lint

## dev: Compile backend and run both API server and Frontend dev server in parallel
dev: build
	@echo "Starting stalkeer development environment (Backend + Frontend)..."
	@echo "API server: http://localhost:8080"
	@echo "Frontend dashboard: http://localhost:5173"
	(trap 'kill 0' SIGINT; ./bin/$(BINARY_NAME) server & cd frontend && npm run dev)

## dev-frontend: Run only the Frontend dev server
dev-frontend:
	@echo "Starting stalkeer development environment (Backend + Frontend)..."
	@echo "API server: http://localhost:8080"
	@echo "Frontend dashboard: http://localhost:5173"
	(trap 'kill 0' SIGINT; cd frontend && npm run dev)

## dev-backend: Run only the Backend API server
dev-backend: build
	@echo "Starting stalkeer development environment (Backend + Frontend)..."
	@echo "API server: http://localhost:8080"
	(trap 'kill 0' SIGINT; ./bin/$(BINARY_NAME) server)

## docker-build: Docker build (if needed later)
docker-build:
	@echo "Building Docker images..."
	@echo "Backend image: $(REGISTRY)/$(BINARY_NAME)"
	docker build -t $(REGISTRY)/$(BINARY_NAME) .

	@echo "Frontend image: $(REGISTRY)/$(BINARY_NAME)-frontend"
	cd frontend/ && docker build -t $(REGISTRY)/$(BINARY_NAME)-frontend .

## docker-build-versioned: Docker build with version
docker-build-versioned:
	docker build --build-arg VERSION=$(VERSION) -t $(REGISTRY)/$(BINARY_NAME):$(VERSION) -t $(REGISTRY)/$(BINARY_NAME):$(COMMIT) -t $(REGISTRY)/$(BINARY_NAME):latest .
	echo "Docker images built:"
	echo "  $(REGISTRY)/$(BINARY_NAME):$(VERSION)"
	echo "  $(REGISTRY)/$(BINARY_NAME):$(COMMIT)"
	echo "  $(REGISTRY)/$(BINARY_NAME):latest"

	cd frontend/ && docker build --build-arg VERSION=$(VERSION) -t $(REGISTRY)/$(BINARY_NAME)-frontend:$(VERSION) -t $(REGISTRY)/$(BINARY_NAME)-frontend:$(COMMIT) -t $(REGISTRY)/$(BINARY_NAME)-frontend:latest .
	echo "Docker images built for frontend:"
	echo "  $(REGISTRY)/$(BINARY_NAME)-frontend:$(VERSION)"
	echo "  $(REGISTRY)/$(BINARY_NAME)-frontend:$(COMMIT)"
	echo "  $(REGISTRY)/$(BINARY_NAME)-frontend:latest"

## docker-push: Docker push to registry
docker-push:
	docker push $(REGISTRY)/$(BINARY_NAME):$(VERSION)
	docker push $(REGISTRY)/$(BINARY_NAME):$(COMMIT)
	docker push $(REGISTRY)/$(BINARY_NAME):latest

	docker push $(REGISTRY)/$(BINARY_NAME)-frontend:$(VERSION)
	docker push $(REGISTRY)/$(BINARY_NAME)-frontend:$(COMMIT)
	docker push $(REGISTRY)/$(BINARY_NAME)-frontend:latest

## docker-build-push: Docker build and push to registry
docker-build-push: docker-build-versioned docker-push

## db-migrate: Run database migrations
db-migrate:
	@echo "Running database migrations..."
	@./$(BIN_DIR)/$(BINARY_NAME) migrate || echo "Build the application first with 'make build'"

## db-drop-create: Drop and create the database
db-drop-create:
	PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -c "DROP DATABASE stalkeer;" || true
	PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -c "CREATE DATABASE stalkeer;"

## db-truncate-tables: Truncate all main tables in the database
db-truncate-tables:
	PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -d stalkeer -c "TRUNCATE channels, movies, tvshows, uncategorized, processed_lines, processing_logs, download_info RESTART IDENTITY CASCADE;"

## download-sample-videos: Download sample video files for testing
download-sample-videos:
	mkdir -p webserver/html/samples || true
	if [ ! -f webserver/html/samples/sample_FRENCH_360p.mkv ] ; then curl -sL -o webserver/html/samples/sample_FRENCH_360p.mkv https://github.com/ietf-wg-cellar/matroska-test-files/raw/refs/heads/master/test_files/test1.mkv; fi
	if [ ! -f webserver/html/samples/sample_FRENCH_720p.mkv ] ; then curl -sL -o webserver/html/samples/sample_FRENCH_720p.mkv https://github.com/ietf-wg-cellar/matroska-test-files/raw/refs/heads/master/test_files/test1.mkv; fi
	if [ ! -f webserver/html/samples/sample_FRENCH_1080p.mkv ] ; then curl -sL -o webserver/html/samples/sample_FRENCH_1080p.mkv 'https://drive.usercontent.google.com/download?id=0BwxFVkl63-lEWDUzUVUtZEw4cDA&export=download&authuser=0&resourcekey=0-xLf9zGIdfdibsOe8L5JWDg'; fi
	if [ ! -f webserver/html/samples/sample_FRENCH_360p.mp4 ] ; then curl -sL -o webserver/html/samples/sample_FRENCH_360p.mp4 https://filesamples.com/samples/video/mp4/sample_640x360.mp4; fi
	if [ ! -f webserver/html/samples/sample_FRENCH_720p.mp4 ] ; then curl -sL -o webserver/html/samples/sample_FRENCH_720p.mp4 http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4; fi
	if [ ! -f webserver/html/samples/sample_FRENCH_1080p.mp4 ] ; then curl -sL -o webserver/html/samples/sample_FRENCH_1080p.mp4 https://d2qguwbxlx1sbt.cloudfront.net/TextInMotion-VideoSample-1080p.mp4; fi

## prepare-sample-videos: Prepare sample videos for testing
prepare-sample-videos:
	@echo "Preparing sample videos for testing..."
	bash scripts/prepare-samples-videos.sh
	@echo "Sample videos prepared."

## help: Display this help message
help:
	@echo "Stalkeer Makefile Commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
