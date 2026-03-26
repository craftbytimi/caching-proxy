#!/bin/bash
# Start Redis for local development

set -e

echo "Starting Redis..."

if command -v docker &> /dev/null; then
    echo "Using Docker to run Redis..."
    docker run --name caching-proxy-redis \
        -p 6379:6379 \
        -d \
        redis:7-alpine
    echo "Redis started on port 6379"
    echo "To stop: docker stop caching-proxy-redis"
    echo "To remove: docker rm caching-proxy-redis"
else
    echo "Docker not found. Please install Docker or start Redis manually."
    exit 1
fi
