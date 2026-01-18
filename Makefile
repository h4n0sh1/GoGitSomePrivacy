# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=gogitsomeprivacy
BINARY_PATH=./build/bin/$(BINARY_NAME)

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

.PHONY: all build clean test coverage lint fmt vet install deps help

all: clean deps fmt vet test build

## build: Build the binary
build:
	@echo "Building..."
	@mkdir -p ./build/bin
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/gogitsomeprivacy

## clean: Clean build files
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf ./build/bin
	@rm -f coverage.txt coverage.html

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -timeout 30s ./...

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout 5m ./...; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## install: Install the binary
install:
	@echo "Installing..."
	$(GOCMD) install $(BUILD_FLAGS) $(LDFLAGS) ./cmd/gogitsomeprivacy

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

## tidy: Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

## run: Run the application
run: build
	@echo "Running application..."
	$(BINARY_PATH)

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
