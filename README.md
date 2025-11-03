# Social Media Service Go

A Go-based social media service API with account management functionality.

## Requirements

- Functional
  - Create posts with images
  - Set text caption on post creation
  - Comment on a post
  - List all posts sorted by number of comments (desc)
  - Provide last 2 comments per post in listings (design target)
  - Cursor-based pagination for posts and comments
  - GDPR delete: delete an account and all posts, images, and comments
  - RESTful JSON API
- Non-Functional
  - Max image size: 100MB
  - Allowed image formats: .png, .jpg, .bmp
  - Save original image, also convert to .jpg and resize to 600x600; serve images as .jpg
  - Max response time (except uploads): 50 ms
  - Minimum throughput: 100 RPS
  - Handle slow/unstable client connections
- Usage Forecast
  - ~1k images/hour uploaded
  - ~100k comments/hour

## Features

- ✅ Account registration and login
- ✅ Password hashing with bcrypt
- ✅ Standardized API response format
- ✅ Environment-based configuration
- ✅ PostgreSQL database support
- ✅ Clean architecture with repository pattern
- ✅ StatsD metrics collection
- ✅ Grafana monitoring dashboard
- ✅ K6 load testing suite
- ✅ Image upload with processing (resize + JPG), original retained
- ✅ Posts listing sorted by comment count with cursor-based pagination

Notes:

- Posts endpoint exposes comment_count and stable cursors; returning the last 2 comments per post is supported in the design and can be added via lightweight join or follow-up calls.

## API Endpoints

### Account Management

- `POST /api/account/register` - Register a new account
- `POST /api/account/login` - Login to account
- `GET /api/account/profile` - Get authenticated account profile
- `DELETE /api/account` - GDPR delete self (deletes account, posts, images, comments)
- `GET /health` - Health check endpoint

### Posts & Images

- `POST /api/posts` - Create post with image (multipart/form-data)

  - Fields:
    - `caption` (string, required)
    - `image` (file, required)
  - Image rules (company requirements):
    - Max size: 100MB
    - Allowed formats: `.png`, `.jpg`, `.bmp`
    - Original image is saved in its original format to storage
    - Processed image is converted to `.jpg` and resized to `600x600`
    - API serves images only as `.jpg`

- `GET /api/posts` - List posts sorted by number of comments (desc) with cursor-based pagination

  - Query params:
    - `cursor` (string, optional) — composite cursor encoding `comment_count|created_at` using URL-safe Base64
    - `limit` (int, optional, default 20, max 100)
  - Sort order: `comment_count DESC, created_at DESC`
  - Response includes `cursor` (next page token) and `has_more`

- `GET /api/posts/by-user/{userId}` - List posts by user (cursor-based)
- `GET /api/posts/{id}` - Get post by ID
- `PUT /api/posts/{id}` - Update post (e.g., caption)
- `DELETE /api/posts/{id}` - Delete a post (cascades comments, images)

### Comments

- `POST /api/comments/by-post/{postId}` - Create a comment on a post
- `GET /api/comments/by-post/{postId}` - List comments on a post (cursor-based)
  - Query params: `cursor`, `limit`
- `GET /api/comments/{id}` - Get comment by ID
- `PUT /api/comments/{id}` - Update a comment
- `DELETE /api/comments/{id}` - Delete a comment
- `GET /api/comments/user/{userId}` - List comments by user (cursor-based)

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

## Domain Model

- Account: `id`, `name`
- Post: `id`, `caption`, `image`, `creator (account)`, `created_at`, `comments (list)`, `comment_count`
- Comment: `id`, `content`, `creator (account)`, `created_at`

## System Design (Concise)

- API Layer: Go HTTP handlers generated from OpenAPI specs (`api/http/*.yaml`), standardized responses in `pkg/response` and request context in `pkg/reqctx`.
- Auth: JWT middleware in `pkg/middleware/auth.go` with endpoint-specific security map.
- Business Logic: Clean architecture modules under `internal/app/{account,post,comment}` with `app` (services), `repo` (data), and `port/http` (transport) layers.
- Storage:
  - Database: PostgreSQL for accounts, posts, comments. Cursor-based pagination uses stable composite cursors (e.g., `comment_count|created_at`) encoded URL-safe Base64.
  - Images: Original uploaded file persisted and a processed `.jpg` (600x600, configurable quality) saved via `pkg/storage` and/or `pkg/s3` client. API serves `.jpg` URLs.
- Observability:
  - Metrics: StatsD and Prometheus metrics (`pkg/statsd`, `pkg/prometheus`), dashboards with Grafana (`grafana/`).
  - Traces/Logs: Structured logging in `pkg/logger`; example Grafana plugins included.
- Performance Targets:
  - <50 ms p95 for non-upload APIs using prepared queries, indexes on `(comment_count desc, created_at desc)` materialized view or CTE, connection pooling.
  - 100 RPS baseline with horizontal scaling behind a reverse proxy; image uploads use streaming and size limits.
- Resilience:
  - Idempotent handlers where relevant, input validation, timeouts, and safe deletion flow for GDPR (service method cascades delete of posts, images, comments).

Implementation status:

- Implemented: registration, login, profile, create/list posts, comments CRUD, cursor-based pagination, posts sorted by comment count, GDPR account delete, image processing and serving rules.
- Design target: include last 2 comments in posts listing response (can be added with additional query/aggregation if needed).

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

### Storage & Image Processing Configuration

- `MAX_FILE_SIZE` — Max upload size in bytes (default: `104857600` = 100MB)
- `ALLOWED_EXTENSIONS` — Allowed file extensions (default: `.png,.jpg,.bmp`)
- `IMAGE_RESIZE_WIDTH` — Processed image width (default: 600)
- `IMAGE_RESIZE_HEIGHT` — Processed image height (default: 600)
- `IMAGE_QUALITY` — JPEG quality 1-100 (default: 85)
- `S3_REGION` — S3/R2 region (default: `auto`)
- `S3_BUCKET` — Bucket name
- `S3_ACCESS_KEY_ID` — Access key
- `S3_SECRET_ACCESS_KEY` — Secret key
- `S3_ENDPOINT` — Optional custom endpoint (e.g., Cloudflare R2)
- `S3_IMAGE_BASE_URL` — Public base URL for serving images

Notes:

- The service uploads two files per post image:
  - Original: `post_<timestamp>_orig.<ext>` (content-type based on original extension)
  - Processed: `post_<timestamp>.jpg` (content-type `image/jpeg`)
- Deletion attempts to remove both processed and original variants.

### Cursor-Based Pagination (Posts Sorted by Comments)

- Composite cursor ensures stable pagination when multiple posts share the same comment count.
- Format (conceptual): `comment_count|created_at` encoded with URL-safe Base64 (no padding).
- Example flow:
  1. Call `GET /api/posts?limit=20`
  2. Use `cursor` from response for the next page: `GET /api/posts?cursor=<token>&limit=20`

### Performance Guidance

- Keep `limit` reasonably small (e.g., 20-50) for most endpoints to ensure <50 ms p95.
- Use pagination for comments and avoid N+1 by batching fetches when enriching posts with last comments.

## API Documentation (OpenAPI)

- Source specs: `api/http/*.yaml`
- Generated swagger docs: `docs/swagger/docs.json`
- Regenerate: `./scripts/swaggerdocs.sh`

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

## 📊 Monitoring & Load Testing

### Monitoring Stack

Start the monitoring stack with StatsD, Graphite, and Grafana:

```bash
# Start monitoring services
docker-compose up -d graphite-statsd grafana

# Access monitoring
# Grafana: http://localhost:3000 (admin/admin123)
# Graphite: http://localhost:8080
```

### Load Testing with K6

Install K6 and run load tests:

```bash
# Install K6 (if not already installed)
./install-k6.sh

# Run quick test
./k6-test.sh

# Run all tests
./run-all-tests.sh

# Run specific test
k6 run scripts/k6/k6-load-test.js
```

**Available Tests:**

- **Basic Load Test**: 100 requests per minute for 2 minutes
- **Multi-Endpoint Test**: Multiple APIs with weighted distribution
- **Stress Test**: High load with up to 100 users for 10 minutes

**Test Results:**

- Real-time metrics in Grafana dashboard
- Detailed results in JSON files
- Performance analysis and recommendations

### Metrics Collected

- **API Metrics**: Request rate, response time, error rate
- **Database Metrics**: Query performance, connection pool
- **System Metrics**: Memory usage, CPU, goroutines

## 📚 Documentation

- [Monitoring Setup](docs/MONITORING-DOCKER.md)
- [K6 Load Testing](scripts/k6/README-K6.md)
- [Metrics Documentation](docs/METRICS.md)

## License

MIT License
