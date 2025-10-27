# Hawkeye Metrics Integration dengan Grafana

## Overview

Package `hawkeye` menyediakan sistem metrics yang terstruktur dan kompatibel dengan Grafana melalui StatsD. Format metrics mengikuti standar: `app.$METRIC_GROUP.$METRIC_NAME`

## Metric Groups yang Tersedia

### API Metrics

- `app.API_IN.execTime` - Waktu eksekusi API inbound
- `app.API_IN.statusCode` - Hitungan status code API inbound
- `app.API_OUT.execTime` - Waktu eksekusi API outbound
- `app.API_OUT.statusCode` - Hitungan status code API outbound

### Database Metrics

- `app.DB.execTime` - Waktu eksekusi query database
- `app.DB.statusCode` - Hitungan operasi database

### Kafka Metrics

- `app.KAFKA_SEND.execTime` - Waktu pengiriman ke Kafka
- `app.KAFKA_SEND.statusCode` - Hitungan pengiriman Kafka
- `app.KAFKA_CONSUMER.execTime` - Waktu konsumsi dari Kafka
- `app.KAFKA_CONSUMER.statusCode` - Hitungan konsumsi Kafka

### System Metrics

- `app.SYSTEM.*` - Metrics sistem (memory, CPU, goroutines)

## Tags yang Tersedia

### Standard Tags

- `group` - Metric group (API_IN, API_OUT, DB, dll)
- `errorCode` - Status (success/failed)
- `HTTPStatus` - HTTP status code
- `protocol` - Protocol (HTTP, gRPC, Kafka, dll)
- `entity` - Entity/endpoint yang diukur

### Custom Tags

- `service` - Nama service
- `env` - Environment (development, staging, production)
- `user_id` - ID user (untuk user-specific metrics)
- `post_id` - ID post (untuk post-specific metrics)

## Konfigurasi Grafana Dashboard

### 1. Data Source Configuration

```yaml
# grafana/provisioning/datasources/graphite.yml
apiVersion: 1

datasources:
  - name: Graphite
    type: graphite
    url: http://graphite:80
    access: proxy
    isDefault: true
```

### 2. Dashboard Queries

#### API Response Time

```graphite
aliasByNode(app.API_IN.execTime.*, 3)
```

#### API Success Rate

```graphite
aliasByNode(divideSeries(app.API_IN.statusCode.*.success, app.API_IN.statusCode.*), 3)
```

#### Database Query Performance

```graphite
aliasByNode(app.DB.execTime.*, 3)
```

#### Error Rate by Endpoint

```graphite
aliasByNode(app.API_IN.statusCode.*.failed, 3)
```

### 3. Grafana Dashboard Panels

#### Panel 1: API Response Time

- **Query**: `aliasByNode(app.API_IN.execTime.*, 3)`
- **Visualization**: Time Series
- **Unit**: Milliseconds
- **Legend**: `{{entity}}`

#### Panel 2: API Success Rate

- **Query**: `aliasByNode(divideSeries(app.API_IN.statusCode.*.success, app.API_IN.statusCode.*), 3)`
- **Visualization**: Stat
- **Unit**: Percent (0-100)
- **Legend**: `{{entity}}`

#### Panel 3: Database Performance

- **Query**: `aliasByNode(app.DB.execTime.*, 3)`
- **Visualization**: Time Series
- **Unit**: Milliseconds
- **Legend**: `{{entity}}`

#### Panel 4: Error Rate

- **Query**: `aliasByNode(app.API_IN.statusCode.*.failed, 3)`
- **Visualization**: Time Series
- **Unit**: Count
- **Legend**: `{{entity}}`

## Environment Variables

```bash
# StatsD Configuration
STATSD_HOST=localhost
STATSD_PORT=8125

# Service Configuration
SERVICE_NAME=social-media-service
ENVIRONMENT=development
METRICS_PREFIX=social-media-service
```

## Contoh Penggunaan

### 1. Setup Metrics

```go
hawkeyeAdapter, err := hawkeye.SetupHawkeyeMetrics()
if err != nil {
    log.Fatal(err)
}
```

### 2. HTTP Middleware

```go
handler := hawkeye.HawkeyeMiddleware(hawkeyeAdapter)(mux)
```

### 3. Custom Metrics

```go
// Record user registration
hawkeyeAdapter.RecordUserRegistration(true, duration)

// Record post creation
hawkeyeAdapter.RecordPostCreation(true, duration, "user123")

// Record database operation
hawkeyeAdapter.RecordDatabaseMetric("SELECT", "users", duration, true)
```

## Monitoring & Alerting

### Grafana Alerts

#### High Error Rate

- **Condition**: `app.API_IN.statusCode.*.failed` > 10 dalam 5 menit
- **Action**: Send notification ke Slack/Email

#### Slow Response Time

- **Condition**: `app.API_IN.execTime.*` > 1000ms dalam 5 menit
- **Action**: Send notification ke Slack/Email

#### Database Performance

- **Condition**: `app.DB.execTime.*` > 500ms dalam 5 menit
- **Action**: Send notification ke Slack/Email

## Best Practices

1. **Metric Naming**: Gunakan format `app.$GROUP.$METRIC` yang konsisten
2. **Tagging**: Gunakan tags yang relevan untuk filtering dan grouping
3. **Sampling**: Gunakan sampling rate yang sesuai untuk mengurangi overhead
4. **Monitoring**: Setup alerting untuk metrics yang critical
5. **Dashboard**: Buat dashboard yang mudah dipahami oleh tim

## Troubleshooting

### Metrics Tidak Muncul di Grafana

1. Cek koneksi StatsD: `telnet localhost 8125`
2. Cek format metrics di StatsD logs
3. Cek Grafana data source configuration
4. Cek query syntax di Grafana

### Performance Issues

1. Kurangi sampling rate jika perlu
2. Gunakan async sending untuk metrics
3. Monitor memory usage dari StatsD client
4. Consider menggunakan batch sending
