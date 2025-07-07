# Makefile for Flood API Database Operations
# Usage: make <target>

# Database configuration
DB_NAME = flood
PSQL = psql -d $(DB_NAME)

# Migration paths
MIGRATIONS_DIR = migrations
ANALYSIS_DIR = migrations/analysis

# Default target
.PHONY: help
help:
	@echo "Flood API Database Operations"
	@echo "============================="
	@echo ""
	@echo "Available targets:"
	@echo "  migrate-db FROM=XXX - Run migrations after XXX (3-digit number)"
	@echo "  benchmark-db        - Run performance benchmarks"
	@echo "  test-db-connection  - Test database connection"
	@echo ""
	@echo "Migration examples:"
	@echo "  make migrate-db FROM=001  # Run migrations 002, 003, etc."
	@echo "  make migrate-db FROM=000  # Run all migrations"
	@echo ""
	@echo "Workflow example:"
	@echo "  make test-connection      # Verify DB is accessible"
	@echo "  make migrate-db FROM=000  # Apply all migrations"
	@echo "  make benchmark            # Test performance"

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
	@echo "🔌 Testing database connection..."
	@$(PSQL) -c "SELECT version();" > /dev/null && echo "Database connection successful" || echo "Database connection failed"
