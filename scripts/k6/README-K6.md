# K6 Load Testing Scripts

Scripts untuk melakukan load testing pada Social Media Service API menggunakan K6.

## 📋 Prerequisites

### 1. Install K6

```bash
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Windows
winget install k6
```

### 2. Start Services

```bash
# Start API server
go run cmd/server/main.go

# Start monitoring stack
docker-compose up -d graphite-statsd grafana
```

## 🚀 Available Tests

### 1. Basic Load Test (`k6-load-test.js`)

- **Target**: 100 requests per minute
- **Duration**: 2 minutes
- **Endpoint**: GET /api/posts?limit=20
- **Users**: 10 virtual users

```bash
# Run basic test
./scripts/run-load-test.sh

# Or directly with K6
k6 run scripts/k6-load-test.js
```

### 2. Multi-Endpoint Test (`k6-multi-endpoint-test.js`)

- **Target**: 100 requests per minute
- **Duration**: 2.5 minutes
- **Endpoints**: Multiple APIs with weighted distribution
- **Users**: 10 virtual users

**Endpoint Distribution:**

- GET /api/posts: 60% of requests
- GET /api/account/profile: 20% of requests
- GET /api/comments: 20% of requests

```bash
k6 run scripts/k6-multi-endpoint-test.js
```

### 3. Stress Test (`k6-stress-test.js`)

- **Target**: High load with spikes
- **Duration**: 10 minutes
- **Users**: Up to 100 virtual users
- **Pattern**: Gradual ramp-up with spike testing

**Load Pattern:**

- 0-1m: Ramp up to 20 users
- 1-3m: Ramp up to 50 users
- 3-6m: Stay at 50 users
- 6-7m: Spike to 100 users
- 7-9m: Ramp down to 50 users
- 9-10m: Ramp down to 0 users

```bash
k6 run scripts/k6-stress-test.js
```

## 🎯 Test Suite

### Interactive Test Suite (`k6-test-suite.sh`)

Script interaktif untuk menjalankan berbagai jenis test:

```bash
chmod +x scripts/k6-test-suite.sh
./scripts/k6-test-suite.sh
```

**Options:**

1. Basic Load Test (100 RPM for 2 minutes)
2. Multi-Endpoint Test (Multiple APIs)
3. Stress Test (High load)
4. All Tests
5. Custom Test

## 📊 Monitoring During Tests

### Real-time Monitoring

- **Grafana Dashboard**: http://localhost:3000 (admin/admin123)
- **Graphite Web UI**: http://localhost:8080

### Key Metrics to Watch

- **Request Rate**: Requests per second
- **Response Time**: Average, median, 95th percentile
- **Error Rate**: Percentage of failed requests
- **Database Performance**: Query execution time
- **System Resources**: CPU, memory usage

## 📈 Test Results

### Output Files

- `k6-results.json`: Detailed test results
- `k6-stress-results.json`: Stress test results

### Sample Output

```
========================================
K6 Load Test Results Summary
========================================
Total Requests: 200
Failed Requests: 5
Success Rate: 97.50%

Response Times:
- Average: 45.23ms
- Median: 42.10ms
- 95th percentile: 89.45ms
- 99th percentile: 156.78ms

Requests per second: 1.67
========================================
```

## 🔧 Customization

### Custom Test Parameters

```bash
# Custom VUs and duration
k6 run --vus 20 --duration 5m scripts/k6-load-test.js

# Custom ramp-up pattern
k6 run --stage 30s:10,1m:20,30s:0 scripts/k6-load-test.js
```

### Modify Test Scenarios

Edit the test files to:

- Change endpoint URLs
- Adjust request weights
- Modify test duration
- Add custom headers
- Change success criteria

## 🚨 Troubleshooting

### Common Issues

1. **API Server Not Running**

   ```bash
   # Check if server is running
   curl http://localhost:8080/health

   # Start server if needed
   go run cmd/server/main.go
   ```

2. **K6 Not Installed**

   ```bash
   # Install K6
   brew install k6
   ```

3. **High Error Rate**

   - Check API server logs
   - Verify database connection
   - Monitor system resources
   - Reduce load if necessary

4. **Slow Response Times**
   - Check database performance
   - Monitor system resources
   - Optimize API endpoints
   - Check network latency

### Performance Tips

1. **Before Testing**

   - Ensure monitoring stack is running
   - Check system resources
   - Verify API endpoints are working

2. **During Testing**

   - Monitor Grafana dashboard
   - Watch for error spikes
   - Check database performance

3. **After Testing**
   - Analyze results
   - Identify bottlenecks
   - Optimize based on findings

## 📚 Additional Resources

- [K6 Documentation](https://k6.io/docs/)
- [K6 JavaScript API](https://k6.io/docs/javascript-api/)
- [Load Testing Best Practices](https://k6.io/docs/testing-guides/)
- [Grafana Dashboard Guide](docs/MONITORING-DOCKER.md)
