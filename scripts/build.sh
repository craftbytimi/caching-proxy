#!/bin/bash
# Build the caching proxy binary

set -e

VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building caching-proxy..."
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Build time: ${BUILD_TIME}"

go build \
    -ldflags="-X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.Commit=${COMMIT}'" \
    -o bin/caching-proxy \
    ./cmd/proxy

echo ""
echo "Build complete: bin/caching-proxy"
