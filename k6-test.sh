#!/bin/bash

# Quick K6 Load Test Runner
# Simple script to run K6 tests from project root

echo "🚀 K6 Load Testing for Social Media Service"
echo "=========================================="

# Check if K6 is installed
if ! command -v k6 &> /dev/null; then
    echo "❌ K6 is not installed. Please install K6 first:"
    echo "   brew install k6"
    exit 1
fi

# Check if API server is running (try multiple ports)
API_PORT=""
for port in 8080 8081 8082; do
    if curl -s http://localhost:$port/health > /dev/null; then
        API_PORT=$port
        break
    fi
done

if [ -z "$API_PORT" ]; then
    echo "❌ API server is not running on any port (8080, 8081, 8082)"
    echo "   Please start the server first:"
    echo "   go run cmd/server/main.go"
    echo "   Or with custom port: SERVER_PORT=8081 go run cmd/server/main.go"
    exit 1
fi

echo "✅ API server is running on port $API_PORT"
echo ""

# Show available tests
echo "Available K6 Tests:"
echo "1. Basic Load Test (100 RPM, 2 minutes)"
echo "2. Multi-Endpoint Test (Multiple APIs)"
echo "3. Stress Test (High load, 10 minutes)"
echo "4. Run Test Suite (Interactive)"
echo ""

read -p "Select test (1-4): " choice

case $choice in
    1)
        echo "📊 Running Basic Load Test..."
        SERVER_PORT=$API_PORT k6 run scripts/k6/k6-load-test.js
        ;;
    2)
        echo "📊 Running Multi-Endpoint Test..."
        SERVER_PORT=$API_PORT k6 run scripts/k6/k6-multi-endpoint-test.js
        ;;
    3)
        echo "📊 Running Stress Test..."
        SERVER_PORT=$API_PORT k6 run scripts/k6/k6-stress-test.js
        ;;
    4)
        echo "📊 Running Test Suite..."
        chmod +x scripts/k6/k6-test-suite.sh
        SERVER_PORT=$API_PORT ./scripts/k6/k6-test-suite.sh
        ;;
    *)
        echo "❌ Invalid choice. Exiting..."
        exit 1
        ;;
esac

echo ""
echo "🎯 Test completed!"
echo ""
echo "📈 View metrics at:"
echo "   - Grafana: http://localhost:3000 (admin/admin123)"
echo "   - InfluxDB: http://localhost:8086 (admin/admin123)"
echo ""
echo "💡 Tips:"
echo "   - Generate more traffic to see metrics in Grafana"
echo "   - Check InfluxDB Data Explorer for raw data"
echo "   - Dashboard: InfluxDB Monitoring Dashboard"

