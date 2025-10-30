#!/bin/bash

# Script to run Docker Compose with .env file

set -e

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found!"
    echo "💡 Creating .env from sample-env..."
    cp sample-env .env
    echo "✅ Please edit .env file with your configuration before running again."
    exit 1
fi

echo "🔨 Building and starting services with Docker Compose..."
docker-compose up -d --build

echo ""
echo "✅ All services started successfully!"
echo ""
echo "🚀 Services running:"
echo "   📱 App: http://localhost:8080"
echo "   📊 Grafana: http://localhost:3000 (admin/admin123)"
echo "   📈 InfluxDB: http://localhost:8086 (admin/admin123)"
echo ""
echo "📝 Useful commands:"
echo "   View logs: docker-compose logs -f"
echo "   View app logs: docker-compose logs -f app"
echo "   Stop all: docker-compose down"
echo "   Stop and remove volumes: docker-compose down -v"

