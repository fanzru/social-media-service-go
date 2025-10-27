#!/bin/bash

# InfluxDB Setup Script
echo "🚀 Setting up InfluxDB monitoring..."
echo "===================================="

# Step 1: Create necessary directories
echo "1. Creating directories..."
mkdir -p data/influxdb
mkdir -p data/grafana

# Step 2: Start Docker services
echo "2. Starting Docker services..."
docker-compose down
docker-compose up -d

# Step 3: Wait for services
echo "3. Waiting for services to start..."
sleep 15

# Step 4: Check services status
echo "4. Checking services status..."

# Check InfluxDB
if curl -s "http://localhost:8086/health" | grep -q "ready"; then
    echo "✅ InfluxDB is running"
else
    echo "❌ InfluxDB is not running"
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
export S3_ACCESS_KEY_ID=dummy
export S3_SECRET_ACCESS_KEY=dummy
export S3_BUCKET_NAME=dummy
export S3_REGION=us-east-1

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

# Step 8: Check metrics in InfluxDB
echo "8. Checking metrics in InfluxDB..."
sleep 5

# Check if metrics are available in InfluxDB
if curl -s -H "Authorization: Token my-super-secret-auth-token" "http://localhost:8086/api/v2/query?org=social-media" -d 'from(bucket:"metrics") |> range(start:-1h) |> filter(fn:(r) => r._measurement == "http_requests_total") |> count()' | grep -q "http_requests_total"; then
    echo "✅ Metrics are available in InfluxDB"
else
    echo "⚠️  Metrics not yet available (may take a few minutes)"
fi

# Step 9: Show results
echo ""
echo "🎉 Setup completed!"
echo ""
echo "📊 Access URLs:"
echo "  - InfluxDB: http://localhost:8086 (admin/admin123)"
echo "  - Grafana: http://localhost:3000 (admin/admin123)"
echo "  - Application: http://localhost:8080"
echo ""
echo "🔍 Check InfluxDB data: http://localhost:8086"
echo "📈 Check Grafana dashboard: http://localhost:3000"
echo ""
echo "📝 InfluxDB Metrics Format:"
echo "  - Measurement: http_requests_total"
echo "  - Tags: group=API_IN, entity=health, method=GET, http_status=200"
echo "  - Fields: value=1"
echo ""
echo "Press Ctrl+C to stop the application server"

# Wait for user to stop
wait $APP_PID
