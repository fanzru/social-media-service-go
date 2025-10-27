# 📊 Prometheus Monitoring Dashboard

Dashboard monitoring yang modern menggunakan Prometheus untuk aplikasi social media service.

## 🎯 Fitur Dashboard

### **Prometheus Monitoring Dashboard** (`prometheus-monitoring-dashboard.json`)

- **API Request Rate**: Success vs Error requests per detik
- **API Response Time**: 95th dan 50th percentile response time
- **Total Success Requests**: Jumlah total request yang berhasil (1 jam)
- **Total Error Requests**: Jumlah total request yang gagal (1 jam)
- **Success Rate**: Persentase keberhasilan request
- **Total Requests**: Total semua request (1 jam)
- **API Requests by Endpoint**: Request per endpoint API
- **Database Queries by Operation**: Query database per operasi

## 📈 Metrik yang Dimonitor

### HTTP Metrics

```prometheus
# Total HTTP requests
http_requests_total{method="GET", endpoint="api/account/profile", status="2xx"}

# HTTP request duration
http_request_duration_seconds{method="GET", endpoint="api/account/profile"}

# HTTP response size
http_response_size_bytes{method="GET", endpoint="api/account/profile"}
```

### Database Metrics

```prometheus
# Total database queries
db_queries_total{operation="SELECT", table="accounts", status="success"}

# Database query duration
db_query_duration_seconds{operation="SELECT", table="accounts"}
```

### System Metrics

```prometheus
# Active connections
active_connections

# Node exporter metrics (CPU, Memory, Disk, etc.)
node_cpu_seconds_total
node_memory_MemAvailable_bytes
node_filesystem_avail_bytes
```

## 🚀 Cara Setup

### 1. Start Services

```bash
# Start monitoring services
docker-compose up -d prometheus node-exporter grafana
```

### 2. Setup Dashboard

```bash
# Run setup script
chmod +x scripts/setup-prometheus-monitoring.sh
./scripts/setup-prometheus-monitoring.sh
```

### 3. Start Application Server

```bash
# Start server with monitoring
chmod +x scripts/run-server-with-prometheus.sh
./scripts/run-server-with-prometheus.sh
```

## 🔗 URLs

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin123)
- **Node Exporter**: http://localhost:9100
- **Application Metrics**: http://localhost:8080/metrics
- **Dashboard**: http://localhost:3000/d/prometheus-monitoring/prometheus-monitoring-dashboard

## 📊 Query Examples

### API Success Rate

```prometheus
# Success requests rate
rate(http_requests_total{job="social-media-app",status=~"2.."}[5m])

# Error requests rate
rate(http_requests_total{job="social-media-app",status=~"4..|5.."}[5m])

# Success rate percentage
sum(rate(http_requests_total{job="social-media-app",status=~"2.."}[5m])) / sum(rate(http_requests_total{job="social-media-app"}[5m])) * 100
```

### API Response Time

```prometheus
# 95th percentile response time
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="social-media-app"}[5m]))

# 50th percentile response time
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{job="social-media-app"}[5m]))
```

### Database Metrics

```prometheus
# Database query rate
rate(db_queries_total{job="social-media-app"}[5m])

# Database query duration
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket{job="social-media-app"}[5m]))
```

### System Metrics

```prometheus
# CPU usage
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Memory usage
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# Disk usage
(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100
```

## 🎨 Dashboard Panels

### 1. **Time Series Charts**

- API Request Rate (Success vs Error)
- API Response Time (95th & 50th percentile)
- API Requests by Endpoint
- Database Queries by Operation

### 2. **Stat Panels**

- Total Success Requests (1h)
- Total Error Requests (1h)
- Success Rate Percentage
- Total Requests (1h)

## 🔧 Configuration

### Prometheus Settings

- **Scrape Interval**: 15 seconds
- **Evaluation Interval**: 15 seconds
- **Retention Time**: 200 hours
- **Targets**:
  - Prometheus itself
  - Node Exporter
  - Social Media App

### Grafana Settings

- **Refresh Rate**: 5 seconds
- **Time Range**: Last 1 hour
- **Theme**: Dark
- **Timezone**: Local

## 📝 Prometheus Configuration

### `config/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]

  - job_name: "node-exporter"
    static_configs:
      - targets: ["node-exporter:9100"]

  - job_name: "social-media-app"
    static_configs:
      - targets: ["host.docker.internal:8080"]
    metrics_path: "/metrics"
    scrape_interval: 5s
```

## 🚨 Troubleshooting

### Dashboard tidak menampilkan data

1. Pastikan aplikasi server sudah running
2. Pastikan endpoint `/metrics` accessible
3. Check Prometheus targets di http://localhost:9090/targets
4. Verify metrics format di http://localhost:8080/metrics

### Prometheus tidak bisa scrape metrics

1. Pastikan aplikasi server running di port 8080
2. Check Prometheus configuration
3. Verify network connectivity
4. Check Prometheus logs

### Grafana tidak bisa connect ke Prometheus

1. Pastikan Prometheus service running
2. Check datasource URL di Grafana
3. Test connection di Grafana datasource settings

## 🎉 Keunggulan Prometheus

### Dibandingkan Graphite:

- ✅ **Modern**: Lebih modern dan aktif dikembangkan
- ✅ **Query Language**: PromQL lebih powerful
- ✅ **Service Discovery**: Auto-discovery untuk targets
- ✅ **Alerting**: Built-in alerting system
- ✅ **Ecosystem**: Lebih banyak integrations
- ✅ **Performance**: Lebih efisien untuk time series data

### Fitur Tambahan:

- **Alerting Rules**: Bisa setup alerting
- **Recording Rules**: Pre-computed metrics
- **Federation**: Multi-level Prometheus setup
- **Long-term Storage**: Integration dengan remote storage

## 🎯 Next Steps

1. **Setup Alerting**: Buat alerting rules untuk monitoring
2. **Add More Metrics**: Tambah custom business metrics
3. **Service Discovery**: Setup auto-discovery untuk dynamic targets
4. **Long-term Storage**: Setup remote storage untuk data retention
5. **Multi-environment**: Setup Prometheus untuk multiple environments

## 🎉 Happy Monitoring!

Dashboard Prometheus memberikan visibility yang excellent terhadap:

- ✅ **Application Performance**: API response times, error rates
- ✅ **System Performance**: CPU, Memory, Disk usage
- ✅ **Database Performance**: Query performance, connection pools
- ✅ **Business Metrics**: Custom application metrics
- ✅ **Infrastructure**: Server health, network metrics

Silakan customize dashboard sesuai kebutuhan aplikasi Anda!
