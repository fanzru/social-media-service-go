#!/bin/bash

# K6 Load Testing Suite for Social Media Service
# Multiple test scenarios with different load patterns

echo "🚀 K6 Load Testing Suite for Social Media Service"
echo "==============================================="

# Check if K6 is installed
if ! command -v k6 &> /dev/null; then
    echo "❌ K6 is not installed. Please install K6 first:"
    echo "   brew install k6"
    echo "   or visit: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Check if the API server is running
echo "🔍 Checking if API server is running..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ API server is not running on localhost:8080"
    echo "   Please start the server first:"
    echo "   go run cmd/server/main.go"
    exit 1
fi

echo "✅ API server is running"

# Function to run a test
run_test() {
    local test_name="$1"
    local test_file="$2"
    local description="$3"
    
    echo ""
    echo "📊 Running: $test_name"
    echo "Description: $description"
    echo "----------------------------------------"
    
    k6 run "$test_file"
    
    echo ""
    echo "✅ $test_name completed!"
}

# Menu for test selection
echo ""
echo "Select test to run:"
echo "1. Basic Load Test (100 RPM for 2 minutes)"
echo "2. Multi-Endpoint Test (Multiple APIs)"
echo "3. Stress Test (High load)"
echo "4. All Tests"
echo "5. Custom Test"
echo ""

read -p "Enter your choice (1-5): " choice

case $choice in
    1)
        run_test "Basic Load Test" "scripts/k6-load-test.js" "100 requests per minute for 2 minutes on /api/posts"
        ;;
    2)
        run_test "Multi-Endpoint Test" "scripts/k6-multi-endpoint-test.js" "Multiple endpoints with weighted distribution"
        ;;
    3)
        echo "📊 Running Stress Test..."
        echo "Description: High load test with 50 users for 5 minutes"
        echo "----------------------------------------"
        k6 run --vus 50 --duration 5m scripts/k6-load-test.js
        ;;
    4)
        echo "📊 Running All Tests..."
        run_test "Basic Load Test" "scripts/k6-load-test.js" "100 requests per minute for 2 minutes"
        sleep 10
        run_test "Multi-Endpoint Test" "scripts/k6-multi-endpoint-test.js" "Multiple endpoints with weighted distribution"
        ;;
    5)
        echo "Custom Test Options:"
        echo "1. Custom VUs and Duration"
        echo "2. Custom Ramp-up Pattern"
        echo ""
        read -p "Enter custom VUs: " vus
        read -p "Enter duration (e.g., 2m, 30s): " duration
        echo "Running custom test with $vus VUs for $duration..."
        k6 run --vus "$vus" --duration "$duration" scripts/k6-load-test.js
        ;;
    *)
        echo "❌ Invalid choice. Exiting..."
        exit 1
        ;;
esac

echo ""
echo "🎯 Load testing completed!"
echo ""
echo "📈 To view detailed metrics:"
echo "   - Grafana Dashboard: http://localhost:3000 (admin/admin123)"
echo "   - Graphite Web UI: http://localhost:8080"
echo ""
echo "📋 Results saved to: k6-results.json"
echo ""
echo "💡 Tips:"
echo "   - Check Grafana dashboard for real-time metrics"
echo "   - Monitor database performance during tests"
echo "   - Look for error rates and response time patterns"
