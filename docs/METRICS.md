# Metrics Collection dengan StatsD dan Grafana

Package ini menyediakan sistem monitoring lengkap untuk aplikasi social media service dengan integrasi StatsD dan Grafana.

## Fitur

- **API Metrics**: Monitoring request/response time, error rates, dan throughput
- **Database Metrics**: Monitoring query performance, connection pool, dan transaction metrics
- **System Metrics**: Monitoring memory usage, CPU, dan goroutine count
- **Custom Metrics**: Fleksibilitas untuk metrics custom sesuai kebutuhan

## Komponen

### 1. StatsD Client (`pkg/statsd/`)

Client untuk mengirim metrics ke StatsD server.

### 2. Metrics Package (`pkg/metrics/`)

High-level interface untuk berbagai jenis metrics:

- `APIMetrics`: Metrics untuk HTTP requests
- `DatabaseMetrics`: Metrics untuk database operations
- `SystemMetrics`: Metrics untuk system resources
- `CustomMetric`: Metrics custom

### 3. Database Metrics Wrapper (`pkg/dbmetrics/`)

Wrapper untuk `sql.DB` dan `sql.Tx` yang otomatis mengumpulkan metrics.

### 4. Middleware (`pkg/middleware/metrics.go`)

Middleware untuk mengumpulkan API metrics secara otomatis.

## Konfigurasi

Tambahkan konfigurasi berikut ke environment variables:

```bash
# StatsD Configuration
STATSD_ENABLED=true
STATSD_HOST=localhost
STATSD_PORT=8125
STATSD_PREFIX=social_media
STATSD_SAMPLING=1.0
```

## Penggunaan

### 1. Setup StatsD Server

Install dan jalankan StatsD server:

```bash
# Menggunakan Docker
docker run -d --name statsd -p 8125:8125/udp -p 8126:8126/tcp graphite/statsd

# Atau install via npm
npm install -g statsd
```

### 2. Setup Grafana

Install Grafana dan konfigurasi data source:

```bash
# Menggunakan Docker
docker run -d --name grafana -p 3000:3000 grafana/grafana
```

Konfigurasi data source di Grafana:

- Type: Graphite
- URL: `http://localhost:8080` (jika menggunakan Graphite sebagai backend)
- Access: Server (Default)

### 3. Dashboard Grafana

#### API Metrics Dashboard

Metrics yang tersedia:

- `social_media.api.request.count` - Jumlah request per endpoint
- `social_media.api.request.duration` - Durasi request
- `social_media.api.request.success` - Request berhasil
- `social_media.api.request.error` - Request error
- `social_media.api.response.time` - Response time
- `social_media.api.response.size` - Ukuran response
- `social_media.api.connections.active` - Koneksi aktif

#### Database Metrics Dashboard

Metrics yang tersedia:

- `social_media.db.query.count` - Jumlah query per operation
- `social_media.db.query.duration` - Durasi query
- `social_media.db.query.success` - Query berhasil
- `social_media.db.query.error` - Query error
- `social_media.db.transaction.count` - Jumlah transaction
- `social_media.db.transaction.duration` - Durasi transaction
- `social_media.db.connections.active` - Koneksi aktif
- `social_media.db.connections.idle` - Koneksi idle
- `social_media.db.connections.max` - Maksimal koneksi

#### System Metrics Dashboard

Metrics yang tersedia:

- `social_media.system.memory.used` - Memory usage
- `social_media.system.memory.total` - Total memory
- `social_media.system.memory.usage_percent` - Persentase memory usage
- `social_media.system.cpu.usage` - CPU usage
- `social_media.system.goroutines.count` - Jumlah goroutine

## Contoh Dashboard Queries

### Request Rate per Second

```
alias(summarize(social_media.api.request.count, '1minute', 'sum', true), 'Requests/min')
```

### Average Response Time

```
alias(averageSeries(social_media.api.response.time), 'Avg Response Time')
```

### Error Rate

```
alias(divideSeries(summarize(social_media.api.request.error, '1minute', 'sum', true), summarize(social_media.api.request.count, '1minute', 'sum', true)), 'Error Rate')
```

### Database Query Performance

```
alias(averageSeries(social_media.db.query.duration), 'Avg Query Time')
```

### Memory Usage

```
alias(social_media.system.memory.usage_percent, 'Memory Usage %')
```

## Monitoring Best Practices

1. **Set Alerting**: Buat alert untuk error rate tinggi, response time lambat, dan resource usage tinggi
2. **Dashboard Organization**: Buat dashboard terpisah untuk API, Database, dan System metrics
3. **Retention Policy**: Atur retention policy yang sesuai untuk storage metrics
4. **Sampling**: Gunakan sampling untuk mengurangi volume metrics di production
5. **Tags**: Gunakan tags untuk filtering dan grouping metrics

## Troubleshooting

### StatsD Connection Issues

- Pastikan StatsD server berjalan di port yang benar
- Check firewall settings untuk UDP port 8125
- Verify network connectivity antara aplikasi dan StatsD server

### Metrics Tidak Muncul di Grafana

- Pastikan data source Grafana terkonfigurasi dengan benar
- Check apakah metrics dikirim dengan prefix yang benar
- Verify retention policy di Graphite/StatsD backend

### Performance Impact

- Monitor overhead dari metrics collection
- Adjust sampling rate jika diperlukan
- Consider async metrics sending untuk mengurangi latency

## Contoh Grafana Dashboard JSON

```json
{
  "dashboard": {
    "title": "Social Media Service Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "summarize(social_media.api.request.count, '1minute', 'sum', true)"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "averageSeries(social_media.api.response.time)"
          }
        ]
      }
    ]
  }
}
```

## Dependencies

- StatsD server (untuk mengumpulkan metrics)
- Grafana (untuk visualisasi)
- Graphite atau InfluxDB (sebagai backend storage)

## Support

Untuk pertanyaan atau masalah terkait metrics, silakan buat issue di repository ini.


