# Monitoring Stack dengan Docker Compose

Docker Compose setup untuk monitoring Social Media Service menggunakan StatsD, Graphite, dan Grafana.

## Quick Start

### 1. Jalankan Monitoring Stack

```bash
# Start semua services (termasuk PostgreSQL dan monitoring)
docker-compose up -d

# Atau hanya start monitoring services
docker-compose up -d graphite-statsd grafana
```

### 2. Akses Services

- **Grafana Dashboard**: http://localhost:3000

  - Username: `admin`
  - Password: `admin123`

- **Graphite Web Interface**: http://localhost:8080

- **StatsD**:
  - UDP Port: `8125`
  - TCP Admin Port: `8126`

### 3. Konfigurasi Grafana

Dashboard sudah otomatis terkonfigurasi dengan:

- Data source Graphite sudah ter-setup
- Dashboard "Social Media Service Monitoring" sudah tersedia
- Auto-refresh setiap 5 detik

## Services

### StatsD

- **Image**: `graphite/statsd:latest`
- **Port**: 8125 (UDP), 8126 (TCP)
- **Fungsi**: Mengumpulkan metrics dari aplikasi Go

### Graphite

- **Image**: `graphiteapp/graphite-statsd:latest`
- **Port**: 8080 (Web), 2003 (Carbon)
- **Fungsi**: Storage backend untuk metrics

### Grafana

- **Image**: `grafana/grafana:latest`
- **Port**: 3000
- **Fungsi**: Visualisasi dan dashboard metrics

### InfluxDB (Optional)

- **Image**: `influxdb:1.8`
- **Port**: 8086
- **Fungsi**: Alternatif backend storage

## Environment Variables

### Grafana

- `GF_SECURITY_ADMIN_PASSWORD`: Password admin Grafana (default: admin123)
- `GF_USERS_ALLOW_SIGN_UP`: Disable user signup (default: false)

### InfluxDB

- `INFLUXDB_DB`: Database name (default: metrics)
- `INFLUXDB_ADMIN_USER`: Admin username (default: admin)
- `INFLUXDB_ADMIN_PASSWORD`: Admin password (default: admin123)

## Dashboard Metrics

Dashboard menampilkan metrics berikut:

### API Metrics

- **Request Rate**: Jumlah request per menit
- **Response Time**: Rata-rata waktu response
- **Error Rate**: Persentase error request

### Database Metrics

- **Query Time**: Rata-rata waktu eksekusi query
- **Connections**: Jumlah koneksi aktif dan idle

### System Metrics

- **Goroutines**: Jumlah goroutine yang berjalan

## Konfigurasi Aplikasi Go

Pastikan aplikasi Go Anda dikonfigurasi dengan environment variables berikut:

```bash
# StatsD Configuration
STATSD_ENABLED=true
STATSD_HOST=localhost
STATSD_PORT=8125
STATSD_PREFIX=social_media
STATSD_SAMPLING=1.0
```

## Troubleshooting

### StatsD tidak menerima metrics

```bash
# Check StatsD container logs
docker logs social-media-statsd

# Test dengan netcat
echo "test.metric:1|c" | nc -u localhost 8125
```

### Grafana tidak bisa connect ke Graphite

```bash
# Check Graphite container logs
docker logs social-media-graphite

# Test Graphite web interface
curl http://localhost:8080
```

### Dashboard tidak menampilkan data

1. Pastikan aplikasi Go mengirim metrics ke StatsD
2. Check data source configuration di Grafana
3. Verify metrics format di Graphite web interface

## Management Commands

### Stop Services

```bash
docker-compose -f docker-compose.monitoring.yml down
```

### Restart Services

```bash
docker-compose -f docker-compose.monitoring.yml restart
```

### View Logs

```bash
# All services
docker-compose -f docker-compose.monitoring.yml logs

# Specific service
docker-compose -f docker-compose.monitoring.yml logs grafana
```

### Clean Up

```bash
# Stop and remove containers
docker-compose -f docker-compose.monitoring.yml down

# Remove volumes (WARNING: This will delete all data)
docker-compose -f docker-compose.monitoring.yml down -v
```

## Customization

### Menambah Dashboard Baru

1. Buat file JSON dashboard di `grafana/dashboards/`
2. Restart Grafana container
3. Dashboard akan otomatis ter-load

### Mengubah Data Source

1. Edit `grafana/provisioning/datasources/graphite.yml`
2. Restart Grafana container

### Menggunakan InfluxDB

1. Jalankan dengan profile influxdb
2. Update data source di Grafana ke InfluxDB
3. Adjust queries di dashboard

## Monitoring Production

Untuk production environment:

1. **Security**: Ganti password default
2. **Persistence**: Gunakan external volumes
3. **Networking**: Gunakan internal networks
4. **Resources**: Set memory dan CPU limits
5. **Backup**: Backup Grafana dashboards dan data

## Support

Untuk pertanyaan atau masalah terkait monitoring stack, silakan buat issue di repository ini.
