# Social Media Service Go

A Go-based social media service API with account management functionality.

## Features

- ✅ Account registration and login
- ✅ Password hashing with bcrypt
- ✅ Standardized API response format
- ✅ Environment-based configuration
- ✅ PostgreSQL database support
- ✅ Clean architecture with repository pattern

## API Endpoints

### Account Management

- `POST /api/account/register` - Register a new account
- `POST /api/account/login` - Login to account
- `GET /health` - Health check endpoint

## Quick Start

### 1. Setup Environment

Copy the sample environment file:

```bash
cp sample-env .env
```

Edit `.env` with your database configuration:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=social_media
DB_SSL_MODE=disable
```

### 2. Setup Database

Run the migration to create the accounts table:

```bash
# Make sure PostgreSQL is running
psql -h localhost -U postgres -d social_media -f migration/sql/000001_create_accounts_table.up.sql
```

### 3. Run the Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### 4. Test the API

Use the provided test script:

```bash
chmod +x scripts/test-api.sh
./scripts/test-api.sh
```

Or test manually with curl:

**Register a new account:**

```bash
curl -X POST http://localhost:8080/api/account/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Login:**

```bash
curl -X POST http://localhost:8080/api/account/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

## API Response Format

All API responses follow this standardized format:

```json
{
  "code": "SUCCESS|FAILED|BAD_REQUEST|UNAUTHORIZED|CONFLICT|INTERNAL_SERVER_ERROR",
  "message": "Human readable message",
  "errors": ["Array of error details"],
  "serverTime": "2024-01-01T00:00:00Z",
  "requestId": "unique-request-id",
  "data": "Response data (varies by endpoint)"
}
```

## Project Structure

```
├── cmd/server/           # Application entry point
├── internal/app/account/ # Account domain
│   ├── app/             # Business logic layer
│   ├── http/            # HTTP handlers
│   └── repo/            # Data access layer
├── infrastructure/config/ # Configuration management
├── pkg/env/             # Environment variable utilities
├── migration/sql/       # Database migrations
├── api/                 # API specifications
└── scripts/             # Utility scripts
```

## Configuration

The application uses environment variables for configuration. See `sample-env` for all available options.

### Key Configuration Variables

- `SERVER_HOST` - Server host (default: localhost)
- `SERVER_PORT` - Server port (default: 8080)
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `JWT_SECRET` - JWT secret key
- `JWT_EXPIRATION` - JWT expiration in hours

## Development

### Dependencies

- Go 1.21+
- PostgreSQL 12+
- Make (optional, for using Makefile)

### Install Dependencies

```bash
go mod tidy
```

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o bin/server cmd/server/main.go
```

## License

MIT License
# social-media-service-go
