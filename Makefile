# Makefile for Flood API
# Usage: make <target>

# Database configuration
DB_NAME = flood
PSQL = psql -d $(DB_NAME)

# Migration paths
MIGRATIONS_DIR = migrations
ANALYSIS_DIR = migrations/analysis

# Docker configuration for integration tests
DOCKER_HOST_VAR = unix://$(HOME)/.colima/default/docker.sock

# Default target
.PHONY: help
help:
	@echo "Flood API Operations"
	@echo "============================="
	@echo ""
	@echo "Available targets:"
	@echo "  migrate-db FROM=XXX       # Run migrations after XXX (3-digit number)"
	@echo "  benchmark-db              # Run performance benchmarks"
	@echo "  test-db-connection        # Test database connection"
	@echo "  db-status                 # Show database status and table info"
	@echo "  sqlc-generate             # Generate Go code from SQL queries"
	@echo "  build                     # Build the application"
	@echo "  build-release             # Build optimized release binary"
	@echo "  run                       # Build and run the application locally"
	@echo "  test                      # Run unit tests"
	@echo "  test-verbose              # Run unit tests with verbose output"
	@echo "  test-coverage             # Run unit tests with coverage report"
	@echo "  test-integration          # Run integration tests (requires Docker)"
	@echo "  test-integration-verbose  # Run integration tests with verbose output"
	@echo "  test-all                  # Run both unit and integration tests"
	@echo "  benchmark     			   # Run performance tests"
	@echo "  clean                     # Clean build artifacts"
	@echo "  fmt                       # Format Go code using go fmt"
	@echo "  vet                       # Run go vet for static analysis"
	@echo "  check                     # Run fmt, vet, build and run all tests (full quality check)"
	@echo ""

# Database status and info
.PHONY: db-status
status:
	@echo "=== Database Status ==="
	@$(PSQL) -c "\dt"
	@echo ""
	@echo "=== Table Record Counts ==="
	@$(PSQL) -c "SELECT 'rainfalls' as table_name, COUNT(*) as record_count FROM public.rainfalls UNION ALL SELECT 'riverlevels' as table_name, COUNT(*) as record_count FROM public.riverlevels UNION ALL SELECT 'stationnames' as table_name, COUNT(*) as record_count FROM public.stationnames;"
	@echo ""
	@echo "=== Indexes ==="
	@$(PSQL) -c "SELECT schemaname, tablename, indexname FROM pg_indexes WHERE schemaname = 'public' ORDER BY tablename, indexname;"

# Migrate from a specific point onwards
# Usage: make migrate-db FROM=001 (runs migrations 002, 003, etc.)
.PHONY: migrate-db
migrate-db:
	@if [ -z "$(FROM)" ]; then \
		echo "Usage: make migrate-db FROM=XXX (where XXX is 3-digit migration number)"; \
		exit 1; \
	fi
	@echo "Running migrations after $(FROM)..."
	@for file in $(MIGRATIONS_DIR)/[0-9][0-9][0-9]_*.sql; do \
		if [ -f "$$file" ]; then \
			filename=$$(basename "$$file"); \
			migration_num=$${filename:0:3}; \
			if [ "$$migration_num" \> "$(FROM)" ]; then \
				echo "Applying migration $$migration_num: $$filename"; \
				$(PSQL) -f "$$file" || exit 1; \
				echo "Migration $$migration_num completed"; \
			fi; \
		fi; \
	done
	@echo "All migrations after $(FROM) completed successfully"


# Performance benchmarking
.PHONY: benchmark-db
benchmark:
	@echo "Running performance benchmarks..."
	@$(PSQL) -f $(ANALYSIS_DIR)/performance_benchmark.sql

.PHONY: test-db-connection
test-connection:
	@echo "Testing database connection..."
	@$(PSQL) -c "SELECT version();" > /dev/null && echo "Database connection successful" || echo "Database connection failed"

# Code generation
.PHONY: sqlc-generate
sqlc-generate:
	@echo "Generating Go code from SQL queries..."
	@sqlc generate
	@echo "Go code generated successfully"

# Build targets
.PHONY: build
build:
	@echo "Building flood-api..."
	@go build -o bin/flood-api ./cmd/flood-api
	@echo "Build completed successfully - binary: bin/flood-api"

.PHONY: build-release
build-release:
	@echo "Building optimized release binary..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/flood-api ./cmd/flood-api
	@echo "Release build completed - binary: bin/flood-api"

.PHONY: run
run: build
	@echo "Starting flood-api on port 9001..."
	@DATABASE_URL="postgres://localhost/flood?sslmode=disable" ./bin/flood-api -port 9001

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "Cleanup completed"

# Test targets
.PHONY: test
test:
	@echo "Running unit tests..."
	@go test ./...
	@echo "All unit tests passed"

.PHONY: test-verbose
test-verbose:
	@echo "Running unit tests with verbose output..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running unit tests with coverage..."
	@go test -cover ./...
	@echo ""
	@echo "Generating detailed coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Integration test targets
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@echo "Checking Docker availability..."
	@if ! docker version > /dev/null 2>&1; then \
		echo "Docker is not running or not available"; \
		echo "Please start Docker/Colima and try again"; \
		echo "For Colima users: colima start"; \
		exit 1; \
	fi
	@echo "Docker is available"
	@echo "Running integration tests with testcontainers..."
	@TESTCONTAINERS_RYUK_DISABLED=true DOCKER_HOST=$(DOCKER_HOST_VAR) go test -tags=integration ./...
	@echo "All integration tests passed"

.PHONY: test-integration-verbose
test-integration-verbose:
	@echo "Running integration tests with verbose output..."
	@echo "Checking Docker availability..."
	@if ! docker version > /dev/null 2>&1; then \
		echo "Docker is not running or not available"; \
		echo "Please start Docker/Colima and try again"; \
		echo "For Colima users: colima start"; \
		exit 1; \
	fi
	@echo "Docker is available"
	@echo "Running integration tests with testcontainers..."
	@TESTCONTAINERS_RYUK_DISABLED=true DOCKER_HOST=$(DOCKER_HOST_VAR) go test -tags=integration -v ./...

.PHONY: test-all
test-all:
	@echo "Running all tests (unit + integration)..."
	@$(MAKE) test
	@$(MAKE) test-integration
	@echo "All tests (unit + integration) passed"

.PHONY: benchmark
benchmark:
	@echo "Running integration-level performance tests (real DB)..."
	@echo "Checking Docker availability..."
	@if ! docker version > /dev/null 2>&1; then \
		echo "Docker is not running or not available"; \
		echo "Please start Docker/Colima and try again"; \
		echo "For Colima users: colima start"; \
		exit 1; \
	fi
	@echo "Docker is available"
	@echo "Running integration performance tests with testcontainers..."
	@TESTCONTAINERS_RYUK_DISABLED=true DOCKER_HOST=$(DOCKER_HOST_VAR) go test -tags=integration -bench=. ./test/integration/
	@echo "Integration performance tests completed"

# Go formatting and code quality
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "Go code formatted successfully"

.PHONY: vet
vet:
	@echo "Running go vet for static analysis..."
	@go vet ./...
	@echo "Static analysis completed successfully"

.PHONY: check
check: fmt vet build test-all
	@echo ""
	@echo "Quality checks passed!"
	@echo "Code formatted"
	@echo "Static analysis passed"
	@echo "Build successful"
	@echo "All tests passed"
	@echo ""
