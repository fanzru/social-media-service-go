


# Load environment variables
include .env
export

# Construct database URL from environment variables
DB_URL = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: migrate-up migrate-down migrate-force migrate-version migrate-create build run deps test clean http-gen
.PHONY: reset-timeseries reset-timeseries-all init-timeseries

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

gen:
	make http-gen
	make swagger-gen

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

# Docker Compose commands
docker-check-env:
	@if [ ! -f .env ]; then \
		echo "❌ Error: .env file not found!"; \
		echo "💡 Creating .env from sample-env..."; \
		cp sample-env .env; \
		echo "✅ Please edit .env file with your configuration before running again."; \
		exit 1; \
	fi

docker-up: docker-check-env
	docker-compose up -d
	@echo "✅ Services started!"
	@echo "📊 Grafana: http://localhost:3000 (admin/admin123)"
	@echo "🚀 App: http://localhost:8080"
	@echo "📈 InfluxDB: http://localhost:8086 (admin/admin123)"

docker-down:
	docker-compose down

docker-build: docker-check-env
	docker-compose up -d --build
	@echo "✅ Services built and started!"
	@echo "📊 Grafana: http://localhost:3000 (admin/admin123)"
	@echo "🚀 App: http://localhost:8080"
	@echo "📈 InfluxDB: http://localhost:8086 (admin/admin123)"

docker-logs:
	docker-compose logs -f

docker-logs-app:
	docker-compose logs -f app

docker-logs-monitor:
	docker logs -f social-media-monitor

docker-logs-influxdb:
	docker logs -f social-media-influxdb

docker-check-monitor:
	@echo "📊 Checking monitor service status..."
	@docker ps --filter "name=social-media-monitor" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "📝 Recent monitor logs:"
	@docker logs --tail 20 social-media-monitor || echo "Monitor container not running"

docker-ps:
	docker-compose ps

docker-restart: docker-down docker-up

# Start all services with monitoring
start-all: docker-up
	@echo "📝 Monitor logs: make docker-logs-monitor"

# Build Docker image
docker-build-image:
	docker build -t social-media-app:latest -f Dockerfile .

# Run Docker container with .env file (standalone, not with docker-compose)
docker-run:
	@if [ ! -f .env ]; then \
		echo "❌ Error: .env file not found!"; \
		echo "💡 Creating .env from sample-env..."; \
		cp sample-env .env; \
		echo "✅ Please edit .env file with your configuration before running again."; \
		exit 1; \
	fi
	docker run -d \
		--name social-media-app \
		--env-file .env \
		-e SERVER_HOST=0.0.0.0 \
		-p 8080:8080 \
		--restart unless-stopped \
		social-media-app:latest
	@echo "✅ Container started!"
	@echo "🚀 App: http://localhost:8080"
	@echo "📝 View logs: docker logs -f social-media-app"

# Stop and remove standalone Docker container
docker-stop:
	docker stop social-media-app || true
	docker rm social-media-app || true

# Reset local time series data (InfluxDB + Prometheus)
reset-timeseries:
	@echo "⚠️  This will DELETE all local time series data (InfluxDB + Prometheus)."
	@echo "    Make sure services are stopped before running."
	@read -p "Proceed? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
	  rm -rf \
	    "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/data/influxdb/"* \
	    "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/data/prometheus/"* ; \
	  echo "✅ Time series data cleared."; \
	else \
	  echo "❎ Cancelled."; \
	fi

# Reset time series AND Grafana SQLite DB (fully clean monitoring stack)
reset-timeseries-all:
	@echo "⚠️  This will DELETE InfluxDB data, Prometheus data, and Grafana DB."
	@echo "    You will lose Grafana users/datasources/dashboards stored in grafana.db."
	@read -p "Proceed? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
	  rm -rf \
	    "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/data/influxdb/"* \
	    "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/data/prometheus/"* ; \
	  rm -f \
	    "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/data/grafana/grafana.db" ; \
	  echo "✅ Monitoring data fully cleared."; \
	else \
	  echo "❎ Cancelled."; \
	fi

# Re-initialize time series stack (InfluxDB buckets, StatsD/Prometheus mapping)
init-timeseries:
	@bash "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/scripts/setup-influxdb.sh" || true
	@bash "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/scripts/setup-statsd-prometheus.sh" || true
	@echo "✅ Time series initialization scripts executed."