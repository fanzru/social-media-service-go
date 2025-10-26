


# Load environment variables
include .env
export

# Construct database URL from environment variables
DB_URL = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: migrate-up migrate-down migrate-force migrate-version migrate-create build run deps test clean http-gen

# Run all pending migrations
migrate-up:
	migrate -path migration/sql -database "$(DB_URL)" up

# Rollback the last migration
migrate-down:
	migrate -path migration/sql -database "$(DB_URL)" down 1

# Force set migration version (use with caution)
migrate-force:
	migrate -path migration/sql -database "$(DB_URL)" force $(VERSION)

# Check current migration version
migrate-version:
	migrate -path migration/sql -database "$(DB_URL)" version

# Create new migration file
migrate-create:
	migrate create -ext sql -dir migration/sql -seq $(NAME)

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Generate HTTP code from OpenAPI specs
http-gen:
	scripts/http.sh

# Generate Swagger documentation from OpenAPI specs
swagger-gen:
	scripts/swaggerdocs.sh

# Show current database configuration
db-info:
	@echo "Database Configuration:"
	@echo "  Host: $(DB_HOST)"
	@echo "  Port: $(DB_PORT)"
	@echo "  User: $(DB_USER)"
	@echo "  Database: $(DB_NAME)"
	@echo "  SSL Mode: $(DB_SSL_MODE)"
	@echo "  URL: $(DB_URL)"

# Setup development environment
dev-setup: deps migrate-up
	@echo "Development environment setup complete!"

# Production build
prod-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go