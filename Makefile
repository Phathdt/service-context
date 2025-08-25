.PHONY: help build test clean lint fmt vet tidy deps run examples

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the module
build: ## Build the module
	@echo "Building..."
	go build ./...

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	go clean -cache -testcache -modcache
	rm -f coverage.out coverage.html

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Run linter (requires golangci-lint)
lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

# Tidy dependencies
tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

# Update dependencies
deps: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Run example applications
run-fiber-example: ## Run the Fiber example application
	@echo "Running Fiber example..."
	cd examples/fiberapp && go run .

# Check for security vulnerabilities
security: ## Check for security vulnerabilities
	@echo "Checking for security vulnerabilities..."
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Full check (format, vet, test)
check: fmt vet test ## Run format, vet, and test

# CI pipeline
ci: tidy fmt vet lint test ## Run all CI checks