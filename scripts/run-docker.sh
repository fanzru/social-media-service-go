#!/bin/bash

# Script to run Docker container with .env file

set -e

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found!"
    echo "💡 Creating .env from sample-env..."
    cp sample-env .env
    echo "✅ Please edit .env file with your configuration before running again."
    exit 1
fi

# Build Docker image
echo "🔨 Building Docker image..."
docker build -t social-media-app:latest -f Dockerfile .

# Stop and remove existing container if exists
echo "🧹 Cleaning up existing container (if any)..."
docker stop social-media-app 2>/dev/null || true
docker rm social-media-app 2>/dev/null || true

# Run Docker container with .env file
echo "🚀 Starting Docker container..."
docker run -d \
    --name social-media-app \
    --env-file .env \
    -e SERVER_HOST=0.0.0.0 \
    -p 8080:8080 \
    --restart unless-stopped \
    social-media-app:latest

echo ""
echo "✅ Container started successfully!"
echo "🚀 App running at: http://localhost:8080"
echo ""
echo "📝 Useful commands:"
echo "   View logs: docker logs -f social-media-app"
echo "   Stop: docker stop social-media-app"
echo "   Remove: docker rm social-media-app"

