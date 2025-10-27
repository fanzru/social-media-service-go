#!/bin/bash

# Complete K6 Load Testing Suite
# Runs all tests in sequence with monitoring

echo "🚀 Complete K6 Load Testing Suite"
echo "================================="

# Check prerequisites
if ! command -v k6 &> /dev/null; then
    echo "❌ K6 is not installed. Please install K6 first:"
    echo "   brew install k6"
    exit 1
fi

if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ API server is not running on localhost:8080"
    echo "   Please start the server first:"
    echo "   go run cmd/server/main.go"
    exit 1
fi

if ! curl -s http://localhost:3000 > /dev/null; then
    echo "⚠️  Grafana is not running. Start monitoring:"
    echo "   docker-compose up -d graphite-statsd grafana"
    echo ""
fi

echo "✅ Prerequisites checked"
echo ""

# Function to run test with monitoring info
run_test() {
    local test_name="$1"
    local test_file="$2"
    local description="$3"
    
    echo "📊 Running: $test_name"
    echo "Description: $description"
    echo "----------------------------------------"
    echo "💡 Monitor real-time metrics at:"
    echo "   - Grafana: http://localhost:3000 (admin/admin123)"
    echo "   - Graphite: http://localhost:8080"
    echo ""
    
    k6 run "$test_file"
    
    echo ""
    echo "✅ $test_name completed!"
    echo ""
    
    # Wait between tests
    if [ "$test_name" != "Stress Test" ]; then
        echo "⏳ Waiting 30 seconds before next test..."
        sleep 30
        echo ""
    fi
}

# Run all tests
echo "🎯 Starting comprehensive load testing..."
echo ""

# Test 1: Basic Load Test
run_test "Basic Load Test" "scripts/k6/k6-load-test.js" "100 requests per minute for 2 minutes on /api/posts"

# Test 2: Multi-Endpoint Test
run_test "Multi-Endpoint Test" "scripts/k6/k6-multi-endpoint-test.js" "Multiple endpoints with weighted distribution"

# Test 3: Stress Test
run_test "Stress Test" "scripts/k6/k6-stress-test.js" "High load test with up to 100 users for 10 minutes"

echo "🎉 All tests completed!"
echo ""
echo "📊 Test Results Summary:"
echo "   - Basic Load Test: k6-results.json"
echo "   - Multi-Endpoint Test: k6-results.json"
echo "   - Stress Test: k6-stress-results.json"
echo ""
echo "📈 View detailed metrics:"
echo "   - Grafana Dashboard: http://localhost:3000 (admin/admin123)"
echo "   - Graphite Web UI: http://localhost:8080"
echo ""
echo "💡 Analysis Tips:"
echo "   - Check response time trends in Grafana"
echo "   - Monitor error rates during stress test"
echo "   - Look for performance bottlenecks"
echo "   - Compare database query performance"

