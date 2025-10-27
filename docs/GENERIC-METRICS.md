# Generic Metrics Convention

## Overview

This project uses a generic metrics naming convention that can be reused across different teams and services. The metrics are designed to be:

- **Generic**: Not tied to specific business logic
- **Reusable**: Can be used by any team/service
- **Standardized**: Consistent naming across all services
- **Scalable**: Easy to aggregate and analyze

## Metrics Categories

### 1. API Metrics (`API_IN` / `API_OUT`)

#### API_IN (Incoming API Requests)

- `API_IN.count` - Total incoming requests
- `API_IN.success` - Successful requests
- `API_IN.error` - Failed requests
- `API_IN.duration` - Request processing time

#### API_OUT (Outgoing API Responses)

- `API_OUT.response_time` - Response time
- `API_OUT.response_size` - Response size in bytes

### 2. Database Metrics (`DB_*`)

#### DB_QUERY

- `DB_QUERY.count` - Total queries executed
- `DB_QUERY.success` - Successful queries
- `DB_QUERY.error` - Failed queries
- `DB_QUERY.duration` - Query execution time

#### DB_CONNECTION

- `DB_CONNECTION.active` - Active connections
- `DB_CONNECTION.idle` - Idle connections
- `DB_CONNECTION.max` - Maximum connections

#### DB_TRANSACTION

- `DB_TRANSACTION.count` - Total transactions
- `DB_TRANSACTION.success` - Successful transactions
- `DB_TRANSACTION.error` - Failed transactions
- `DB_TRANSACTION.duration` - Transaction duration

### 3. System Metrics (`SYSTEM.*`)

- `SYSTEM.memory.used` - Memory usage in bytes
- `SYSTEM.memory.total` - Total memory in bytes
- `SYSTEM.memory.usage_percent` - Memory usage percentage
- `SYSTEM.cpu.usage` - CPU usage percentage
- `SYSTEM.goroutines.count` - Number of goroutines
- `SYSTEM.gc.pause` - GC pause time in ms
- `SYSTEM.gc.count` - GC count

## Usage Examples

### For Different Services

#### E-commerce Service

```go
// Instead of: ecommerce.order.create.success
metrics.API().RequestSuccess("POST", "orders")

// Results in: stats_counts.{prefix}.API_IN.success
```

#### Payment Service

```go
// Instead of: payment.transaction.process.duration
metrics.Database().QueryStart("SELECT", "transactions")

// Results in: stats.timers.{prefix}.DB_QUERY.duration
```

#### User Service

```go
// Instead of: user.profile.update.error
metrics.API().RequestError("PUT", "profiles", "validation_error")

// Results in: stats_counts.{prefix}.API_IN.error
```

## Tags

All metrics include relevant tags for filtering and aggregation:

### API Metrics Tags

- `method`: HTTP method (GET, POST, PUT, DELETE)
- `path`: Normalized path (e.g., "api.users", "api.orders.id")
- `error_type`: Error classification (client_error, server_error)

### Database Metrics Tags

- `operation`: SQL operation (SELECT, INSERT, UPDATE, DELETE)
- `table`: Database table name
- `error_type`: Database error type

## Dashboard Queries

### Request Rate

```graphite
summarize(stats_counts.{prefix}.API_IN.success, "1minute", "sum", true)
```

### Response Time

```graphite
stats.timers.{prefix}.API_OUT.response_time.mean
```

### Error Rate

```graphite
divideSeries(
  summarize(stats_counts.{prefix}.API_IN.error, "1minute", "sum", true),
  summarize(stats_counts.{prefix}.API_IN.success, "1minute", "sum", true)
)
```

### Database Performance

```graphite
averageSeries(stats.timers.{prefix}.DB_QUERY.duration)
```

## Benefits

1. **Consistency**: Same metrics across all services
2. **Reusability**: Dashboard templates can be reused
3. **Scalability**: Easy to aggregate metrics from multiple services
4. **Maintainability**: Standardized naming reduces confusion
5. **Team Collaboration**: Different teams can use the same conventions

## Migration Guide

### From Service-Specific Metrics

```go
// Old way
metrics.Increment("social_media.posts.create.success")

// New way
metrics.API().RequestSuccess("POST", "posts")
```

### Dashboard Updates

```graphite
# Old
stats_counts.social_mediaapi.request.success

# New
stats_counts.{prefix}.API_IN.success
```

## Configuration

Set the prefix in your environment:

```bash
STATSD_PREFIX=your_service_name
```

This will result in metrics like:

- `stats_counts.your_service_name.API_IN.success`
- `stats.timers.your_service_name.DB_QUERY.duration`
- `your_service_name.SYSTEM.goroutines.count`
