#!/bin/bash

# StatsD + Prometheus Setup Script
echo "🚀 Setting up StatsD + Prometheus monitoring..."
echo "=============================================="

# Step 1: Create necessary directories
echo "1. Creating directories..."
mkdir -p data/prometheus
mkdir -p data/grafana
mkdir -p config

# Step 2: Start Docker services
echo "2. Starting Docker services..."
docker-compose down
docker-compose up -d

# Step 3: Wait for services
echo "3. Waiting for services to start..."
sleep 15

# Step 4: Check services status
echo "4. Checking services status..."

# Check StatsD Exporter
if curl -s "http://localhost:9102/metrics" | head -1 > /dev/null; then
    echo "✅ StatsD Exporter is running"
else
    echo "❌ StatsD Exporter is not running"
fi

# Check Prometheus
if curl -s "http://localhost:9090/api/v1/query?query=up" > /dev/null; then
    echo "✅ Prometheus is running"
else
    echo "❌ Prometheus is not running"
fi

# Check Grafana
if curl -s "http://localhost:3000/api/health" > /dev/null; then
    echo "✅ Grafana is running"
else
    echo "❌ Grafana is not running"
fi

# Step 5: Start application
echo "5. Starting application server..."
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASSWORD=password
export DATABASE_DBNAME=social_media
export DATABASE_SSLMODE=disable
export SERVER_HOST=localhost
export SERVER_PORT=8080
export JWT_SECRET=your-secret-key
export JWT_EXPIRATION=24

# Start server in background
go run cmd/server/main.go &
APP_PID=$!

# Wait for server
sleep 5

# Step 6: Test application
echo "6. Testing application..."
if curl -s "http://localhost:8080/health" > /dev/null; then
    echo "✅ Application is running"
else
    echo "❌ Application is not running"
    kill $APP_PID
    exit 1
fi

# Step 7: Generate test data
echo "7. Generating test data..."
for i in {1..20}; do
    curl -s "http://localhost:8080/health" > /dev/null
    curl -s "http://localhost:8080/api/account/profile" > /dev/null
    curl -s "http://localhost:8080/api/posts" > /dev/null
    sleep 0.5
done

# Step 8: Check metrics in Prometheus
echo "8. Checking metrics in Prometheus..."
sleep 5

# Check if StatsD metrics are available
if curl -s "http://localhost:9090/api/v1/query?query=statsd_metric" | grep -q "result"; then
    echo "✅ StatsD metrics are available in Prometheus"
else
    echo "⚠️  StatsD metrics not yet available (may take a few minutes)"
fi

# Step 9: Show results
echo ""
echo "🎉 Setup completed!"
echo ""
echo "📊 Access URLs:"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (admin/admin123)"
echo "  - Application: http://localhost:8080"
echo "  - StatsD Exporter: http://localhost:9102/metrics"
echo ""
echo "🔍 Check Prometheus targets: http://localhost:9090/targets"
echo "📈 Check Grafana dashboard: http://localhost:3000"
echo ""
echo "📝 StatsD Metrics Format:"
echo "  - API metrics: app.API_IN.statusCode;group=API_IN;entity=health;HTTPStatus=200"
echo "  - Timing metrics: app.API_IN.execTime;group=API_IN;entity=health"
echo ""
echo "Press Ctrl+C to stop the application server"

# Wait for user to stop
wait $APP_PID
